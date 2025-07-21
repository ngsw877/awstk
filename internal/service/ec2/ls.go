package ec2

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"awstk/internal/service/cfn"
	"awstk/internal/service/common"
)

// ListEc2Instances 現在のリージョンのEC2インスタンス一覧を取得する
func ListEc2Instances(ec2Client *ec2.Client) ([]Instance, error) {
	result, err := ec2Client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("EC2インスタンス一覧の取得に失敗: %w", err)
	}

	var instances []Instance
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// 終了済みのインスタンスは除外
			if instance.State.Name == types.InstanceStateNameTerminated {
				continue
			}

			// インスタンス名を取得（Nameタグから）
			instanceName := "（名前なし）"
			for _, tag := range instance.Tags {
				if *tag.Key == "Name" && tag.Value != nil {
					instanceName = *tag.Value
					break
				}
			}

			instances = append(instances, Instance{
				InstanceId:   *instance.InstanceId,
				InstanceName: instanceName,
				State:        string(instance.State.Name),
			})
		}
	}

	return instances, nil
}

// SelectInstanceInteractively EC2インスタンス一覧を表示してユーザーに選択させる
func SelectInstanceInteractively(ec2Client *ec2.Client) (string, error) {
	fmt.Println("EC2インスタンス一覧を取得中...")

	instances, err := ListEc2Instances(ec2Client)
	if err != nil {
		return "", fmt.Errorf("❌ EC2インスタンス一覧の取得に失敗: %w", err)
	}

	if len(instances) == 0 {
		return "", fmt.Errorf("❌ 利用可能なEC2インスタンスが見つかりません")
	}

	// インスタンス一覧を表示
	fmt.Println("\n📋 利用可能なEC2インスタンス:")
	
	columns := []common.TableColumn{
		{Header: "番号"},
		{Header: "インスタンスID"},
		{Header: "インスタンス名"},
		{Header: "状態"},
	}
	
	data := make([][]string, len(instances))
	for i, instance := range instances {
		data[i] = []string{
			fmt.Sprintf("%d", i+1),
			instance.InstanceId,
			instance.InstanceName,
			instance.State,
		}
	}
	
	common.PrintTable("EC2インスタンス一覧", columns, data)

	// ユーザーに選択させる
	fmt.Print("\n操作するインスタンスの番号を入力してください: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("❌ 入力の読み取りに失敗: %w", err)
	}

	// 入力を数値に変換
	input = strings.TrimSpace(input)
	selectedNum, err := strconv.Atoi(input)
	if err != nil {
		return "", fmt.Errorf("❌ 無効な番号です: %s", input)
	}

	// 範囲チェック
	if selectedNum < 1 || selectedNum > len(instances) {
		return "", fmt.Errorf("❌ 番号は1から%dの間で入力してください", len(instances))
	}

	selectedInstance := instances[selectedNum-1]
	fmt.Printf("✅ 選択されたインスタンス: %s (%s)\n",
		selectedInstance.InstanceName, selectedInstance.InstanceId)

	return selectedInstance.InstanceId, nil
}

// ListEc2InstancesFromStack 指定されたCloudFormationスタックに属するEC2インスタンス一覧を取得する
func ListEc2InstancesFromStack(ec2Client *ec2.Client, cfnClient *cloudformation.Client, stackName string) ([]Instance, error) {
	ids, err := cfn.GetAllEc2FromStack(cfnClient, stackName)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []Instance{}, nil
	}

	all, err := ListEc2Instances(ec2Client)
	if err != nil {
		return nil, err
	}

	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	var instances []Instance
	for _, ins := range all {
		if _, ok := idSet[ins.InstanceId]; ok {
			instances = append(instances, ins)
		}
	}

	return instances, nil
}
