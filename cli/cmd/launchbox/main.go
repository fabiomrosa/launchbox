package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var Version = "dev"

const (
	DefaultNodeEngine = "24.x"
	DefaultPnpmEngine = "pnpm@10.25.0"
)

var (
	projectDir        string
	projectName       string
	projectRunner     string
	projectPath       string
	projectAppPath    string
	projectPagesPath  string
	projectPublicPath string
	projectImagesPath string
)

var spinner *Spinner

//go:embed all:template
var templateFiles embed.FS

func main() {
	spinner = NewSpinner("dots")

	version := fmt.Sprintf("🚀 Launchbox CLI — version: \033[1;32m%s\033[0m\n", Version)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Print(version)
			os.Exit(0)
		}
	}

	fmt.Println(version)

	var confirmation bool

	projectDefaultName := "my-app"

	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("📁 Type project directory (~/):").
				Placeholder("path/to/directory").
				Value(&projectDir).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					if s == "" {
						return fmt.Errorf("directory is required")
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("📦 Type project name:").
				Placeholder(projectDefaultName).
				Value(&projectName),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("⚙️ Choose your runner:").
				Options(
					huh.NewOption("🐳 Docker", "docker"),
					huh.NewOption("📦 PnPM", "pnpm"),
				).
				Value(&projectRunner),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("All done?").
				Affirmative("Yep").
				Negative("No, skip").
				Value(&confirmation),
		),
	).Run(); err != nil {
		log.Fatal("❌ \033[1;31mError getting inputs for project:\033[0m ", err)
	}

	resolvedPath, err := resolvePathFromHome(projectDir)
	if err != nil {
		log.Fatal("❌ \033[1;31mError resolving path:\033[0m ", err)
	}
	projectDir = filepath.Clean(resolvedPath)

	projectName = strings.TrimSpace(projectName)
	if projectName == "" {
		projectName = projectDefaultName
	}

	if !confirmation {
		log.Fatal("⚠️ \033[1;33mSkipped project creation.\033[0m")
	}

	projectPath = filepath.Join(projectDir, projectName)

	// Never touch an existing directory: on failure the project path is removed.
	if _, err := os.Stat(projectPath); err == nil {
		log.Fatalf("❌ \033[1;31mDirectory already exists:\033[0m %s", projectPath)
	} else if !os.IsNotExist(err) {
		log.Fatal("❌ \033[1;31mError checking project directory:\033[0m ", err)
	}

	switch projectRunner {
	case "docker":
		err = createProjectWithDocker()
	case "pnpm":
		err = createProjectWithPnpm()
	default:
		err = fmt.Errorf("unknown runner: %q", projectRunner)
	}

	if err != nil {
		spinner.Stop()
		printWarnings()
		if rmErr := os.RemoveAll(projectPath); rmErr != nil {
			log.Printf("⚠️ \033[1;33mFailed to remove %s:\033[0m %v", projectPath, rmErr)
		}
		log.Fatal("❌ \033[1;31mError creating project:\033[0m ", err)
	}

	spinner.Stop()

	clickableProjectName := fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", projectPath, projectName)
	fmt.Println("✅ \033[1;32mProject created successfully:\033[0m", clickableProjectName)
	printWarnings()
}

func resolvePathFromHome(path string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Failed to get home directory: %w", err)
	}

	// Handles "~/path"
	if strings.HasPrefix(path, "~") {
		return filepath.Join(home, strings.TrimPrefix(path, "~")), nil
	}
	if filepath.IsAbs(path) {
		// Paths inside home are kept as they are
		if path == home || strings.HasPrefix(path, home+string(filepath.Separator)) {
			return path, nil
		}
		// Any other absolute path is treated as relative to home, e.g. "/Desktop" → "~/Desktop"
		return filepath.Join(home, strings.TrimPrefix(path, "/")), nil
	}

	return filepath.Join(home, path), nil
}

// pnpmStep is a pnpm command shared by both runners, so they can't drift apart.
type pnpmStep struct {
	message string
	inDir   bool // run in projectDir (the parent) instead of projectPath
	args    []string
	failure string
}

func pnpmSteps() []pnpmStep {
	return []pnpmStep{
		{
			message: "⚙️ Generating Next.js project using create-next-app...",
			inDir:   true,
			args: []string{
				"dlx", "create-next-app@latest", projectName,
				"--skip-install", "--yes", "--ts", "--app", "--tailwind", "--eslint",
			},
			failure: "Failed to create project using create-next-app",
		},
		{
			message: "📦 Add dependencies via pnpm...",
			args: []string{
				"add",
				"body-parser", "compression", "cors", "express",
			},
			failure: "Failed to install dependencies",
		},
		{
			message: "📦 Add dev dependencies via pnpm...",
			args: []string{
				"add", "-D",
				"husky", "prettier", "lint-staged",
				"@commitlint/cli", "@commitlint/config-conventional",
				"@trivago/prettier-plugin-sort-imports", "next-sitemap",
			},
			failure: "Failed to install dev dependencies",
		},
		{
			message: "📦 Running pnpm install to populate .bin...",
			args:    []string{"install"},
			failure: "Failed to run pnpm install",
		},
		{
			message: "🐶 Initializing Husky...",
			args:    []string{"exec", "husky", "init"},
			failure: "Failed to init Husky",
		},
	}
}

func (s pnpmStep) dir() string {
	if s.inDir {
		return projectDir
	}
	return projectPath
}

// run executes cmd, hiding its output unless this is a debug build.
func run(cmd *exec.Cmd) error {
	if !debug {
		cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	}
	return cmd.Run()
}

const dockerImageTemp = "transient-launchbox"

// dockerRun runs args in the transient image, with dir mounted at /app.
func dockerRun(dir string, args ...string) *exec.Cmd {
	return exec.Command("docker", append([]string{
		"run", "--rm",
		"-e", "PNPM_STORE_DIR=/root/.local/share/pnpm/store",
		"--mount", "type=tmpfs,destination=/root/.local/share/pnpm",
		"-v", dir + ":/app",
		"-w", "/app",
		dockerImageTemp,
	}, args...)...)
}

// fixOwnership hands the project back to the host user, since rootful
// Docker on Linux leaves the files it creates owned by root.
func fixOwnership() error {
	uid, ok := fileOwnerUID(projectPath)
	if !ok || uid == os.Getuid() {
		return nil
	}
	spinner.Start("🔑 Fixing file ownership...")
	owner := fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
	if err := run(dockerRun(projectPath, "chown", "-R", owner, "/app")); err != nil {
		return fmt.Errorf("Failed to fix file ownership")
	}
	return nil
}

func createProjectWithDocker() (err error) {
	// Step 1: Create temporary context for transient build
	tmpDir, err := os.MkdirTemp("", "docker-context-*")
	if err != nil {
		return fmt.Errorf("Failed creating temporary context")
	}
	defer os.RemoveAll(tmpDir)

	// Step 2: Save embedded Dockerfile into a temporary file
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	content, err := templateFiles.ReadFile("template/root/docker/Dockerfile")
	if err != nil {
		return fmt.Errorf("Failed reading embedded Dockerfile")
	}
	if err := os.WriteFile(dockerfilePath, content, 0o644); err != nil {
		return fmt.Errorf("Failed saving temporary Dockerfile")
	}

	// Step 3: Build transient image
	spinner.Start("🔨 Building transient Docker image...")
	buildCmd := exec.Command(
		"docker", "build",
		"--pull", "--no-cache",
		"-f", dockerfilePath,
		"-t", dockerImageTemp,
		tmpDir,
	)
	if err := run(buildCmd); err != nil {
		return fmt.Errorf("Failed building transient image")
	}
	// Remove the image once done; deferred first so it runs after fixOwnership, which needs it
	defer run(exec.Command("docker", "rmi", "-f", dockerImageTemp))

	// Step 4: Create the target directory to mount
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return fmt.Errorf("Failed to create target directory: %w", err)
	}

	// On failure, hand the files back too, so they can be removed
	defer func() {
		if err != nil {
			fixOwnership()
		}
	}()

	// Step 5: Run the pnpm steps inside the transient image
	for _, step := range pnpmSteps() {
		spinner.Start(step.message)
		if err := run(dockerRun(step.dir(), append([]string{"pnpm"}, step.args...)...)); err != nil {
			return fmt.Errorf("%s", step.failure)
		}
	}

	// Step 6: Hand files back to the host user
	if err := fixOwnership(); err != nil {
		return err
	}

	// Step 7: Configure project
	spinner.Start("🛠️ Configure project...")
	if err := configureProject(); err != nil {
		return err
	}

	return nil
}

func createProjectWithPnpm() error {
	// Step 1: Ensure the target directory exists
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return fmt.Errorf("Failed to create target directory: %w", err)
	}

	// Step 2: Run the pnpm steps
	for _, step := range pnpmSteps() {
		spinner.Start(step.message)
		cmd := exec.Command("pnpm", step.args...)
		cmd.Dir = step.dir()
		if err := run(cmd); err != nil {
			return fmt.Errorf("%s", step.failure)
		}
	}

	// Step 3: Configure project
	spinner.Start("🛠️ Configure project...")
	if err := configureProject(); err != nil {
		return err
	}

	return nil
}

func configureProject() error {
	// Ensure package.json exists
	packageJson := filepath.Join(projectPath, "package.json")
	if _, err := os.Stat(packageJson); os.IsNotExist(err) {
		return fmt.Errorf("package.json not found in %s", projectPath)
	}

	// Ensure tsconfig.json exists
	tsconfigJson := filepath.Join(projectPath, "tsconfig.json")
	if _, err := os.Stat(tsconfigJson); os.IsNotExist(err) {
		return fmt.Errorf("tsconfig.json not found in %s", projectPath)
	}

	// Step 1: Update .husky/pre-commit with lint-staged
	spinner.Start("🪝 Update pre-commit hook with lint-staged...")
	preCommitHookPath := filepath.Join(projectPath, ".husky", "pre-commit")
	preCommitContent := strings.TrimSpace(`
		pnpm lint-staged
	`) + "\n" // <- Shell scripts must end with a newline.
	if err := os.WriteFile(preCommitHookPath, []byte(preCommitContent), 0o755); err != nil {
		return fmt.Errorf("Failed to write pre-commit hook")
	}
	if err := os.Chmod(preCommitHookPath, 0o755); err != nil {
		return fmt.Errorf("Failed to chmod pre-commit hook")
	}

	// Step 2: Create .husky/commit-msg with commitlint
	spinner.Start("🪝 Create commit-msg hook with commitlint...")
	commitMsgHookPath := filepath.Join(projectPath, ".husky", "commit-msg")
	commitMsgContent := strings.TrimSpace(`
		pnpm commitlint --edit "$1"
	`) + "\n" // <- Shell scripts must end with a newline.
	if err := os.WriteFile(commitMsgHookPath, []byte(commitMsgContent), 0o755); err != nil {
		return fmt.Errorf("Failed to write commit-msg hook")
	}
	if err := os.Chmod(commitMsgHookPath, 0o755); err != nil {
		return fmt.Errorf("Failed to chmod commit-msg hook")
	}

	// Step 3: Add configs and format package.json
	spinner.Start("📝 Add configs and format package.json")
	data, err := os.ReadFile(packageJson)
	if err != nil {
		return fmt.Errorf("Failed to read package.json")
	}
	nextVersion = gjson.GetBytes(data, "dependencies.next").String()
	if nextVersion == "" {
		warnf("package.json has no \"next\" dependency")
	}
	// postbuild and heroku-postbuild rely on the build script
	if !gjson.GetBytes(data, "scripts.build").Exists() {
		warnf("package.json has no scripts.build")
	}
	// Add scripts.dev
	data, err = sjson.SetBytesOptions(
		data,
		"scripts.dev", "./scripts/dev-auto-port.sh",
		&sjson.Options{Optimistic: true},
	)
	if err != nil {
		return fmt.Errorf("Failed to set scripts.dev")
	}
	// Add scripts.start
	data, err = sjson.SetBytesOptions(
		data,
		"scripts.start", "NODE_ENV=production node server.js",
		&sjson.Options{Optimistic: true},
	)
	if err != nil {
		return fmt.Errorf("Failed to set scripts.start")
	}
	// Add scripts.postbuild
	data, err = sjson.SetBytesOptions(
		data,
		"scripts.postbuild", "next-sitemap --config next-sitemap.config.mjs",
		&sjson.Options{Optimistic: true},
	)
	if err != nil {
		return fmt.Errorf("Failed to set scripts.postbuild")
	}
	// Add scripts.heroku-postbuild
	data, err = sjson.SetBytesOptions(
		data,
		"scripts.heroku-postbuild", "pnpm run build",
		&sjson.Options{Optimistic: true},
	)
	if err != nil {
		return fmt.Errorf("Failed to set scripts.heroku-postbuild")
	}
	// Add engines.node
	data, err = sjson.SetBytesOptions(
		data,
		"engines.node", DefaultNodeEngine,
		&sjson.Options{Optimistic: true},
	)
	if err != nil {
		return fmt.Errorf("Failed to set engines.node")
	}
	// Add packageManager
	data, err = sjson.SetBytesOptions(
		data,
		"packageManager", DefaultPnpmEngine,
		&sjson.Options{Optimistic: true},
	)
	if err != nil {
		return fmt.Errorf("Failed to set packageManager")
	}
	var updatedPackageJson bytes.Buffer
	if err := json.Indent(&updatedPackageJson, data, "", "  "); err != nil {
		return fmt.Errorf("Failed to format package.json")
	}
	if err := os.WriteFile(packageJson, updatedPackageJson.Bytes(), 0o644); err != nil {
		return fmt.Errorf("Failed to write updated package.json")
	}

	// Step 4: Add .env.example and asdf entries to .gitignore
	spinner.Start("📝 Update .gitignore")
	if err := updateGitignore(filepath.Join(projectPath, ".gitignore")); err != nil {
		return err
	}

	// Step 5: Change import alias config in tsconfig.json
	spinner.Start("📝 Change import alias config in tsconfig.json")
	data, err = os.ReadFile(tsconfigJson)
	if err != nil {
		return fmt.Errorf("Failed to read tsconfig.json")
	}
	if !gjson.GetBytes(data, "compilerOptions.paths").Exists() {
		warnf("tsconfig.json has no compilerOptions.paths (expected to replace the @/* alias)")
	}
	// Change import alias root
	data, err = sjson.SetBytesOptions(
		data,
		"compilerOptions.paths.@/*", []string{"./app/*"},
		&sjson.Options{Optimistic: true},
	)
	if err != nil {
		return fmt.Errorf("Failed to reset compilerOptions.paths.@/*")
	}
	// Add baseUrl
	data, err = sjson.SetBytesOptions(
		data,
		"compilerOptions.baseUrl", ".",
		&sjson.Options{Optimistic: true},
	)
	if err != nil {
		return fmt.Errorf("Failed to set compilerOptions.baseUrl")
	}
	var updatedTsconfigJson bytes.Buffer
	if err := json.Indent(&updatedTsconfigJson, data, "", "  "); err != nil {
		return fmt.Errorf("Failed to format tsconfig.json")
	}
	if err := os.WriteFile(tsconfigJson, updatedTsconfigJson.Bytes(), 0o644); err != nil {
		return fmt.Errorf("Failed to write updated tsconfig.json")
	}

	// Step 6: Remove files in root directory
	spinner.Start("📁 Customize project directory")
	// The template's next.config.mjs wires Tailwind through @tailwindcss/turbopack,
	// mirroring the next.config.ts it replaces
	if nextConfig, err := os.ReadFile(filepath.Join(projectPath, "next.config.ts")); err == nil &&
		!bytes.Contains(nextConfig, []byte("@tailwindcss/turbopack")) {
		warnf("next.config.ts no longer uses @tailwindcss/turbopack (template next.config.mjs relies on it)")
	}
	for _, f := range []string{
		filepath.Join(projectPath, "next.config.ts"),
	} {
		if err := removeExpected(f); err != nil {
			return err
		}
	}

	// Step 7: Remove files and customize app directory
	spinner.Start("📁 Customize app directory")
	projectAppPath = filepath.Join(projectPath, "app")
	for _, f := range []string{
		filepath.Join(projectAppPath, "layout.tsx"),
		filepath.Join(projectAppPath, "page.tsx"),
	} {
		if err := removeExpected(f); err != nil {
			return err
		}
	}
	projectPagesPath = filepath.Join(projectAppPath, "(pages)")
	if err := os.MkdirAll(projectPagesPath, 0o755); err != nil {
		return fmt.Errorf("Failed to create app/(pages) folder")
	}

	// Step 8: Customize public directory
	spinner.Start("📁 Customize public directory")
	projectPublicPath = filepath.Join(projectPath, "public")
	for _, f := range []string{
		filepath.Join(projectPublicPath, "file.svg"),
		filepath.Join(projectPublicPath, "globe.svg"),
		filepath.Join(projectPublicPath, "vercel.svg"),
		filepath.Join(projectPublicPath, "window.svg"),
	} {
		if err := removeExpected(f); err != nil {
			return err
		}
	}
	projectImagesPath = filepath.Join(projectPublicPath, "images")
	if err := os.MkdirAll(projectImagesPath, 0o755); err != nil {
		return fmt.Errorf("Failed to create public/images folder")
	}
	if err := os.Rename(
		filepath.Join(projectPublicPath, "next.svg"),
		filepath.Join(projectImagesPath, "next.svg"),
	); os.IsNotExist(err) {
		warnf("public/next.svg not found (expected to move it to public/images)")
	} else if err != nil {
		return fmt.Errorf("Failed to move next.svg")
	}

	// Step 9: Copy template files
	spinner.Start("📋 Copy template files...")
	if err := addCustomFiles(); err != nil {
		return err
	}

	return nil
}

func updateGitignore(gitignorePath string) error {
	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			warnf(".gitignore not found")
			return nil
		}
		return fmt.Errorf("Failed to read .gitignore: %w", err)
	}
	content := string(data)
	if strings.Contains(content, "!.env.example") && strings.Contains(content, ".tool-versions") {
		return nil
	}
	lines := strings.Split(content, "\n")
	var out []string
	inserted := false
	for _, line := range lines {
		out = append(out, line)

		// Find the `.env*` line
		if strings.TrimSpace(line) == ".env*" && !inserted {
			out = append(out,
				"!.env.example",
				"",
				"# asdf",
				".tool-versions",
			)
			inserted = true
		}
	}
	if !inserted {
		warnf(".gitignore has no \".env*\" line (.env.example is not un-ignored)")
		return nil
	}
	newContent := strings.Join(out, "\n")
	if err := os.WriteFile(gitignorePath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("Failed to write updated .gitignore: %w", err)
	}

	return nil
}

func addCustomFiles() error {
	type group struct {
		dest    string
		baseDir string   // e.g. "template/root"
		files   []string // relative to baseDir
	}

	const (
		rootDir   = "template/root"
		appDir    = "template/app"
		pagesDir  = "template/pages"
		imagesDir = "template/images"
	)

	groups := []group{
		{
			dest:    projectPath,
			baseDir: rootDir,
			files: []string{
				"README.md",
				"Procfile",
				"scripts",
				"docker",
				"docker-compose.yml",
				".editorconfig",
				".env.example",
				"commitlint.config.mjs",
				"lint-staged.config.mjs",
				"next-sitemap.config.mjs",
				"next.config.mjs",
				"prettier.config.mjs",
				"redirects.json",
				"server.js",
			},
		},
		{
			dest:    projectAppPath,
			baseDir: appDir,
			files: []string{
				"layout.tsx",
				"not-found.tsx",
				"globals.css",
				"components",
				"config",
				"stylesheets",
			},
		},
		{
			dest:    projectPagesPath,
			baseDir: pagesDir,
			files: []string{
				"page.tsx",
			},
		},
		{
			dest:    projectImagesPath,
			baseDir: imagesDir,
			files: []string{
				"opengraph-image.jpg",
				"noise-pattern.png",
			},
		},
	}

	for _, g := range groups {
		stripPrefix := g.baseDir + "/"

		for _, rel := range g.files {
			path := g.baseDir + "/" + rel

			if _, err := templateFiles.ReadDir(path); err == nil {
				if err := copyEmbeddedDir(path, g.dest, stripPrefix); err != nil {
					return err
				}
				continue
			}

			if err := copyEmbeddedFile(path, g.dest, stripPrefix); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyEmbeddedDir(dirPath, targetBase, stripPrefix string) error {
	entries, err := templateFiles.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("Failed to read embedded dir %s: %w", dirPath, err)
	}

	for _, entry := range entries {
		// embed.FS paths always use forward slashes, even on Windows
		entryPath := path.Join(dirPath, entry.Name())
		if entry.IsDir() {
			if err := copyEmbeddedDir(entryPath, targetBase, stripPrefix); err != nil {
				return err
			}
			continue
		}
		if err := copyEmbeddedFile(entryPath, targetBase, stripPrefix); err != nil {
			return err
		}
	}

	return nil
}

func copyEmbeddedFile(srcPath, targetBase, stripPrefix string) error {
	data, err := templateFiles.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("Failed to read embedded file %s: %w", srcPath, err)
	}

	relativePath := strings.TrimPrefix(srcPath, stripPrefix)
	destPath := filepath.Join(targetBase, filepath.FromSlash(relativePath))

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("Failed to create directory for %s: %w", destPath, err)
	}

	mode := os.FileMode(0o644)
	if strings.HasSuffix(relativePath, ".sh") ||
		strings.HasPrefix(relativePath, "scripts/") {
		mode = 0o755
	}

	if err := os.WriteFile(destPath, data, mode); err != nil {
		return fmt.Errorf("Failed to write file %s: %w", destPath, err)
	}

	return nil
}
