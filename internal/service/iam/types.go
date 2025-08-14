package iam

import "time"

// ListOptions defines filter options for listing unused IAM resources.
type ListOptions struct {
	Roles      bool
	Policies   bool
	UnusedDays int
	Exclude    []string
}

// UnusedResult aggregates unused roles and policies.
type UnusedResult struct {
	Roles    []UnusedRole
	Policies []UnusedPolicy
}

// UnusedRole represents an unused IAM role candidate.
type UnusedRole struct {
	Name     string
	Arn      string
	LastUsed *time.Time // nil means never used
	Reason   string     // optional: exclusion/skip reasons
}

// UnusedPolicy represents an unused IAM customer managed policy.
type UnusedPolicy struct {
	Name string
	Arn  string
	Note string // e.g., "未アタッチ"
}

// AllResult aggregates full lists of roles and policies.
type AllResult struct {
	Roles    []RoleItem
	Policies []PolicyItem
}

// RoleItem represents a general IAM role info for list-all.
type RoleItem struct {
	Name            string
	Arn             string
	LastUsed        *time.Time // populated when Details is true (via GetRole)
	IsServiceLinked bool
}

// PolicyItem represents a general IAM policy info for list-all.
type PolicyItem struct {
	Name            string
	Arn             string
	AttachmentCount int32
}
