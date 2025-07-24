package rds

// Instance RdsInstance RDSインスタンスの情報を格納する構造体
type Instance struct {
	InstanceId string
	Engine     string
	Status     string
}
