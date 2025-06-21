package aws

// Context AwsContext は認証情報を保持
type Context struct {
	Profile string
	Region  string
}
