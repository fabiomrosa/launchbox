package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// setupProject points the project globals at a temporary directory.
func setupProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	projectPath = dir
	projectAppPath = filepath.Join(dir, "app")
	projectPagesPath = filepath.Join(dir, "app", "(pages)")
	projectImagesPath = filepath.Join(dir, "public", "images")
	warnings, nextVersion = nil, ""
	spinner = NewSpinner("dots")
	t.Cleanup(spinner.Stop)
	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// fakeCreateNextApp writes the files create-next-app is expected to generate.
func fakeCreateNextApp(t *testing.T, dir string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "package.json"), `{"scripts":{"build":"next build"},"dependencies":{"next":"16.0.0"}}`)
	writeFile(t, filepath.Join(dir, "tsconfig.json"), `{"compilerOptions":{"paths":{"@/*":["./*"]}}}`)
	writeFile(t, filepath.Join(dir, ".gitignore"), "# env files\n.env*\n")
	for _, f := range []string{
		"next.config.ts", "app/layout.tsx", "app/page.tsx",
		"public/file.svg", "public/globe.svg", "public/vercel.svg", "public/window.svg", "public/next.svg",
	} {
		writeFile(t, filepath.Join(dir, f), "x")
	}
	if err := os.MkdirAll(filepath.Join(dir, ".husky"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestAddCustomFiles(t *testing.T) {
	dir := setupProject(t)

	if err := addCustomFiles(); err != nil {
		t.Fatal(err)
	}

	for _, f := range []string{
		".env.example",
		"server.js",
		"docker/Dockerfile",
		"scripts/dev-auto-port.sh",
		"app/layout.tsx",
		"app/components/GoogleTagManager/index.tsx",
		"app/(pages)/page.tsx",
		"public/images/opengraph-image.jpg",
	} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("expected %s: %v", f, err)
		}
	}

	info, err := os.Stat(filepath.Join(dir, "scripts/dev-auto-port.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Errorf("dev-auto-port.sh is not executable: %v", info.Mode())
	}
}

func TestUpdateGitignore(t *testing.T) {
	setupProject(t)
	path := filepath.Join(t.TempDir(), ".gitignore")
	writeFile(t, path, "# env files\n.env*\n\n# vercel\n.vercel\n")

	// Running twice must not duplicate the entries
	for range 2 {
		if err := updateGitignore(path); err != nil {
			t.Fatal(err)
		}
	}

	data, _ := os.ReadFile(path)
	want := "# env files\n.env*\n!.env.example\n\n# asdf\n.tool-versions\n\n# vercel\n.vercel\n"
	if string(data) != want {
		t.Errorf("got %q, want %q", data, want)
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
}

func TestUpdateGitignoreWarnings(t *testing.T) {
	setupProject(t)
	dir := t.TempDir()

	if err := updateGitignore(filepath.Join(dir, "missing")); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, ".gitignore")
	writeFile(t, path, "node_modules\n")
	if err := updateGitignore(path); err != nil {
		t.Fatal(err)
	}

	if len(warnings) != 2 {
		t.Errorf("expected 2 warnings, got %v", warnings)
	}
}

func TestConfigureProject(t *testing.T) {
	dir := setupProject(t)
	fakeCreateNextApp(t, dir)

	if err := configureProject(); err != nil {
		t.Fatal(err)
	}

	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
	if nextVersion != "16.0.0" {
		t.Errorf("nextVersion = %q", nextVersion)
	}
	for _, f := range []string{"next.config.ts", "app/page.tsx", "public/vercel.svg", "public/next.svg"} {
		if _, err := os.Stat(filepath.Join(dir, f)); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed", f)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "public/images/next.svg")); err != nil {
		t.Errorf("expected next.svg moved to public/images: %v", err)
	}

	tsconfig, _ := os.ReadFile(filepath.Join(dir, "tsconfig.json"))
	if !strings.Contains(string(tsconfig), `"./app/*"`) {
		t.Errorf("tsconfig alias not updated: %s", tsconfig)
	}
}

func TestConfigureProjectWarnsOnChangedOutput(t *testing.T) {
	dir := setupProject(t)
	writeFile(t, filepath.Join(dir, "package.json"), `{}`)
	writeFile(t, filepath.Join(dir, "tsconfig.json"), `{}`)
	if err := os.MkdirAll(filepath.Join(dir, ".husky"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := configureProject(); err != nil {
		t.Fatal(err)
	}

	// next dependency, build script, .gitignore, tsconfig paths,
	// 3 root/app files, 4 public svgs and next.svg
	if len(warnings) != 12 {
		t.Errorf("expected 12 warnings, got %d: %v", len(warnings), warnings)
	}
}

func TestResolvePathFromHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	for input, want := range map[string]string{
		"~/code/app":             filepath.Join(home, "code/app"),
		"code/app":               filepath.Join(home, "code/app"),
		"/Desktop":               filepath.Join(home, "Desktop"),
		filepath.Join(home, "x"): filepath.Join(home, "x"),
		home + "2/x":             filepath.Join(home, strings.TrimPrefix(home+"2/x", "/")),
	} {
		got, err := resolvePathFromHome(input)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("resolvePathFromHome(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestFileOwnerUID(t *testing.T) {
	uid, ok := fileOwnerUID(t.TempDir())
	if !ok || uid != os.Getuid() {
		t.Errorf("fileOwnerUID = %d, %v; want %d, true", uid, ok, os.Getuid())
	}
	if _, ok := fileOwnerUID(filepath.Join(t.TempDir(), "missing")); ok {
		t.Error("expected false for a missing path")
	}
}

func TestCreateNextAppFlags(t *testing.T) {
	args := pnpmSteps()[0].args

	// create-next-app silently ignores unknown flags, so guard the ones that matter
	for _, flag := range []string{"--skip-install", "--yes", "--ts", "--app", "--tailwind", "--eslint"} {
		if !slices.Contains(args, flag) {
			t.Errorf("create-next-app args missing %s: %v", flag, args)
		}
	}
}
