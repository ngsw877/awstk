package cmd

import (
	"awstk/internal/aws"
	"awstk/internal/service"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
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
		ec2Client, err := aws.NewClient[*ec2.Client](aws.Context{Region: region, Profile: profile})
		if err != nil {
			return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
		}

		regions, err := service.ListRegions(ec2Client, false)
		if err != nil {
			return fmt.Errorf("❌ リージョン一覧取得でエラー: %w", err)
		}

		if len(regions) == 0 {
			fmt.Println("リージョンが見つかりませんでした")
			return nil
		}

		fmt.Printf("利用可能なリージョン一覧: (全%d件)\n", len(regions))
		for i, regionName := range regions {
			fmt.Printf("  %3d. %s\n", i+1, regionName)
		}

		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(RegionCmd)
	RegionCmd.AddCommand(regionLsCmd)
}
