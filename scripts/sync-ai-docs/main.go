package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func main() {
	if err := syncDocs(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func syncDocs() error {
	rulesDir := ".cursor/rules"
	claudeFile := "CLAUDE.md"

	// .mdc ファイルを全て取得
	files, err := filepath.Glob(filepath.Join(rulesDir, "*.mdc"))
	if err != nil {
		return fmt.Errorf("failed to find .mdc files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no .mdc files found in %s", rulesDir)
	}

	// ファイル名でソート（一貫した順序にするため）
	sort.Strings(files)

	var content strings.Builder

	for i, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		// frontmatter除去して結合
		cleanContent := removeFrontmatter(string(data))
		content.WriteString(cleanContent)

		// 最後のファイル以外は改行を追加
		if i < len(files)-1 {
			content.WriteString("\n\n")
		}
	}

	// CLAUDE.md に書き込み
	if err := os.WriteFile(claudeFile, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write CLAUDE.md: %w", err)
	}

	fmt.Println("CLAUDE.md has been updated successfully")
	return nil
}

func removeFrontmatter(content string) string {
	// YAML frontmatter を除去
	re := regexp.MustCompile(`(?s)^---\n.*?\n---\n`)
	return re.ReplaceAllString(content, "")
}
