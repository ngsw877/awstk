package cloudfront

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
)

// SelectDistribution は複数のディストリビューションから一つを選択します
func SelectDistribution(client *cloudfront.Client, distributionIds []string) (string, error) {
	fmt.Println("\n複数のCloudFrontディストリビューションが見つかりました。選択してください:")

	// 各ディストリビューションの詳細情報を取得して表示
	for i, id := range distributionIds {
		input := &cloudfront.GetDistributionInput{
			Id: &id,
		}

		result, err := client.GetDistribution(context.Background(), input)
		if err != nil {
			// エラーが発生してもIDは表示
			fmt.Printf("  %d. %s (詳細情報の取得に失敗)\n", i+1, id)
			continue
		}

		dist := result.Distribution
		comment := ""
		if dist.DistributionConfig.Comment != nil {
			comment = *dist.DistributionConfig.Comment
		}
		domainName := ""
		if dist.DomainName != nil {
			domainName = *dist.DomainName
		}

		fmt.Printf("  %d. %s - %s (%s)\n", i+1, id, domainName, comment)
	}

	// ユーザーの選択を待つ
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n番号を入力してください (1-" + fmt.Sprintf("%d", len(distributionIds)) + "): ")
	
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("入力エラー: %w", err)
	}

	// 選択番号を解析
	choice, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || choice < 1 || choice > len(distributionIds) {
		return "", fmt.Errorf("無効な選択です")
	}

	selectedId := distributionIds[choice-1]
	fmt.Printf("\n✅ ディストリビューション '%s' を選択しました\n", selectedId)
	
	return selectedId, nil
}