package ec2

// Instance Ec2Instance EC2インスタンスの情報を格納する構造体
type Instance struct {
	InstanceId   string
	InstanceName string
	State        string
}
