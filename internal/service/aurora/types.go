package aurora



// Cluster AuroraCluster Auroraクラスターの情報を格納する構造体
type Cluster struct {
	ClusterId string
	Engine    string
	Status    string
}

// CapacityInfo AuroraCapacityInfo Aurora Serverless v2のAcu情報を格納する構造体
type CapacityInfo struct {
	ClusterId    string
	CurrentAcu   float64
	MinAcu       float64
	MaxAcu       float64
	IsServerless bool
	Status       string
}
