package route53

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

// ListHostedZones ListHostedZonesはRoute53のホストゾーンを一覧表示します
func ListHostedZones(client *route53.Client) error {
	ctx := context.Background()
	var zones []HostedZoneInfo
	paginator := route53.NewListHostedZonesPaginator(client, &route53.ListHostedZonesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("ホストゾーン一覧の取得エラー: %w", err)
		}

		for _, zone := range output.HostedZones {
			info := HostedZoneInfo{
				Id:          extractZoneId(*zone.Id),
				Name:        *zone.Name,
				RecordCount: *zone.ResourceRecordSetCount,
				IsPrivate:   zone.Config != nil && zone.Config.PrivateZone,
			}

			if zone.Config != nil && zone.Config.Comment != nil {
				info.Comment = *zone.Config.Comment
			}

			if zone.CallerReference != nil {
				info.CallerRef = *zone.CallerReference
			}

			zones = append(zones, info)
		}
	}

	if len(zones) == 0 {
		fmt.Println("ホストゾーンが見つかりませんでした。")
		return nil
	}

	// Display zones
	fmt.Printf("🔍 %d個のホストゾーンが見つかりました:\n\n", len(zones))

	// Calculate column widths
	maxNameLen := 20
	maxIdLen := 14
	for _, zone := range zones {
		if len(zone.Name) > maxNameLen {
			maxNameLen = len(zone.Name)
		}
		if len(zone.Id) > maxIdLen {
			maxIdLen = len(zone.Id)
		}
	}

	// ヘッダーを表示
	fmt.Printf("%-*s  %-*s  %-10s  %-12s  %s\n",
		maxNameLen, "ドメイン名",
		maxIdLen, "ゾーンID",
		"レコード数",
		"タイプ",
		"コメント")
	fmt.Println(strings.Repeat("-", maxNameLen+maxIdLen+50))

	// Print zones
	for _, zone := range zones {
		zoneType := "パブリック"
		if zone.IsPrivate {
			zoneType = "プライベート"
		}

		comment := zone.Comment
		if comment == "" {
			comment = "-"
		}

		fmt.Printf("%-*s  %-*s  %-10d  %-12s  %s\n",
			maxNameLen, zone.Name,
			maxIdLen, zone.Id,
			zone.RecordCount,
			zoneType,
			comment)
	}

	return nil
}

// getHostedZoneIdByNameはドメイン名からホストゾーンIDを取得します
func getHostedZoneIdByName(client *route53.Client, domainName string) (string, error) {
	ctx := context.Background()
	// Ensure domain name ends with a dot
	if !strings.HasSuffix(domainName, ".") {
		domainName += "."
	}

	paginator := route53.NewListHostedZonesPaginator(client, &route53.ListHostedZonesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("ホストゾーン一覧の取得エラー: %w", err)
		}

		for _, zone := range output.HostedZones {
			if *zone.Name == domainName {
				return extractZoneId(*zone.Id), nil
			}
		}
	}

	return "", fmt.Errorf("ドメイン %s のホストゾーンが見つかりませんでした", domainName)
}

// extractZoneIdは完全なリソースIDからゾーンIDを抽出します
// 例: "/hostedzone/Z1234567890ABC" -> "Z1234567890ABC"
func extractZoneId(fullId string) string {
	parts := strings.Split(fullId, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullId
}

// getHostedZoneDetailsは特定のホストゾーンの詳細情報を取得します
func getHostedZoneDetails(client *route53.Client, zoneId string) (*types.HostedZone, error) {
	ctx := context.Background()
	output, err := client.GetHostedZone(ctx, &route53.GetHostedZoneInput{
		Id: &zoneId,
	})
	if err != nil {
		return nil, fmt.Errorf("ホストゾーン詳細の取得エラー: %w", err)
	}
	return output.HostedZone, nil
}
