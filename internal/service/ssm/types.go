package ssm

// SsmSessionOptions はSSMセッション開始のパラメータを格納する構造体
type SsmSessionOptions struct {
	Region     string
	Profile    string
	InstanceId string
}
