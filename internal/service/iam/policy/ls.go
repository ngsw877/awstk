package policy

import (
	iamsvc "awstk/internal/service/iam"
	"context"
	"fmt"

	sdkiam "github.com/aws/aws-sdk-go-v2/service/iam"
)

// ListOptions defines options for listing IAM policies.
type ListOptions struct {
	UnattachedOnly bool
	Exclude        []string
}

// List prints IAM customer managed policies. If UnattachedOnly is true, only unattached ones are printed.
func List(client *sdkiam.Client, opts ListOptions) error {
	if client == nil {
		return fmt.Errorf("iam client is nil")
	}

	_ = context.Background()

	listOpts := iamsvc.ListOptions{
		Roles:    false,
		Policies: true,
		Exclude:  opts.Exclude,
	}

	if opts.UnattachedOnly {
		res, err := iamsvc.ListUnused(client, listOpts)
		if err != nil {
			return fmt.Errorf("未アタッチポリシー一覧の取得に失敗: %w", err)
		}
		iamsvc.DisplayUnused(res, listOpts)
		return nil
	}

	all, err := iamsvc.ListAll(client, listOpts)
	if err != nil {
		return fmt.Errorf("ポリシー一覧の取得に失敗: %w", err)
	}
	iamsvc.DisplayAll(all, listOpts)
	return nil
}
