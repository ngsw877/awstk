package cmd

import (
	"awsfunc/internal"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	auroraClusterID string
	auroraProfile   string
)

var auroraCmd = &cobra.Command{
	Use:   "aurora",
	Short: "Aurora関連の操作を行うコマンド群",
	Long:  "AWS Aurora DBクラスターの操作を行うCLIコマンド群です。",
}

var auroraStartClusterCmd = &cobra.Command{
	Use:   "start-cluster",
	Short: "Aurora DBクラスターを起動する",
	Long: `指定したAurora DBクラスターを起動します。

例:
  awsfunc aurora start-cluster -d <aurora-cluster-identifier> [-P <aws-profile>]
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if auroraClusterID == "" {
			return fmt.Errorf("❌ Aurora DBクラスター識別子は必須です")
		}
		fmt.Printf("Aurora DBクラスター (%s) を起動します...\n", auroraClusterID)

		err := internal.StartAuroraCluster(auroraClusterID, auroraProfile)

		if err != nil {
			fmt.Printf("❌ Aurora DBクラスターの起動に失敗しました。")
		}

		fmt.Println("✅ Aurora DBクラスターの起動を開始しました。起動完了まで数十分かかります。")
		return nil
	},
	SilenceUsage: true,
}

func init() {
	RootCmd.AddCommand(auroraCmd)
	auroraCmd.AddCommand(auroraStartClusterCmd)
	auroraStartClusterCmd.Flags().StringVarP(&auroraClusterID, "db-cluster-identifier", "d", "", "Aurora DBクラスター識別子（必須）")
}
