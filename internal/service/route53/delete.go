package route53

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

// DeleteHostedZone DeleteHostedZoneはRoute53のホストゾーンとすべてのレコードを削除します
func DeleteHostedZone(client *route53.Client, identifier string, opts DeleteOptions) error {
	ctx := context.Background()
	var zoneId string
	var zoneName string
	var err error

	// ゾーンIDを取得
	if opts.UseId {
		zoneId = identifier
		// ゾーン詳細を取得して名前を取得
		zone, err := getHostedZoneDetails(client, zoneId)
		if err != nil {
			return fmt.Errorf("ホストゾーン詳細の取得エラー: %w", err)
		}
		zoneName = *zone.Name
	} else {
		// ドメイン名として扱う
		zoneName = identifier
		if !strings.HasSuffix(zoneName, ".") {
			zoneName += "."
		}
		zoneId, err = getHostedZoneIdByName(client, zoneName)
		if err != nil {
			return err
		}
	}

	fmt.Printf("🔍 ホストゾーンが見つかりました: %s (ID: %s)\n", zoneName, zoneId)

	// すべてのレコードを一覧取得
	records, err := listAllRecords(client, zoneId)
	if err != nil {
		return fmt.Errorf("レコード一覧の取得エラー: %w", err)
	}

	// NSとSOAレコードを除外
	var recordsToDelete []RecordSetInfo
	for _, record := range records {
		if record.Type != types.RRTypeNs && record.Type != types.RRTypeSoa {
			recordsToDelete = append(recordsToDelete, record)
		}
	}

	if opts.DryRun {
		fmt.Println("\n[ドライラン] 以下を削除します:")
		fmt.Printf("- ホストゾーン: %s (ID: %s)\n", zoneName, zoneId)
		fmt.Printf("- %d個のリソースレコードセット:\n", len(recordsToDelete))
		for _, record := range recordsToDelete {
			fmt.Printf("  - %s (%s)\n", record.Name, record.Type)
		}
		return nil
	}

	// 削除確認
	if !opts.Force {
		fmt.Printf("\n⚠️  以下を完全に削除します:\n")
		fmt.Printf("- ホストゾーン: %s (ID: %s)\n", zoneName, zoneId)
		fmt.Printf("- %d個のリソースレコードセット\n", len(recordsToDelete))

		if !confirmPrompt("\n本当に続行しますか？") {
			fmt.Println("削除がキャンセルされました。")
			return nil
		}
	}

	// レコード削除
	if len(recordsToDelete) > 0 {
		fmt.Printf("\n🗑️  %d個のレコードを削除中...\n", len(recordsToDelete))
		deletedCount, failedCount := deleteRecords(client, zoneId, recordsToDelete)

		if failedCount > 0 {
			fmt.Printf("⚠️  %d個のレコードの削除に失敗しました\n", failedCount)
		}
		if deletedCount > 0 {
			fmt.Printf("✅ %d個のレコードを削除しました\n", deletedCount)
		}
	}

	// ホストゾーン削除
	fmt.Printf("\n🗑️  ホストゾーンを削除中...\n")
	_, err = client.DeleteHostedZone(ctx, &route53.DeleteHostedZoneInput{
		Id: &zoneId,
	})
	if err != nil {
		return fmt.Errorf("ホストゾーンの削除エラー: %w", err)
	}

	fmt.Printf("✅ ホストゾーンを正常に削除しました: %s (ID: %s)\n", zoneName, zoneId)
	return nil
}

// listAllRecordsはホストゾーン内のすべてのリソースレコードセットを一覧取得します
func listAllRecords(client *route53.Client, zoneId string) ([]RecordSetInfo, error) {
	ctx := context.Background()
	var records []RecordSetInfo
	paginator := route53.NewListResourceRecordSetsPaginator(client, &route53.ListResourceRecordSetsInput{
		HostedZoneId: &zoneId,
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("リソースレコードセット一覧の取得エラー: %w", err)
		}

		for _, recordSet := range output.ResourceRecordSets {
			info := RecordSetInfo{
				Name:          *recordSet.Name,
				Type:          recordSet.Type,
				TTL:           recordSet.TTL,
				AliasTarget:   recordSet.AliasTarget,
				SetIdentifier: recordSet.SetIdentifier,
				Weight:        recordSet.Weight,
				Region:        recordSet.Region,
				Failover:      recordSet.Failover,
				HealthCheckId: recordSet.HealthCheckId,
			}

			// レコード値を抽出
			for _, record := range recordSet.ResourceRecords {
				if record.Value != nil {
					info.Records = append(info.Records, *record.Value)
				}
			}

			records = append(records, info)
		}
	}

	return records, nil
}

// deleteRecordsは複数のリソースレコードセットを削除します
func deleteRecords(client *route53.Client, zoneId string, records []RecordSetInfo) (deleted, failed int) {
	ctx := context.Background()
	// レコードをバッチ処理（Route53は1リクエストあたり最大1000変更までサポート）
	batchSize := 100

	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		changes := make([]types.Change, 0, len(batch))

		for _, record := range batch {
			// ResourceRecordSetを再構築
			recordSet := types.ResourceRecordSet{
				Name:          &record.Name,
				Type:          record.Type,
				TTL:           record.TTL,
				AliasTarget:   record.AliasTarget,
				SetIdentifier: record.SetIdentifier,
				Weight:        record.Weight,
				Region:        record.Region,
				Failover:      record.Failover,
				HealthCheckId: record.HealthCheckId,
			}

			// リソースレコードを追加
			for _, value := range record.Records {
				v := value // ポインタ用にコピーを作成
				recordSet.ResourceRecords = append(recordSet.ResourceRecords, types.ResourceRecord{
					Value: &v,
				})
			}

			changes = append(changes, types.Change{
				Action:            types.ChangeActionDelete,
				ResourceRecordSet: &recordSet,
			})
		}

		// バッチ削除を実行
		_, err := client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: &zoneId,
			ChangeBatch: &types.ChangeBatch{
				Changes: changes,
			},
		})

		if err != nil {
			failed += len(batch)
			fmt.Printf("  ❌ %d個のレコードのバッチ削除に失敗: %v\n", len(batch), err)
		} else {
			deleted += len(batch)
			fmt.Printf("  ✓ %d個のレコードをバッチ削除\n", len(batch))
		}
	}

	return deleted, failed
}

// confirmPromptはユーザーに確認を求めます
func confirmPrompt(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", message)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
