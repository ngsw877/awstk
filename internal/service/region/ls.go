package region

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// ListRegions はAWSリージョンの一覧を取得する関数 (公開)
func ListRegions(ec2Client *ec2.Client, showAllRegions bool) ([]AwsRegion, error) {
	regions, err := listRegions(ec2Client, showAllRegions)
	if err != nil {
		return nil, err
	}

	// 内部型から公開型に変換
	var result []AwsRegion
	for _, region := range regions {
		result = append(result, AwsRegion{
			RegionName:  region.regionName,
			OptInStatus: region.optInStatus,
		})
	}

	return result, nil
}

// GroupRegions はリージョンを有効/無効でグループ化する関数 (公開)
func GroupRegions(regions []AwsRegion) ([]AwsRegion, []AwsRegion) {
	var available, disabled []AwsRegion

	for _, region := range regions {
		switch region.OptInStatus {
		case "opt-in-not-required", "opted-in":
			available = append(available, region)
		default:
			disabled = append(disabled, region)
		}
	}

	return available, disabled
}

// listRegions retrieves all AWS regions (private)
func listRegions(ec2Client *ec2.Client, showAllRegions bool) ([]awsRegion, error) {
	input := &ec2.DescribeRegionsInput{
		AllRegions: &showAllRegions,
	}

	result, err := ec2Client.DescribeRegions(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("リージョン一覧の取得に失敗: %w", err)
	}

	var regions []awsRegion
	for _, region := range result.Regions {
		regions = append(regions, awsRegion{
			regionName:  *region.RegionName,
			optInStatus: *region.OptInStatus,
		})
	}

	return regions, nil
}
