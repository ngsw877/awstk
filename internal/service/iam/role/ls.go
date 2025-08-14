package role

import (
	iamsvc "awstk/internal/service/iam"
	"context"
	"fmt"

	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
)

// ListOptions defines options for listing IAM roles.
type ListOptions struct {
	UnusedDays int
	Exclude    []string
}

// List prints IAM roles. If UnusedDays > 0, it prints only unused roles.
func List(client *sdkiam.Client, opts ListOptions) error {
	// Ensure client is not nil
	if client == nil {
		return fmt.Errorf("iam client is nil")
	}

	// Sanity: small no-op to reference context package per style (not strictly needed)
	_ = context.Background()

	listOpts := iamsvc.ListOptions{
		Roles:      true,
		Policies:   false,
		UnusedDays: opts.UnusedDays,
		Exclude:    opts.Exclude,
	}

	if opts.UnusedDays > 0 {
		res, err := iamsvc.ListUnused(client, listOpts)
		if err != nil {
			return fmt.Errorf("未使用ロール一覧の取得に失敗: %w", err)
		}
		iamsvc.DisplayUnused(res, listOpts)
		return nil
	}

	all, err := iamsvc.ListAll(client, listOpts)
	if err != nil {
		return fmt.Errorf("ロール一覧の取得に失敗: %w", err)
	}
	iamsvc.DisplayAll(all, listOpts)
	return nil
}
