package cfn

import (
	"awstk/internal/aws"
	"awstk/internal/cli"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// DeployOptions はデプロイコマンドのオプション
type DeployOptions struct {
	TemplatePath  string
	StackName     string
	Parameters    map[string]string
	ParameterFile string
	NoExecute     bool
}

// DeployStack は指定したテンプレートファイルからCloudFormationスタックをデプロイ
func DeployStack(ctx aws.Context, opts DeployOptions) error {
	// テンプレートファイルの存在確認
	if _, err := os.Stat(opts.TemplatePath); os.IsNotExist(err) {
		return fmt.Errorf("テンプレートファイルが見つかりません: %s", opts.TemplatePath)
	}

	// AWS CLIコマンドの引数を構築
	args := []string{
		"cloudformation", "deploy",
		"--template-file", opts.TemplatePath,
		"--stack-name", opts.StackName,
	}

	// パラメータを追加
	parameters, err := resolveParameters(opts.Parameters, opts.ParameterFile)
	if err != nil {
		return err
	}

	if len(parameters) > 0 {
		args = append(args, "--parameter-overrides")
		for key, value := range parameters {
			args = append(args, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Capabilitiesを追加（常にNAMED_IAMを付与）
	args = append(args, "--capabilities", "CAPABILITY_NAMED_IAM")

	// --no-execute-changeset オプション
	if opts.NoExecute {
		args = append(args, "--no-execute-changeset")
	}

	fmt.Printf("🚀 CloudFormationスタックをデプロイ中...\n")
	fmt.Printf("   スタック名: %s\n", opts.StackName)
	fmt.Printf("   テンプレート: %s\n", opts.TemplatePath)

	// AWS CLIコマンドを実行
	if err := cli.ExecuteAwsCommand(ctx, args); err != nil {
		return fmt.Errorf("デプロイに失敗しました: %w", err)
	}

	if opts.NoExecute {
		fmt.Printf("\n✅ Change Setの作成が完了しました\n")
		fmt.Printf("   AWS Management Consoleで内容を確認し、手動で実行してください\n")
	} else {
		fmt.Printf("\n✅ デプロイが完了しました\n")
	}

	return nil
}

// resolveParameters はパラメータ指定を解決する
// ParameterFileが.jsonで終わる場合はJSONファイルとして読み込む
func resolveParameters(params map[string]string, paramFile string) (map[string]string, error) {
	// ParameterFileが指定されている場合
	if paramFile != "" {
		// .json拡張子チェック
		if !strings.HasSuffix(strings.ToLower(paramFile), ".json") {
			return nil, fmt.Errorf("パラメータファイルは.json形式である必要があります: %s", paramFile)
		}

		// ファイル存在確認
		if _, err := os.Stat(paramFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("パラメータファイルが見つかりません: %s", paramFile)
		}

		// JSONファイルを読み込み
		data, err := os.ReadFile(paramFile)
		if err != nil {
			return nil, fmt.Errorf("パラメータファイルの読み込みに失敗しました: %w", err)
		}

		// JSONをパース
		var fileParams map[string]string
		if err := json.Unmarshal(data, &fileParams); err != nil {
			return nil, fmt.Errorf("パラメータファイルのJSON解析に失敗しました: %w", err)
		}

		return fileParams, nil
	}

	// map形式のパラメータをそのまま返す
	return params, nil
}
