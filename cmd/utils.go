package cmd

import "awsfunc/internal"

// region, profile から AwsContext を生成する共通関数
func getAwsContext() internal.AwsContext {
	return internal.AwsContext{
		Region:  region,
		Profile: profile,
	}
}
