package cmd

import (
	"awstk/internal/aws"
	regionSvc "awstk/internal/service/region"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/spf13/cobra"
)

var showAllRegions bool

const regionLsAlias = "regions"

// RegionCmd represents the region command
var RegionCmd = &cobra.Command{
	Use:     "region",
	Aliases: []string{regionLsAlias},
	Short:   "リージョン関連の操作",
	Long: `AWSリージョンに関する情報を取得します。

使用例:
  ` + AppName + ` region ls # サブコマンドでリージョン一覧を表示
  ` + AppName + ` regions # エイリアスで直接リージョン一覧を表示`,
}

var regionLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "利用可能なAWSリージョンを一覧表示",
	Long: `利用可能なAWSリージョンの一覧を表示します。

デフォルトでは有効なリージョン（opt-in-not-required と opted-in）のみを表示します。
--all フラグを使用すると、無効なリージョンも含めて全てのリージョンを表示します。

使用例:
  ` + AppName + ` region ls
  ` + AppName + ` region ls --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listRegions(showAllRegions)
	},
	SilenceUsage: true,
}

func listRegions(showAllRegions bool) error {
	ec2Client, err := aws.NewClient[*ec2.Client](awsCtx)
	if err != nil {
		return fmt.Errorf("AWS設定の読み込みエラー: %w", err)
	}

	output, err := regionSvc.GetFormattedRegionList(ec2Client, showAllRegions)
	if err != nil {
		return fmt.Errorf("❌ リージョン一覧取得でエラー: %w", err)
	}

	fmt.Print(output)
	return nil
}

func init() {
	RootCmd.AddCommand(RegionCmd)
	RegionCmd.AddCommand(regionLsCmd)

	// --all フラグをregionコマンドにPersistentFlagsとして登録（サブコマンドでも利用可能）
	RegionCmd.PersistentFlags().BoolVarP(&showAllRegions, "all", "a", false, "無効なリージョンも含めて全てのリージョンを表示")

	// エイリアスが直接実行された場合の処理
	RegionCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// エイリアスで呼ばれた場合、lsコマンドのロジックを実行
		if cmd.CalledAs() == regionLsAlias {
			return listRegions(showAllRegions)
		}
		// 'region' コマンドが直接呼ばれた場合はヘルプを表示
		return cmd.Help()
	}
}
