/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"awsfunc/cmd"
	_ "awsfunc/cmd/s3"
)

func main() {
	cmd.Execute()
}
