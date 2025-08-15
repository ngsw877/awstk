package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// This tool syncs project-level .mcp.json to .cursor/mcp.json with basic JSON validation.
func main() {
	const src = ".mcp.json"
	dstDir := ".cursor"
	dst := filepath.Join(dstDir, "mcp.json")

	content, err := os.ReadFile(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "missing %s: %v\n", src, err)
		os.Exit(1)
	}

	// Validate JSON to prevent committing malformed configuration
	var v any
	if err := json.Unmarshal(content, &v); err != nil {
		fmt.Fprintf(os.Stderr, "invalid JSON in %s: %v\n", src, err)
		os.Exit(1)
	}

	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create %s: %v\n", dstDir, err)
		os.Exit(1)
	}

	if err := os.WriteFile(dst, content, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", dst, err)
		os.Exit(1)
	}

	fmt.Println("✅ Synced .mcp.json -> .cursor/mcp.json")
}
