#!/bin/bash
set -e

# --- Load Envs ---
source "$(dirname "$0")/env.sh"

# --- Dry run mode ---
DRY_RUN="${DRY_RUN:-false}"

# --- Helpers ---
help_message() {
cat << EOF
Command-line interface for building and releasing the Launchbox CLI.

${_BOLD}${_GREEN}Env Variables:${_RESET}
    ${_BOLD}DRY_RUN=true${_RESET}    Preview changes without applying them

${_BOLD}${_GREEN}Usage:${_RESET}
    $0 ${_BOLD}<command>${_RESET}

${_BOLD}${_GREEN}Commands:${_RESET}
    ${_BOLD}-h, --help${_RESET}     Display this usage information
    ${_BOLD}patch${_RESET}          Auto-bump patch version
    ${_BOLD}minor${_RESET}          Auto-bump minor version
    ${_BOLD}major${_RESET}          Auto-bump major version
    ${_BOLD}x.y.z${_RESET}          Explicitly set version

EOF
}

safe_sed() {
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "$@"
  else
    sed -i "$@"
  fi
}

bump_version() {
  local ver=$1 bump=$2
  local major minor patch
  IFS='.' read -r major minor patch <<< "$ver"
  case "$bump" in
    patch) patch=$((patch + 1));;
    minor) minor=$((minor + 1)); patch=0;;
    major) major=$((major + 1)); minor=0; patch=0;;
    *) echo "❌ Invalid bump: $bump"; exit 1;;
  esac
  echo "${major}.${minor}.${patch}"
}

version_gt() {
  [ "$(printf '%s\n' "$1" "$2" | sort -V | head -n1)" != "$1" ]
}

# --- Get current version ---
CURRENT_VERSION=$(grep -oE '"version": *"[^"]+"' "$PACKAGE_JSON" | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')

# --- Determine new version ---
INPUT="$1"
if [[ -z "$INPUT" || "$INPUT" =~ ^(-h|--help)$ ]]; then
  help_message
  exit 0
elif [[ "$INPUT" =~ ^(patch|minor|major)$ ]]; then
  VERSION=$(bump_version "$CURRENT_VERSION" "$INPUT")
elif [[ "$INPUT" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  VERSION="$INPUT"
else
  echo "${_RED}❌ Usage:${_RESET} $0 [patch|minor|major|x.y.z]"
  exit 1
fi

# --- Normalize versions (strip leading v if present) ---
CURRENT_VERSION=${CURRENT_VERSION#v}
VERSION=${VERSION#v}
TAG="v$VERSION"

echo "🔢 Current version: $CURRENT_VERSION"
echo "🚀 New version:     $VERSION"

# --- Validate version bump ---
if [ "$VERSION" = "$CURRENT_VERSION" ]; then
  echo "${_RED}❌ Error:${_RESET} Version is the same as current"
  exit 1
fi

if ! version_gt "$VERSION" "$CURRENT_VERSION"; then
  echo "${_RED}❌ Error:${_RESET} Version $VERSION is not greater than $CURRENT_VERSION"
  exit 1
fi

# --- Check if tag already exists before continuing ---
if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "${_RED}❌ Error:${_RESET} Git tag $TAG already exists"
  exit 1
fi

# --- Check for uncommitted changes before modifying anything ---
if [[ -n "$(git status --porcelain)" && "$DRY_RUN" != true ]]; then
  echo "${_YELLOW}⚠️ Git working tree is dirty. Uncommitted changes may interfere with release.${_RESET}"
  read -rp "${_BOLD}🤔 Continue anyway?${_RESET} [y/N] " CONFIRM
  [[ "$CONFIRM" =~ ^[Yy]$ ]] || exit 1
fi

echo "📦 Building binaries..."

# --- Build binaries ---
if [[ "$DRY_RUN" == true ]]; then
  echo "🧪 Skipping binary build and SHA updates (dry run)"
else
  echo "📦 Building binaries for version: $VERSION"
  export VERSION="$VERSION"
  ./build.sh
fi

# --- Compute SHA256 and patch formula ---
for ARCH in arm64 amd64; do
  # only update Homebrew SHA for macOS binaries
  FILE="$OUTPUT_DIR/${APP_NAME}-darwin-${ARCH}"
  if [ ! -f "$FILE" ]; then
    if [[ "$DRY_RUN" == true ]]; then
      echo "🧪 Would skip SHA update (file not found in dry run): $FILE"
      continue
    else
      echo "⚠️  Skipping: $FILE not found."
      continue
    fi
  fi

  if [ ! -f "$HOMEBREW_FORMULA" ]; then
    if [[ "$DRY_RUN" == true ]]; then
      echo "🧪 Would skip formula SHA update (file not found in dry run)"
      continue
    else
      echo "${_RED}❌ Error:${_RESET} Formula not found at $HOMEBREW_FORMULA"
      exit 1
    fi
  fi

  SHA256=$(shasum -a 256 "$FILE" | awk '{print $1}')

  if [[ "$DRY_RUN" == true ]]; then
    echo "🧪 Would update SHA256 for macOS $ARCH → $SHA256"
    continue
  fi

  if [ "$ARCH" = "arm64" ]; then
    echo "🔐 Updating SHA256 for macOS ARM in formula…"
    sed -i.bak \
      "/on_arm do/,/end/ s|sha256 \".*\"|sha256 \"${SHA256}\"|" \
      "$HOMEBREW_FORMULA" && rm "${HOMEBREW_FORMULA}.bak"
  elif [ "$ARCH" = "amd64" ]; then
    echo "🔐 Updating SHA256 for macOS Intel in formula…"
    sed -i.bak \
      "/on_intel do/,/end/ s|sha256 \".*\"|sha256 \"${SHA256}\"|" \
      "$HOMEBREW_FORMULA" && rm "${HOMEBREW_FORMULA}.bak"
  fi
done

# --- Update changelog ---
echo "📝 Updating CHANGELOG.md..."
touch "$CHANGELOG_FILE"

# Get the "initial commit" hash
INITIAL_COMMIT=$(git rev-list --max-parents=0 HEAD)

# Only warn (don’t break) if not named "initial commit"
FIRST_MESSAGE=$(git log -1 --format=%s "$INITIAL_COMMIT")
if ! echo "$FIRST_MESSAGE" | grep -qi "initial commit"; then
  echo "${_YELLOW}⚠️ Warning:${_RESET} First commit is not named 'initial commit' — message: '$FIRST_MESSAGE'"
fi

# Get existing tags
EXISTING_TAGS=$(git tag --sort=-v:refname | grep -E '^v?[0-9]+\.[0-9]+\.[0-9]+$' | grep -v "^$TAG$" || true)

# Get last version tag or fallback to initial commit
LAST_TAG=$(echo "$EXISTING_TAGS" | head -n1)
if [[ -z "$LAST_TAG" ]]; then
  echo "🔍 No previous version tag found. Using initial commit."
  FROM_COMMIT="$INITIAL_COMMIT"
else
  echo "🔍 Last tag: $LAST_TAG"
  FROM_COMMIT="$LAST_TAG"
fi

# Get commits since last tag (exclude merges)
if [[ "$FROM_COMMIT" == "$INITIAL_COMMIT" ]]; then
  RAW_COMMITS=$(git log "$FROM_COMMIT" --pretty=format:"- %s" --no-merges || true)
else
  RAW_COMMITS=$(git log "$FROM_COMMIT"..HEAD --pretty=format:"- %s" --no-merges || true)
fi
echo "🔍 Raw commits:"
if [[ -z "$RAW_COMMITS" ]]; then
  echo "    [empty]"
else
  while IFS= read -r line; do
    echo "    $line"
  done <<< "$RAW_COMMITS"
fi

# Filter out commits with [skip changelog] (if any) and release commits from changelog
COMMITS=$(echo "$RAW_COMMITS" | grep -v '\[skip changelog\]' | grep -v 'chore(release)' || true)
echo "🔍 After filtering and excluding release commits:"
if [[ -z "$COMMITS" ]]; then
  echo "    [empty]"
else
  while IFS= read -r line; do
    echo "    $line"
  done <<< "$COMMITS"
fi

# Group commits by type
DOCS=$(echo "$COMMITS" | grep -Ei '^\s*-\s*docs(\([^)]+\))?:' || true)
FEATURES=$(echo "$COMMITS" | grep -Ei '^\s*-\s*feat(\([^)]+\))?:' || true)
FIXES=$(echo "$COMMITS" | grep -Ei '^\s*-\s*fix(\([^)]+\))?:' || true)
PERF=$(echo "$COMMITS" | grep -Ei '^\s*-\s*perf(\([^)]+\))?:' || true)
REFACTORS=$(echo "$COMMITS" | grep -Ei '^\s*-\s*refactor(\([^)]+\))?:' || true)
OTHERS=$(echo "$COMMITS" | grep -vE '^\s*-\s*(feat|fix|refactor|docs|perf)' || true)

GROUPED=""
echo -e "🧮 Grouped commits preview:"

if [[ -n "$DOCS" ]]; then
  echo -e "    📚 ${_BOLD}Documentation:${_RESET}"
  while IFS= read -r line; do
    echo "    $line"
  done <<< "$DOCS"
  GROUPED="${GROUPED}\n\n### 📚 Documentation\n$DOCS"
fi

if [[ -n "$FEATURES" ]]; then
  echo -e "    ✨ ${_BOLD}Features:${_RESET}"
  while IFS= read -r line; do
    echo "    $line"
  done <<< "$FEATURES"
  GROUPED="${GROUPED}\n\n### ✨ Features\n$FEATURES"
fi

if [[ -n "$FIXES" ]]; then
  echo -e "    🐛 ${_BOLD}Fixes:${_RESET}"
  while IFS= read -r line; do
    echo "    $line"
  done <<< "$FIXES"
  GROUPED="${GROUPED}\n\n### 🐛 Fixes\n$FIXES"
fi

if [[ -n "$PERF" ]]; then
  echo -e "    ⚡️ ${_BOLD}Performance Improvements:${_RESET}"
  while IFS= read -r line; do
    echo "    $line"
  done <<< "$PERF"
  GROUPED="${GROUPED}\n\n### ⚡️ Performance Improvements\n$PERF"
fi

if [[ -n "$REFACTORS" ]]; then
  echo -e "    🧠 ${_BOLD}Refactors:${_RESET}"
  while IFS= read -r line; do
    echo "    $line"
  done <<< "$REFACTORS"
  GROUPED="${GROUPED}\n\n### 🧠 Refactors\n$REFACTORS"
fi

# Special case: only initial commit
if [ "$(echo "$COMMITS" | wc -l)" -eq 1 ] && echo "$COMMITS" | grep -qE '^- initial commit$'; then
  echo -e "    🪄 ${_BOLD}Initial Commit only:${_RESET}"
  echo "    $COMMITS"
  GROUPED="$COMMITS"
else
  if [[ -n "$OTHERS" ]]; then
    echo -e "    🔧 ${_BOLD}Other Changes:${_RESET}"
    while IFS= read -r line; do
      echo "    $line"
    done <<< "$OTHERS"
    GROUPED="${GROUPED}\n\n### 🔧 Other Changes\n$OTHERS"
  fi
fi

# Fallback if no commits found
if [[ -z "$GROUPED" ]]; then
  GROUPED="- Add changes."
fi

# Clean up GROUPED for proper formatting
GROUPED="$(echo -e "$GROUPED" | sed -E '/^[[:space:]]*$/d')"

if [[ "$DRY_RUN" == true ]]; then
  echo -e "${_BOLD}🧪 Changelog Preview:${_RESET}"
  echo -e "    ## $TAG - $(date +%Y-%m-%d)\n"
  echo -e "$GROUPED" | sed 's/^/    /'
  echo -e "\n${_YELLOW}🧪 Dry run completed. No changes were made.${_RESET}"
  exit 0
else
  # Fallback if no meaningful commits found (prevent empty entry)
  if [[ "$GROUPED" == "- Add changes." ]]; then
    echo "${_YELLOW}⚠️ No valid commits found. Skipping changelog update.${_RESET}"
    exit 0
  fi

  # Create temporary body
  TMP_CHANGELOG_BODY=$(mktemp)

  # Extract version entries only (skip header)
  awk '
    BEGIN { in_body = 0 }
    /^## \[/ { in_body = 1 }
    {
      if (in_body) print
    }
  ' "$CHANGELOG_FILE" > "$TMP_CHANGELOG_BODY"

  # Write new changelog with header + new entry + previous entries
  {
    echo "# CHANGELOG"
    echo
    echo "We keep all the changes in this project documents in this file following [Keep a CHANGELOG](http://keepachangelog.com/en/0.3.0/)'s rules."
    echo "We always version our apps using [Semantic Versioning 2.0](http://semver.org/)."
    echo

    # Append new entry
    printf "## [%s](%s/releases/tag/v%s) - %s\n\n%s\n\n" \
      "$VERSION" "$REPOSITORY_URL" "$VERSION" "$(date +%Y-%m-%d)" "$GROUPED"

    # Append previous body
    cat "$TMP_CHANGELOG_BODY"
  } > "$CHANGELOG_FILE.tmp"

  # Replace CHANGELOG and delete temporary body
  mv "$CHANGELOG_FILE.tmp" "$CHANGELOG_FILE"
  rm "$TMP_CHANGELOG_BODY"

  echo "📥 Adding..."
  git add "$CHANGELOG_FILE"

  echo "💾 Committing..."
  if git log -1 --format=%s | grep -q "chore(release): $TAG"; then
    git commit --amend --no-edit
  else
    git commit -m "docs(changelog): add $TAG"
  fi
fi

# --- Update versions ---
if [[ "$DRY_RUN" == true ]]; then
  echo "🧪 Would update version to $VERSION in:"
  echo "    - $PACKAGE_JSON"
  echo "    - $HOMEBREW_FORMULA"
  echo "🧪 Would commit and tag: v$VERSION"
else
  echo "📦 Updating version in package.json..."
  safe_sed -E "s/\"version\": *\"[^\"]+\"/\"version\": \"$VERSION\"/" "$PACKAGE_JSON"

  echo "🍺 Updating version in Homebrew formula..."
  safe_sed "s|version \".*\"|version \"$VERSION\"|" "$HOMEBREW_FORMULA"

  echo "📥 Adding..."
  git add "$OUTPUT_DIR" "$PACKAGE_JSON" "$HOMEBREW_FORMULA"

  echo "💾 Committing..."
  git commit -m "chore(release): v$VERSION"

  echo "🏷️ Tagging..."
  git tag -a "$TAG" -m "Release version $VERSION"
fi

if [[ "$DRY_RUN" == true ]]; then
  echo "${_YELLOW}🧪 Dry run completed. No changes were made.${_RESET}"
else
  echo -n "${_BOLD}🚀 Push changes and tags to remote?${_RESET} [y/N] "
  read -r ANSWER
  if [[ "$ANSWER" =~ ^[Yy]$ ]]; then
    git push && git push --tags
  else
    echo "📌 ${_YELLOW}Skipped push. Remember to run: git push && git push --tags${_RESET}"
  fi
fi
