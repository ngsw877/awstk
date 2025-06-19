package region

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
