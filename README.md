<div align="center">
  <img alt="Launchbox" src="./launchbox.png" width="300" />
</div>

# Launchbox

> 🛠️ Command-line interface for creating boilerplate Next.js applications.

## Requirements

- macOS or Linux (on Windows, use [WSL](https://learn.microsoft.com/windows/wsl/))
- [Node.js](https://nodejs.org) v22+ with pnpm, or [Docker](https://www.docker.com) if you don’t want Node.js installed locally

## Install

Prebuilt binaries for the latest version are in [`bin/`](./bin):

```sh
# macOS (Apple Silicon)
install -m 755 bin/launchbox-darwin-arm64 /usr/local/bin/launchbox

# macOS (Intel)
install -m 755 bin/launchbox-darwin-amd64 /usr/local/bin/launchbox

# Linux
install -m 755 bin/launchbox-linux-amd64 /usr/local/bin/launchbox
```

Or build from source with Go 1.24+:

```sh
make build
```

## Distribution

> [!NOTE]
> This repository is kept as a portfolio piece; the package registries below are no longer published to.

Releases were fully automated: `release.sh` bumps the version, cross-compiles the binaries, updates the Homebrew formula checksums, generates the CHANGELOG from Conventional Commits and tags the release. Pushing the tag triggers [`release.yml`](./.github/workflows/release.yml), which is wired to publish to:

- **Homebrew**, via a dedicated tap (`brew tap fabiomrosa/launchbox`) using the formula in [`formula/`](./formula)
- **npm**, as `@fabiomrosa/launchbox`, where [`postinstall.js`](./postinstall.js) picks the right binary for the platform
- **GitHub Releases**, with the binaries attached

A weekly [upstream check](./.github/workflows/upstream-check.yml) generates a real project with the latest `create-next-app` and fails when its output drifts from what the CLI expects.

## Build and Release

### 🔖 Release

```sh
make major            # bump major version
make minor            # bump minor version
make patch            # bump patch version
make release v=1.2.3  # set an explicit version
```

### 🧪 Build only (compile without tagging)

```sh
make build
```

### 🧯 Dry run (safe mode) to verify what would happen, without making real changes

```sh
make dry v=major  # bump major version
make dry v=minor  # bump minor version
make dry v=patch  # bump patch version
make dry v=1.2.3  # set an explicit version
```

## Usage

> [!TIP]
> Don’t want to install Node.js, npm, or pnpm?
> You can run Launchbox via Docker instead — just make sure Docker is installed and running.

### 🚀 Run the CLI

```sh
launchbox
```

### 🔍 Check version

```sh
launchbox -v
launchbox --version
launchbox version
```

### 💻 Run CLI Locally (dev mode)

```sh
go run ./cli/cmd/launchbox
```
