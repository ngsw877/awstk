package elb

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// ListLoadBalancers はロードバランサー一覧を表示する
func ListLoadBalancers(client *elasticloadbalancingv2.Client, opts ListOptions) error {
	// ロードバランサー一覧を取得
	lbs, err := describeLoadBalancers(client, opts.LoadBalancerType)
	if err != nil {
		return fmt.Errorf("ロードバランサー一覧取得エラー: %w", err)
	}

	if len(lbs) == 0 {
		typeMsg := "ロードバランサー"
		if opts.LoadBalancerType != "" {
			typeMsg = fmt.Sprintf("%sタイプのロードバランサー", strings.ToUpper(opts.LoadBalancerType))
		}
		fmt.Printf("%sが見つかりませんでした\n", typeMsg)
		return nil
	}

	// 削除保護情報を取得
	lbInfos := []LoadBalancerInfo{}
	for _, lb := range lbs {
		info, err := getLoadBalancerInfo(client, lb)
		if err != nil {
			return fmt.Errorf("ロードバランサー情報取得エラー: %w", err)
		}

		// フィルタリング
		if opts.ProtectedOnly && !info.DeletionProtection {
			continue
		}

		lbInfos = append(lbInfos, info)
	}

	// 表示
	displayLoadBalancers(lbInfos, opts.ShowDetails)
	return nil
}

// describeLoadBalancers はロードバランサー一覧を取得する
func describeLoadBalancers(client *elasticloadbalancingv2.Client, lbTypeFilter string) ([]types.LoadBalancer, error) {
	var allLBs []types.LoadBalancer
	var nextMarker *string

	for {
		input := &elasticloadbalancingv2.DescribeLoadBalancersInput{
			Marker: nextMarker,
		}

		resp, err := client.DescribeLoadBalancers(context.Background(), input)
		if err != nil {
			return nil, err
		}

		// タイプでフィルタ
		for _, lb := range resp.LoadBalancers {
			if shouldIncludeLoadBalancer(lb, lbTypeFilter) {
				allLBs = append(allLBs, lb)
			}
		}

		if resp.NextMarker == nil {
			break
		}
		nextMarker = resp.NextMarker
	}

	return allLBs, nil
}

// shouldIncludeLoadBalancer はロードバランサーを含めるかどうかを判定
func shouldIncludeLoadBalancer(lb types.LoadBalancer, typeFilter string) bool {
	if typeFilter == "" {
		// フィルタが空の場合は全て含める
		return true
	}

	lbType := string(lb.Type)
	switch strings.ToLower(typeFilter) {
	case "alb":
		return lbType == string(types.LoadBalancerTypeEnumApplication)
	case "nlb":
		return lbType == string(types.LoadBalancerTypeEnumNetwork)
	case "gwlb":
		return lbType == string(types.LoadBalancerTypeEnumGateway)
	default:
		return false
	}
}

// getLoadBalancerInfo はロードバランサーの詳細情報を取得する
func getLoadBalancerInfo(client *elasticloadbalancingv2.Client, lb types.LoadBalancer) (LoadBalancerInfo, error) {
	info := LoadBalancerInfo{
		Name:    *lb.LoadBalancerName,
		ARN:     *lb.LoadBalancerArn,
		DNSName: *lb.DNSName,
		State:   string(lb.State.Code),
		Type:    getLBTypeDisplay(lb.Type),
		Scheme:  string(lb.Scheme),
	}

	// VPC ID
	if lb.VpcId != nil {
		info.VPCId = *lb.VpcId
	}

	// 作成時刻
	if lb.CreatedTime != nil {
		info.CreatedTime = lb.CreatedTime.Format("2006-01-02 15:04:05")
	}

	// アベイラビリティゾーン
	for _, az := range lb.AvailabilityZones {
		if az.ZoneName != nil {
			info.AvailabilityZones = append(info.AvailabilityZones, *az.ZoneName)
		}
	}

	// 属性を取得（削除保護など）
	attrInput := &elasticloadbalancingv2.DescribeLoadBalancerAttributesInput{
		LoadBalancerArn: lb.LoadBalancerArn,
	}
	attrResp, err := client.DescribeLoadBalancerAttributes(context.Background(), attrInput)
	if err != nil {
		return info, fmt.Errorf("属性取得エラー: %w", err)
	}

	for _, attr := range attrResp.Attributes {
		if attr.Key != nil && *attr.Key == "deletion_protection.enabled" && attr.Value != nil {
			info.DeletionProtection = *attr.Value == "true"
		}
	}

	// リスナー数を取得
	listenersInput := &elasticloadbalancingv2.DescribeListenersInput{
		LoadBalancerArn: lb.LoadBalancerArn,
	}
	listenersResp, err := client.DescribeListeners(context.Background(), listenersInput)
	if err == nil {
		info.ListenerCount = len(listenersResp.Listeners)
	}

	// ターゲットグループ数を取得
	tgInput := &elasticloadbalancingv2.DescribeTargetGroupsInput{
		LoadBalancerArn: lb.LoadBalancerArn,
	}
	tgResp, err := client.DescribeTargetGroups(context.Background(), tgInput)
	if err == nil {
		info.TargetGroupCount = len(tgResp.TargetGroups)
	}

	return info, nil
}

// getLBTypeDisplay はロードバランサータイプの表示名を取得
func getLBTypeDisplay(lbType types.LoadBalancerTypeEnum) string {
	switch lbType {
	case types.LoadBalancerTypeEnumApplication:
		return "ALB"
	case types.LoadBalancerTypeEnumNetwork:
		return "NLB"
	case types.LoadBalancerTypeEnumGateway:
		return "GWLB"
	default:
		return string(lbType)
	}
}

// displayLoadBalancers はロードバランサー一覧を表示する
func displayLoadBalancers(lbs []LoadBalancerInfo, showDetails bool) {
	fmt.Printf("\n🔍 ロードバランサー一覧（%d件）\n", len(lbs))
	fmt.Println(strings.Repeat("=", 80))

	if showDetails {
		// 詳細表示
		for i, lb := range lbs {
			fmt.Printf("\n[%d] %s (%s)\n", i+1, lb.Name, lb.Type)
			fmt.Printf("    状態: %s\n", lb.State)
			fmt.Printf("    スキーマ: %s\n", lb.Scheme)
			fmt.Printf("    DNS名: %s\n", lb.DNSName)
			fmt.Printf("    削除保護: %s\n", formatBool(lb.DeletionProtection))
			fmt.Printf("    リスナー数: %d\n", lb.ListenerCount)
			fmt.Printf("    ターゲットグループ数: %d\n", lb.TargetGroupCount)
			fmt.Printf("    VPC ID: %s\n", lb.VPCId)
			fmt.Printf("    AZ: %s\n", strings.Join(lb.AvailabilityZones, ", "))
			fmt.Printf("    作成日時: %s\n", lb.CreatedTime)
		}
	} else {
		// 簡易表示（テーブル形式）
		fmt.Printf("%-35s %-5s %-10s %-8s %-10s %-5s %-5s\n",
			"名前", "タイプ", "状態", "スキーマ", "削除保護", "TG数", "ﾘｽﾅｰ")
		fmt.Println(strings.Repeat("-", 80))

		for _, lb := range lbs {
			protection := "無効"
			if lb.DeletionProtection {
				protection = "🔒有効"
			}
			fmt.Printf("%-35s %-5s %-10s %-8s %-10s %-5d %-5d\n",
				truncate(lb.Name, 35),
				lb.Type,
				lb.State,
				lb.Scheme,
				protection,
				lb.TargetGroupCount,
				lb.ListenerCount,
			)
		}
	}
	fmt.Println()
}

// formatBool はブール値を日本語で表示する
func formatBool(b bool) string {
	if b {
		return "有効"
	}
	return "無効"
}

// truncate は文字列を指定長で切り詰める
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// GetLoadBalancersByFilter はフィルターに一致するロードバランサーを取得する
func GetLoadBalancersByFilter(client *elasticloadbalancingv2.Client, filter string, lbType string) ([]types.LoadBalancer, error) {
	allLBs, err := describeLoadBalancers(client, lbType)
	if err != nil {
		return nil, err
	}

	var filtered []types.LoadBalancer
	for _, lb := range allLBs {
		if lb.LoadBalancerName != nil && strings.Contains(*lb.LoadBalancerName, filter) {
			filtered = append(filtered, lb)
		}
	}

	return filtered, nil
}

// IsDeletionProtected は削除保護が有効かチェックする
func IsDeletionProtected(client *elasticloadbalancingv2.Client, arn string) (bool, error) {
	input := &elasticloadbalancingv2.DescribeLoadBalancerAttributesInput{
		LoadBalancerArn: &arn,
	}

	resp, err := client.DescribeLoadBalancerAttributes(context.Background(), input)
	if err != nil {
		return false, err
	}

	for _, attr := range resp.Attributes {
		if attr.Key != nil && *attr.Key == "deletion_protection.enabled" && attr.Value != nil {
			protected, _ := strconv.ParseBool(*attr.Value)
			return protected, nil
		}
	}

	return false, nil
}
