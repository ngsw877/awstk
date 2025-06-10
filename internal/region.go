package internal

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// Region はAWSリージョンの情報を格納する構造体
type Region struct {
	RegionName  string
	OptInStatus string
}

// RegionGroups はリージョンを有効/無効で分類したグループ
type RegionGroups struct {
	Available []Region // 有効なリージョン (opt-in-not-required と opted-in)
	Disabled  []Region // 無効なリージョン (not-opted-in)
}

// ListRegions は利用可能なAWSリージョン一覧を取得する
func ListRegions(awsCtx AwsContext, showAllRegions bool) ([]Region, error) {
	cfg, err := LoadAwsConfig(awsCtx)
	if err != nil {
		return nil, fmt.Errorf("AWS設定のロードに失敗: %w", err)
	}

	client := ec2.NewFromConfig(cfg)

	// DescribeRegionsの入力パラメータを設定
	input := &ec2.DescribeRegionsInput{}
	if showAllRegions {
		input.AllRegions = &showAllRegions
	}

	result, err := client.DescribeRegions(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("リージョン一覧の取得に失敗: %w", err)
	}

	var regions []Region
	for _, region := range result.Regions {
		regions = append(regions, Region{
			RegionName:  *region.RegionName,
			OptInStatus: string(*region.OptInStatus),
		})
	}

	return regions, nil
}

// GroupRegions はリージョンを有効/無効で分類する
func GroupRegions(regions []Region) RegionGroups {
	var groups RegionGroups

	for _, region := range regions {
		if region.OptInStatus == "opt-in-not-required" || region.OptInStatus == "opted-in" {
			groups.Available = append(groups.Available, region)
		} else {
			groups.Disabled = append(groups.Disabled, region)
		}
	}

	return groups
}
