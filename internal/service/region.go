package service

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// AWSRegion represents an AWS region
type AWSRegion struct {
	RegionName  string
	OptInStatus string
}

// RegionGroups represents grouped regions by availability
type RegionGroups struct {
	Available []AWSRegion
	Disabled  []AWSRegion
}

// ListRegions retrieves all AWS regions
func ListRegions(ec2Client *ec2.Client, showAllRegions bool) ([]AWSRegion, error) {
	input := &ec2.DescribeRegionsInput{
		AllRegions: &showAllRegions,
	}

	result, err := ec2Client.DescribeRegions(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("リージョン一覧の取得に失敗: %w", err)
	}

	var regions []AWSRegion
	for _, region := range result.Regions {
		regions = append(regions, AWSRegion{
			RegionName:  *region.RegionName,
			OptInStatus: *region.OptInStatus,
		})
	}

	return regions, nil
}

// GroupRegions groups regions by availability
func GroupRegions(regions []AWSRegion) RegionGroups {
	var groups RegionGroups

	for _, region := range regions {
		switch region.OptInStatus {
		case "opt-in-not-required", "opted-in":
			groups.Available = append(groups.Available, region)
		default:
			groups.Disabled = append(groups.Disabled, region)
		}
	}

	return groups
}
