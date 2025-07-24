package main

import (
	"awstk/cmd"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
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

	// ルートコマンドはdocs/README.mdとして生成
	if err := genSingleMarkdown(cmd.RootCmd, filepath.Join(docsDir, "README.md")); err != nil {
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

	fileCount := len(serviceCommands) + 1 // サービスファイル数 + README.md
	fmt.Printf("✅ Documentation generated in %s (%d files)\n", docsDir, fileCount)
}

// shouldRemoveInheritedFlags は継承フラグを削除すべきかチェック
func shouldRemoveInheritedFlags(cmdName string) bool {
	return cmdName == "env" || cmdName == "version"
}

// customLinkHandler はドキュメント内のリンクをカスタマイズ
func customLinkHandler(name string) string {
	// awstk -> README
	if name == "awstk" {
		return "README"
	}

	// awstk_aurora_ls -> aurora#awstk-aurora-ls
	// awstk_aurora -> aurora
	parts := strings.Split(name, "_")
	if len(parts) >= 2 && parts[0] == "awstk" {
		serviceName := parts[1]

		// サブコマンドがある場合はアンカー付きリンク（同一ファイル内）
		if len(parts) > 2 {
			anchor := strings.ReplaceAll(name, "_", "-")
			return serviceName + "#" + anchor
		}
		// サービスレベルのコマンド（別ファイル）
		return serviceName
	}

	// デフォルト
	return name
}

// genSingleMarkdown は単一のコマンドのドキュメントを生成
func genSingleMarkdown(cmd *cobra.Command, filename string) error {
	buf := new(bytes.Buffer)
	// カスタムリンク関数でリンク形式を調整
	if err := doc.GenMarkdownCustom(cmd, buf, customLinkHandler); err != nil {
		return err
	}

	content := buf.String()
	if shouldRemoveInheritedFlags(cmd.Name()) {
		content = removeInheritedFlagsSection(content)
	}

	// リンクの修正
	content = fixMarkdownLinks(content)

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
		// カスタムリンク関数でリンク形式を調整
		if err := doc.GenMarkdownCustom(cmd, buf, customLinkHandler); err != nil {
			return fmt.Errorf("failed to generate markdown for %s: %w", cmd.CommandPath(), err)
		}

		cmdDoc := buf.String()

		// envとversionコマンドの場合、継承フラグセクションを削除
		if shouldRemoveInheritedFlags(serviceName) ||
			(cmd.Parent() != nil && shouldRemoveInheritedFlags(cmd.Parent().Name())) {
			cmdDoc = removeInheritedFlagsSection(cmdDoc)
		}

		// リンクの修正
		cmdDoc = fixMarkdownLinks(cmdDoc)

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

// fixMarkdownLinks はMarkdown内のリンクを修正
func fixMarkdownLinks(content string) string {
	// awstk.md -> README.md
	content = strings.ReplaceAll(content, "(awstk.md)", "(README.md)")

	// serviceName#anchor.md -> serviceName.md#anchor
	// 例: aurora#awstk-aurora-ls.md -> aurora.md#awstk-aurora-ls
	re := regexp.MustCompile(`\((\w+)#([\w-]+)\.md\)`)
	content = re.ReplaceAllString(content, "($1.md#$2)")

	return content
}
