package route53

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

// HostedZoneInfo HostedZoneInfoはRoute53ホストゾーンの情報を保持します
type HostedZoneInfo struct {
	Id          string
	Name        string
	RecordCount int64
	IsPrivate   bool
	Comment     string
	CallerRef   string
	CreatedDate time.Time
}

// RecordSetInfo RecordSetInfoはRoute53リソースレコードセットの情報を保持します
type RecordSetInfo struct {
	Name          string
	Type          types.RRType
	TTL           *int64
	Records       []string
	AliasTarget   *types.AliasTarget
	SetIdentifier *string
	Weight        *int64
	Region        types.ResourceRecordSetRegion
	Failover      types.ResourceRecordSetFailover
	HealthCheckId *string
}

// DeleteOptions DeleteOptionsは削除操作のオプションを保持します
type DeleteOptions struct {
	UseId  bool
	Force  bool
	DryRun bool
}

// DeleteResult DeleteResultは削除操作の結果を保持します
type DeleteResult struct {
	ZoneId         string
	ZoneName       string
	RecordsDeleted int
	RecordsFailed  int
	Success        bool
	Error          error
}
