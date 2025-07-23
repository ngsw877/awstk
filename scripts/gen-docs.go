package main

import (
	"awstk/cmd"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func main() {
	docsDir := "./docs"

	// 既存のdocsディレクトリをクリーン
	if err := os.RemoveAll(docsDir); err != nil {
		log.Fatalf("Failed to clean docs directory: %v", err)
	}

	// ディレクトリ作成
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		log.Fatalf("Failed to create docs directory: %v", err)
	}

	// サービスごとにドキュメントを生成
	serviceCommands := make(map[string][]*cobra.Command)
	
	// ルートコマンドは個別に生成
	if err := genSingleMarkdown(cmd.RootCmd, filepath.Join(docsDir, "awstk.md")); err != nil {
		log.Fatalf("Failed to generate root documentation: %v", err)
	}
	
	// サブコマンドをサービスごとにグループ化
	for _, subCmd := range cmd.RootCmd.Commands() {
		if !subCmd.IsAvailableCommand() || subCmd.IsAdditionalHelpTopicCommand() {
			continue
		}
		
		serviceName := subCmd.Name()
		serviceCommands[serviceName] = append(serviceCommands[serviceName], subCmd)
		
		// サブコマンドの子コマンドも収集
		for _, childCmd := range subCmd.Commands() {
			if childCmd.IsAvailableCommand() && !childCmd.IsAdditionalHelpTopicCommand() {
				serviceCommands[serviceName] = append(serviceCommands[serviceName], childCmd)
			}
		}
	}

	// サービスごとに単一ファイルを生成
	for serviceName, commands := range serviceCommands {
		filename := filepath.Join(docsDir, fmt.Sprintf("%s.md", serviceName))
		if err := genServiceMarkdown(serviceName, commands, filename); err != nil {
			log.Printf("Failed to generate documentation for %s: %v", serviceName, err)
			continue
		}
	}

	fileCount := len(serviceCommands) + 1 // サービスファイル数 + awstk.md
	fmt.Printf("✅ Documentation generated in %s (%d files)\n", docsDir, fileCount)
}

// shouldRemoveInheritedFlags は継承フラグを削除すべきかチェック
func shouldRemoveInheritedFlags(cmdName string) bool {
	return cmdName == "env" || cmdName == "version"
}

// genSingleMarkdown は単一のコマンドのドキュメントを生成
func genSingleMarkdown(cmd *cobra.Command, filename string) error {
	buf := new(bytes.Buffer)
	if err := doc.GenMarkdown(cmd, buf); err != nil {
		return err
	}
	
	content := buf.String()
	if shouldRemoveInheritedFlags(cmd.Name()) {
		content = removeInheritedFlagsSection(content)
	}
	
	return os.WriteFile(filename, []byte(content), 0644)
}

// genServiceMarkdown はサービスの全コマンドを1つのファイルにまとめて生成
func genServiceMarkdown(serviceName string, commands []*cobra.Command, filename string) error {
	var content strings.Builder
	
	// ファイルヘッダー
	content.WriteString(fmt.Sprintf("# %s Commands\n\n", serviceName))
	content.WriteString(fmt.Sprintf("This document describes all `%s` related commands.\n\n", serviceName))
	content.WriteString("## Table of Contents\n\n")
	
	// 目次を生成
	for _, cmd := range commands {
		cmdPath := cmd.CommandPath()
		anchor := strings.ReplaceAll(cmdPath, " ", "-")
		content.WriteString(fmt.Sprintf("- [%s](#%s)\n", cmdPath, anchor))
	}
	content.WriteString("\n---\n\n")
	
	// 各コマンドのドキュメントを追加
	for _, cmd := range commands {
		buf := new(bytes.Buffer)
		if err := doc.GenMarkdown(cmd, buf); err != nil {
			return fmt.Errorf("failed to generate markdown for %s: %w", cmd.CommandPath(), err)
		}
		
		cmdDoc := buf.String()
		
		// envとversionコマンドの場合、継承フラグセクションを削除
		if shouldRemoveInheritedFlags(serviceName) || 
		   (cmd.Parent() != nil && shouldRemoveInheritedFlags(cmd.Parent().Name())) {
			cmdDoc = removeInheritedFlagsSection(cmdDoc)
		}
		
		// セクション区切りを追加
		content.WriteString(cmdDoc)
		content.WriteString("\n---\n\n")
	}
	
	return os.WriteFile(filename, []byte(content.String()), 0644)
}

// removeInheritedFlagsSection は継承フラグセクションを削除
func removeInheritedFlagsSection(content string) string {
	lines := strings.Split(content, "\n")
	result := []string{}
	inInheritedSection := false
	
	for _, line := range lines {
		if strings.HasPrefix(line, "### Options inherited from parent commands") {
			inInheritedSection = true
			continue
		}
		
		// 次のセクションに到達したら除外モードを解除
		if inInheritedSection && (strings.HasPrefix(line, "### ") || strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "##### ")) {
			inInheritedSection = false
		}
		
		if !inInheritedSection {
			result = append(result, line)
		}
	}
	
	return strings.Join(result, "\n")
}