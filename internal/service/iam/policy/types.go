package policy

// ListOptions defines options for listing IAM policies.
type ListOptions struct {
	UnattachedOnly bool
	Exclude        []string
}

// PolicyItem represents a general IAM policy info for list-all.
type PolicyItem struct {
	Name            string
	Arn             string
	AttachmentCount int32
}

// UnusedPolicy represents an unused IAM customer managed policy.
type UnusedPolicy struct {
	Name string
	Arn  string
	Note string
}
