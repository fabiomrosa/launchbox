#!/bin/bash

# --- Constants ---
readonly APP_NAME="launchbox"
readonly REPOSITORY_URL="https://github.com/fabiomrosa/launchbox"
readonly CHANGELOG_FILE="CHANGELOG.md"
readonly HOMEBREW_FORMULA="formula/${APP_NAME}.rb"
readonly PACKAGE_JSON="package.json"
readonly OUTPUT_DIR="./bin"
readonly PLATFORMS=("darwin/arm64" "darwin/amd64" "linux/amd64")
readonly GO_WORKDIR="cli"
readonly GO_ENTRYPOINT="cmd/launchbox"

# --- Formatting ---
readonly _RESET=$(tput sgr0)
readonly _RED=$(tput setaf 1)
readonly _GREEN=$(tput setaf 2)
readonly _YELLOW=$(tput setaf 3)
readonly _BOLD=$(tput bold)
readonly _UNDER=$(tput smul)
readonly _XUNDER=$(tput rmul)
