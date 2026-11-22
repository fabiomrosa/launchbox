package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// The CLI patches the output of create-next-app@latest, so it relies on files
// and settings that upstream may rename or drop. Instead of failing, each broken
// assumption is recorded here and reported once the project is created.
var (
	warnings    []string
	nextVersion string // "next" version found in the generated package.json
)

func warnf(format string, args ...any) {
	warnings = append(warnings, fmt.Sprintf(format, args...))
}

// relToProject returns path relative to the project, for readable warnings.
func relToProject(path string) string {
	if rel, err := filepath.Rel(projectPath, path); err == nil {
		return rel
	}
	return path
}

// removeExpected removes a file create-next-app is expected to generate,
// warning when it's missing.
func removeExpected(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		warnf("%s not found (expected to remove it)", relToProject(path))
		return nil
	}
	if err != nil {
		return fmt.Errorf("Failed to remove %s", path)
	}
	return nil
}

func printWarnings() {
	if len(warnings) == 0 {
		return
	}

	version := nextVersion
	if version == "" {
		version = "unknown"
	}

	fmt.Printf("⚠️ \033[1;33mcreate-next-app output changed (next %s) — check the template:\033[0m\n", version)
	for _, w := range warnings {
		fmt.Println("   •", w)
	}
}
