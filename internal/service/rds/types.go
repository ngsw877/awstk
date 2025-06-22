package rds

// RdsInstance RDSインスタンスの情報を格納する構造体
type RdsInstance struct {
	InstanceId string
	Engine     string
	Status     string
}
