package aws

// AwsContext は認証情報を保持
type AwsContext struct {
	Profile string
	Region  string
}
