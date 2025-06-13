package cmd

import (
	"awstk/internal/aws"
	"awstk/internal/service"
	"fmt"

	"github.com/spf13/cobra"
)

// RegionCmd represents the region command
var RegionCmd = &cobra.Command{
	Use:   "region",
	Short: "AWSリージョン操作コマンド",
	Long:  `AWSリージョンに関する情報を操作するためのコマンド群です。`,
}

var regionLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "利用可能なAWSリージョン一覧を表示するコマンド",
	Long:  `利用可能なAWSリージョンの一覧を表示します。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		awsClients, err := aws.NewAwsClients(aws.AwsContext{Region: region, Profile: profile})
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		ec2Client := awsClients.Ec2()

		// 既存の関数シグネチャに合わせてboolパラメータを追加
		regions, err := service.ListRegions(ec2Client, false)
		if err != nil {
			return fmt.Errorf("❌ リージョン一覧取得でエラー: %w", err)
		}

		if len(regions) == 0 {
			fmt.Println("利用可能なリージョンが見つかりませんでした")
			return nil
		}

		// リージョンを有効/無効で分類
		groups := service.GroupRegions(regions)

		// 有効なリージョンの表示
		fmt.Printf("利用可能なAWSリージョン一覧: (全%d件)\n", len(groups.Available))
		for i, region := range groups.Available {
			fmt.Printf("  %3d. %-20s (%s)\n", i+1, region.RegionName, region.OptInStatus)
		}

		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(RegionCmd)
	RegionCmd.AddCommand(regionLsCmd)
}
