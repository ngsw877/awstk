package policy

// ListOptions IamPolicyListOptions IAMポリシー一覧取得時のオプション
type ListOptions struct {
	UnattachedOnly bool
	Exclude        []string
}

// PolicyItem IamPolicy IAMポリシー一覧表示用の情報
type PolicyItem struct {
	Name            string
	Arn             string
	AttachmentCount int32
}

// UnusedPolicy IamPolicyUnused 未使用IAMポリシーの情報
type UnusedPolicy struct {
	Name string
	Arn  string
	Note string
}
