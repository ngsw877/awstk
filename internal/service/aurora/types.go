package aurora

// AuroraCluster Auroraクラスターの情報を格納する構造体
type AuroraCluster struct {
	ClusterId string
	Engine    string
	Status    string
}

// AuroraCapacityInfo Aurora Serverless v2のAcu情報を格納する構造体
type AuroraCapacityInfo struct {
	ClusterId    string
	CurrentAcu   float64
	MinAcu       float64
	MaxAcu       float64
	IsServerless bool
	Status       string
}
