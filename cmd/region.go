package cmd

import (
	"fmt"

	"awstk/internal"

	"github.com/spf13/cobra"
)

var showAllRegions bool

const regionLsAlias = "regions"

// regionCmd represents the region command
var regionCmd = &cobra.Command{
	Use:     "region",
	Aliases: []string{regionLsAlias},
	Short:   "リージョン関連の操作",
	Long: `AWSリージョンに関する情報を取得します。

使用例:
  ` + AppName + ` region ls      # サブコマンドでリージョン一覧を表示
  ` + AppName + ` regions        # エイリアスで直接リージョン一覧を表示`,
}

// regionLsCmd represents the region ls command
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
		awsCtx := getAwsContext()
		return listRegions(awsCtx, showAllRegions)
	},
}

func listRegions(awsCtx internal.AwsContext, showAllRegions bool) error {
	regions, err := internal.ListRegions(awsCtx, showAllRegions)
	if err != nil {
		return err
	}

	if len(regions) == 0 {
		fmt.Println("リージョンが見つかりませんでした。")
		return nil
	}

	// リージョンを有効/無効で分類
	groups := internal.GroupRegions(regions)

	// 有効なリージョンの表示
	fmt.Println("\n--- 有効なリージョン ---")
	if len(groups.Available) == 0 {
		fmt.Println("  (なし)")
	} else {
		for _, region := range groups.Available {
			fmt.Printf("  %-20s (%s)\n", region.RegionName, region.OptInStatus)
		}
	}

	// --all フラグが指定されている場合のみ無効なリージョンを表示
	if showAllRegions {
		fmt.Println("\n--- 無効なリージョン ---")
		if len(groups.Disabled) == 0 {
			fmt.Println("  (なし)")
		} else {
			for _, region := range groups.Disabled {
				fmt.Printf("  %-20s (%s)\n", region.RegionName, region.OptInStatus)
			}
		}
	}

	return nil
}

func init() {
	RootCmd.AddCommand(regionCmd)
	regionCmd.AddCommand(regionLsCmd)

	// --all フラグをregionコマンドにPersistentFlagsとして登録（サブコマンドでも利用可能）
	regionCmd.PersistentFlags().BoolVar(&showAllRegions, "all", false, "無効なリージョンも含めて全てのリージョンを表示")

	// エイリアスが直接実行された場合の処理
	regionCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// エイリアスで呼ばれた場合、lsコマンドのロジックを実行
		if cmd.CalledAs() == regionLsAlias {
			awsCtx := getAwsContext()
			return listRegions(awsCtx, showAllRegions)
		}
		// 'region' コマンドが直接呼ばれた場合はヘルプを表示
		return cmd.Help()
	}
}
