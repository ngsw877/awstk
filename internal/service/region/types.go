package region

// awsRegion represents an AWS region (private)
type awsRegion struct {
	regionName  string
	optInStatus string
}

// AwsRegion represents an AWS region (public)
type AwsRegion struct {
	RegionName  string
	OptInStatus string
}
