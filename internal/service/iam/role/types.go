package role

import "time"

// ListOptions defines options for listing IAM roles.
type ListOptions struct {
	UnusedDays int
	Exclude    []string
}

// RoleItem represents a general IAM role info for list-all.
type RoleItem struct {
	Name            string
	Arn             string
	LastUsed        *time.Time
	IsServiceLinked bool
}

// UnusedRole represents an unused IAM role candidate.
type UnusedRole struct {
	Name     string
	Arn      string
	LastUsed *time.Time
}
