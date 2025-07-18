package cloudfront

// DistributionInfo はCloudFrontディストリビューションの情報を保持する構造体
type DistributionInfo struct {
	Id         string
	DomainName string
	Comment    string
	Enabled    bool
}

