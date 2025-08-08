package cmd

import (
	"awstk/internal/service/env"
	"fmt"

	"github.com/spf13/cobra"
)

// EnvCmd represents the env command
var EnvCmd = &cobra.Command{
	Use:   "env",
	Short: "AWS環境変数の管理コマンド",
	Long: `AWS関連の環境変数を管理するためのコマンド群です。
スタック名(AWS_STACK_NAME)やプロファイル(AWS_PROFILE)などの環境変数を設定・表示・削除できます。`,
}

var (
	envStackName string
	envProfile   string
)

var envSetCmd = &cobra.Command{
	Use:   "set",
	Short: "環境変数の設定方法を表示",
	Long: `指定した環境変数を設定するためのexportコマンドを表示します。

例:
  ` + AppName + ` env set -S my-stack
  ` + AppName + ` env set -P my-profile
  ` + AppName + ` env set -S my-stack -P my-profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var commands []string

		if envStackName != "" {
			exportCmd, err := env.GetExportCommand("stack", envStackName)
			if err != nil {
				return err
			}
			commands = append(commands, exportCmd)
		}

		if envProfile != "" {
			exportCmd, err := env.GetExportCommand("profile", envProfile)
			if err != nil {
				return err
			}
			commands = append(commands, exportCmd)
		}

		if len(commands) == 0 {
			return fmt.Errorf("❌ エラー: -S (スタック名) または -P (プロファイル) を指定してください")
		}

		fmt.Println("✅ 以下のコマンドを実行して環境変数を設定してください：")
		for _, cmd := range commands {
			fmt.Println(cmd)
		}
		return nil
	},
}

var envShowCmd = &cobra.Command{
	Use:   "show",
	Short: "環境変数の現在値を表示",
	Long: `現在設定されているAWS関連の環境変数を表示します。

例:
  ` + AppName + ` env show`,
	RunE: func(cmd *cobra.Command, args []string) error {
		env.ShowAllVariables()
		return nil
	},
}

var envUnsetCmd = &cobra.Command{
	Use:   "unset",
	Short: "環境変数の削除方法を表示",
	Long: `指定した環境変数を削除するためのunsetコマンドを表示します。

例:
  ` + AppName + ` env unset -S
  ` + AppName + ` env unset -P
  ` + AppName + ` env unset -S -P`,
	RunE: func(cmd *cobra.Command, args []string) error {
		unsetStack, _ := cmd.Flags().GetBool("stack")
		unsetProfile, _ := cmd.Flags().GetBool("profile")

		commands := []string{}

		if unsetStack {
			unsetCmd, err := env.GetUnsetCommand("stack")
			if err != nil {
				return err
			}
			commands = append(commands, unsetCmd)
		}

		if unsetProfile {
			unsetCmd, err := env.GetUnsetCommand("profile")
			if err != nil {
				return err
			}
			commands = append(commands, unsetCmd)
		}

		if len(commands) == 0 {
			return fmt.Errorf("❌ エラー: -S (スタック名) または -P (プロファイル) を指定してください")
		}

		fmt.Println("✅ 以下のコマンドを実行して環境変数を削除してください：")
		for _, cmd := range commands {
			fmt.Println(cmd)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(EnvCmd)
	EnvCmd.AddCommand(envSetCmd)
	EnvCmd.AddCommand(envShowCmd)
	EnvCmd.AddCommand(envUnsetCmd)

	// env set のフラグ
	envSetCmd.Flags().StringVarP(&envStackName, "stack", "S", "", "設定するスタック名")
	envSetCmd.Flags().StringVarP(&envProfile, "profile", "P", "", "設定するプロファイル名")
	// どちらか1つ必須
	envSetCmd.MarkFlagsOneRequired("stack", "profile")

	// env show のフラグは不要（常に全て表示）

	// env unset のフラグ
	envUnsetCmd.Flags().BoolP("stack", "S", false, "スタック名を削除")
	envUnsetCmd.Flags().BoolP("profile", "P", false, "プロファイル名を削除")
	// どちらか1つ必須
	envUnsetCmd.MarkFlagsOneRequired("stack", "profile")
}
