package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// awsRegion represents an AWS region
type awsRegion struct {
	regionName  string
	optInStatus string
}

// regionGroups represents grouped regions by availability (private)
type regionGroups struct {
	available []awsRegion
	disabled  []awsRegion
}

// GetFormattedRegionList retrieves and formats region list as string for display
func GetFormattedRegionList(ec2Client *ec2.Client, showAllRegions bool) (string, error) {
	regions, err := listRegions(ec2Client, showAllRegions)
	if err != nil {
		return "", err
	}

	if len(regions) == 0 {
		return "リージョンが見つかりませんでした", nil
	}

	groups := groupRegions(regions)
	var output strings.Builder

	if showAllRegions {
		// --all オプション時は有効/無効を分けて表示
		output.WriteString(fmt.Sprintf("AWSリージョン一覧: (全%d件)\n\n", len(regions)))

		if len(groups.available) > 0 {
			output.WriteString(fmt.Sprintf("✅ 有効なリージョン (%d件):\n", len(groups.available)))
			for i, region := range groups.available {
				output.WriteString(fmt.Sprintf("  %3d. %s (%s)\n", i+1, region.regionName, region.optInStatus))
			}
		}

		if len(groups.disabled) > 0 {
			output.WriteString(fmt.Sprintf("\n❌ 無効なリージョン (%d件):\n", len(groups.disabled)))
			for i, region := range groups.disabled {
				output.WriteString(fmt.Sprintf("  %3d. %s (%s)\n", i+1, region.regionName, region.optInStatus))
			}
		}
	} else {
		// デフォルトは有効なリージョンのみ表示
		output.WriteString(fmt.Sprintf("利用可能なリージョン一覧: (全%d件)\n", len(groups.available)))
		for i, region := range groups.available {
			output.WriteString(fmt.Sprintf("  %3d. %s\n", i+1, region.regionName))
		}
	}

	return output.String(), nil
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

// groupRegions groups regions by availability (private)
func groupRegions(regions []awsRegion) regionGroups {
	var groups regionGroups

	for _, region := range regions {
		switch region.optInStatus {
		case "opt-in-not-required", "opted-in":
			groups.available = append(groups.available, region)
		default:
			groups.disabled = append(groups.disabled, region)
		}
	}

	return groups
}
