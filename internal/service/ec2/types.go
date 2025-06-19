package ec2

// Ec2Instance EC2インスタンスの情報を格納する構造体
type Ec2Instance struct {
	InstanceId   string
	InstanceName string
	State        string
}
