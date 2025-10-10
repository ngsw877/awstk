package cfn

import (
	"awstk/internal/aws"
	"awstk/internal/cli"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
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
		// エラー時にスタックイベントを取得して整形表示
		fmt.Fprintf(os.Stderr, "\n📋 エラーの詳細:\n\n")

		// AWS SDKでスタックイベントを取得
		if displayErr := displayFailedEvents(ctx, opts.StackName); displayErr != nil {
			fmt.Fprintf(os.Stderr, "⚠️  イベント情報の取得に失敗しました: %v\n", displayErr)
		}

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

// displayFailedEvents はスタックの失敗イベントを読みやすく表示する
func displayFailedEvents(ctx aws.Context, stackName string) error {
	// AWS SDK設定をロード
	cfg, err := aws.LoadAwsConfig(ctx)
	if err != nil {
		return fmt.Errorf("AWS設定の読み込みに失敗: %w", err)
	}

	// CloudFormation クライアントを作成
	client := cloudformation.NewFromConfig(cfg)

	// スタックイベントを取得
	input := &cloudformation.DescribeStackEventsInput{
		StackName: awssdk.String(stackName),
	}

	result, err := client.DescribeStackEvents(context.Background(), input)
	if err != nil {
		return fmt.Errorf("スタックイベントの取得に失敗: %w", err)
	}

	// 失敗イベントのみをフィルタ（リソースIDごとに最新のものだけ）
	seenResources := make(map[string]bool)
	failedEvents := []types.StackEvent{}
	for _, event := range result.StackEvents {
		status := string(event.ResourceStatus)
		resourceId := awssdk.ToString(event.LogicalResourceId)

		// 既に表示したリソースはスキップ
		if seenResources[resourceId] {
			continue
		}

		if strings.HasSuffix(status, "_FAILED") {
			failedEvents = append(failedEvents, event)
			seenResources[resourceId] = true

			if len(failedEvents) >= 5 { // 最大5件まで
				break
			}
		}
	}

	if len(failedEvents) == 0 {
		fmt.Fprintf(os.Stderr, "⚠️  失敗イベントが見つかりませんでした\n")
		return nil
	}

	// 読みやすい形式で表示
	for i, event := range failedEvents {
		if i > 0 {
			fmt.Fprintf(os.Stderr, "\n")
		}
		fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Fprintf(os.Stderr, "📍 リソース: %s\n", awssdk.ToString(event.LogicalResourceId))
		fmt.Fprintf(os.Stderr, "⏰ 時刻: %s\n", event.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(os.Stderr, "❌ ステータス: %s\n", event.ResourceStatus)

		if event.ResourceStatusReason != nil {
			fmt.Fprintf(os.Stderr, "💬 理由:\n")
			// 長いメッセージを折り返して表示
			reason := awssdk.ToString(event.ResourceStatusReason)
			const maxWidth = 70
			for len(reason) > 0 {
				if len(reason) <= maxWidth {
					fmt.Fprintf(os.Stderr, "   %s\n", reason)
					break
				}
				// 適切な位置で折り返し
				breakPoint := maxWidth
				for breakPoint > 0 && reason[breakPoint] != ' ' {
					breakPoint--
				}
				if breakPoint == 0 {
					breakPoint = maxWidth
				}
				fmt.Fprintf(os.Stderr, "   %s\n", reason[:breakPoint])
				reason = strings.TrimSpace(reason[breakPoint:])
			}
		}
	}
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

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
