//go:build e2e

package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestE2E generates a real project against create-next-app@latest and fails
// when the upstream output no longer matches what the CLI expects.
//
//	go test -tags e2e -run TestE2E ./cmd/launchbox                # pnpm runner
//	E2E_RUNNER=docker go test -tags e2e -run TestE2E ./cmd/launchbox
//
// Set E2E_DIR to keep the generated project (e.g. to build it afterwards).
func TestE2E(t *testing.T) {
	dir := os.Getenv("E2E_DIR")
	if dir == "" {
		dir = t.TempDir()
	}

	projectDir, projectName = dir, "e2e-app"
	projectPath = filepath.Join(projectDir, projectName)
	warnings, nextVersion = nil, ""
	spinner = NewSpinner("dots")

	var err error
	switch runner := os.Getenv("E2E_RUNNER"); runner {
	case "", "pnpm":
		err = createProjectWithPnpm()
	case "docker":
		err = createProjectWithDocker()
	default:
		t.Fatalf("unknown E2E_RUNNER %q", runner)
	}
	spinner.Stop()

	if err != nil {
		t.Fatal(err)
	}
	t.Logf("next %s", nextVersion)
	for _, w := range warnings {
		t.Errorf("create-next-app output changed: %s", w)
	}

	for _, f := range []string{
		"server.js",
		".env.example",
		".husky/pre-commit",
		".husky/commit-msg",
		"node_modules/.bin/next",
		"node_modules/express",
		"app/(pages)/page.tsx",
	} {
		if _, err := os.Stat(filepath.Join(projectPath, f)); err != nil {
			t.Errorf("expected %s: %v", f, err)
		}
	}

	if uid, ok := fileOwnerUID(filepath.Join(projectPath, "package.json")); ok && uid != os.Getuid() {
		t.Errorf("package.json owned by uid %d, want %d", uid, os.Getuid())
	}

	// Files create-next-app starts adding go unnoticed otherwise (like AGENTS.md did)
	for _, dir := range []string{".", "app"} {
		entries, err := os.ReadDir(filepath.Join(projectPath, dir))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			name := filepath.Join(dir, entry.Name())
			if !knownFiles[name] {
				t.Errorf("new file from create-next-app: %s (handle it in the CLI, then add it to knownFiles)", name)
			}
		}
	}
}

// knownFiles is everything a generated project has at its root and in app/,
// from create-next-app plus the template.
var knownFiles = map[string]bool{
	".editorconfig":           true,
	".env.example":            true,
	".git":                    true,
	".gitignore":              true,
	".husky":                  true,
	"AGENTS.md":               true,
	"CLAUDE.md":               true,
	"Procfile":                true,
	"README.md":               true,
	"app":                     true,
	"commitlint.config.mjs":   true,
	"docker":                  true,
	"docker-compose.yml":      true,
	"eslint.config.mjs":       true,
	"lint-staged.config.mjs":  true,
	"next-env.d.ts":           true,
	"next-sitemap.config.mjs": true,
	"next.config.mjs":         true,
	"node_modules":            true,
	"package.json":            true,
	"pnpm-lock.yaml":          true,
	"pnpm-workspace.yaml":     true,
	"postcss.config.mjs":      true,
	"prettier.config.mjs":     true,
	"public":                  true,
	"redirects.json":          true,
	"scripts":                 true,
	"server.js":               true,
	"tsconfig.json":           true,
	"app/(pages)":             true,
	"app/components":          true,
	"app/config":              true,
	"app/favicon.ico":         true,
	"app/globals.css":         true,
	"app/layout.tsx":          true,
	"app/not-found.tsx":       true,
	"app/stylesheets":         true,
}
