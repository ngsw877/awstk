package elb

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// CleanupLoadBalancersByFilter はフィルターに一致するロードバランサーを削除する
func CleanupLoadBalancersByFilter(client *elasticloadbalancingv2.Client, filter string, withTargetGroups bool, lbType string) error {
	// フィルターに一致するロードバランサーを取得
	lbs, err := GetLoadBalancersByFilter(client, filter, lbType)
	if err != nil {
		return fmt.Errorf("ロードバランサー一覧取得エラー: %w", err)
	}

	if len(lbs) == 0 {
		typeMsg := "ロードバランサー"
		if lbType != "" {
			typeMsg = fmt.Sprintf("%sタイプのロードバランサー", strings.ToUpper(lbType))
		}
		fmt.Printf("フィルター '%s' に一致する%sが見つかりませんでした\n", filter, typeMsg)
		return nil
	}

	// 削除対象のロードバランサーと削除保護状態を表示
	fmt.Printf("\n🎯 削除対象のロードバランサー（%d件）:\n", len(lbs))
	fmt.Println(strings.Repeat("-", 70))

	protectedCount := 0
	for i, lb := range lbs {
		protected, err := IsDeletionProtected(client, *lb.LoadBalancerArn)
		if err != nil {
			return fmt.Errorf("削除保護状態の確認エラー: %w", err)
		}

		protectionStatus := "無効"
		if protected {
			protectionStatus = "🔒有効"
			protectedCount++
		}

		lbTypeStr := getLBTypeDisplay(lb.Type)
		fmt.Printf("%d. %s [%s] (削除保護: %s)\n", i+1, *lb.LoadBalancerName, lbTypeStr, protectionStatus)
	}

	if protectedCount > 0 {
		fmt.Printf("\n⚠️  %d件のロードバランサーで削除保護が有効です。削除前に自動的に解除されます。\n", protectedCount)
	}

	// ターゲットグループも削除する場合の確認
	if withTargetGroups {
		fmt.Println("\n📌 関連するターゲットグループも削除されます")
	}

	// 確認プロンプト
	fmt.Printf("\n本当に削除しますか？ (yes/no): ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "yes" && answer != "y" {
		fmt.Println("削除をキャンセルしました")
		return nil
	}

	// 削除実行
	fmt.Println("\n削除を開始します...")
	for _, lb := range lbs {
		lbTypeStr := getLBTypeDisplay(lb.Type)
		fmt.Printf("  %s [%s] を処理中...\n", *lb.LoadBalancerName, lbTypeStr)

		if err := deleteLoadBalancer(client, lb, withTargetGroups); err != nil {
			fmt.Printf("❌ %s の削除に失敗: %v\n", *lb.LoadBalancerName, err)
			continue
		}
		fmt.Printf("✅ %s を削除しました\n", *lb.LoadBalancerName)
	}

	fmt.Println("\n✨ ロードバランサーの削除が完了しました")
	return nil
}

// deleteLoadBalancer は単一のロードバランサーを削除する
func deleteLoadBalancer(client *elasticloadbalancingv2.Client, lb types.LoadBalancer, withTargetGroups bool) error {
	// 削除保護の確認と解除
	protected, err := IsDeletionProtected(client, *lb.LoadBalancerArn)
	if err != nil {
		return fmt.Errorf("削除保護状態の確認エラー: %w", err)
	}

	if protected {
		fmt.Printf("    🔓 削除保護を解除中...\n")
		if err := disableDeletionProtection(client, *lb.LoadBalancerArn); err != nil {
			return fmt.Errorf("削除保護の解除エラー: %w", err)
		}
		// 削除保護解除が反映されるまで少し待つ
		time.Sleep(2 * time.Second)
	}

	// ターゲットグループを先に削除（指定された場合）
	if withTargetGroups {
		if err := deleteRelatedTargetGroups(client, *lb.LoadBalancerArn); err != nil {
			fmt.Printf("    ⚠️  ターゲットグループ削除エラー: %v\n", err)
			// ターゲットグループの削除に失敗してもロードバランサー削除は続行
		}
	}

	// ロードバランサー削除
	deleteInput := &elasticloadbalancingv2.DeleteLoadBalancerInput{
		LoadBalancerArn: lb.LoadBalancerArn,
	}

	_, err = client.DeleteLoadBalancer(context.Background(), deleteInput)
	if err != nil {
		return err
	}

	return nil
}

// disableDeletionProtection は削除保護を無効化する
func disableDeletionProtection(client *elasticloadbalancingv2.Client, arn string) error {
	input := &elasticloadbalancingv2.ModifyLoadBalancerAttributesInput{
		LoadBalancerArn: &arn,
		Attributes: []types.LoadBalancerAttribute{
			{
				Key:   strPtr("deletion_protection.enabled"),
				Value: strPtr("false"),
			},
		},
	}

	_, err := client.ModifyLoadBalancerAttributes(context.Background(), input)
	return err
}

// deleteRelatedTargetGroups は関連するターゲットグループを削除する
func deleteRelatedTargetGroups(client *elasticloadbalancingv2.Client, lbArn string) error {
	// ロードバランサーに関連するターゲットグループを取得
	tgInput := &elasticloadbalancingv2.DescribeTargetGroupsInput{
		LoadBalancerArn: &lbArn,
	}

	tgResp, err := client.DescribeTargetGroups(context.Background(), tgInput)
	if err != nil {
		return err
	}

	if len(tgResp.TargetGroups) == 0 {
		return nil
	}

	fmt.Printf("    🗑️  関連するターゲットグループ（%d件）を削除中...\n", len(tgResp.TargetGroups))

	// 各ターゲットグループを削除
	for _, tg := range tgResp.TargetGroups {
		deleteInput := &elasticloadbalancingv2.DeleteTargetGroupInput{
			TargetGroupArn: tg.TargetGroupArn,
		}

		_, err := client.DeleteTargetGroup(context.Background(), deleteInput)
		if err != nil {
			// エラーが発生してもログ出力して続行
			fmt.Printf("      ⚠️  %s の削除に失敗: %v\n", *tg.TargetGroupName, err)
		} else {
			fmt.Printf("      ✓ %s を削除しました\n", *tg.TargetGroupName)
		}
	}

	return nil
}

// strPtr は文字列へのポインタを返す（ヘルパー関数）
func strPtr(s string) *string {
	return &s
}
