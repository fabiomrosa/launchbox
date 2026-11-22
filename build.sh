#!/bin/bash
set -e

# --- Load Envs ---
source "$(dirname "$0")/env.sh"

# --- Determine version to inject ---
VERSION="${VERSION:-dev}"
LD_FLAGS="-X main.Version=${VERSION} -s -w"

echo "📦 Injecting version: $VERSION"

# --- Clear binaries ---
[ -d "$OUTPUT_DIR" ] && rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# --- Build binaries ---
(
  cd "$GO_WORKDIR"

  if [ ! -f "$GO_ENTRYPOINT/main.go" ]; then
    echo "${_RED}❌ Error:${_RESET} Entry point ./$GO_ENTRYPOINT/main.go not found"
    exit 1
  fi

  for PLATFORM in "${PLATFORMS[@]}"; do
    IFS="/" read -r GOOS GOARCH <<< "$PLATFORM"
    OUTPUT_NAME="${APP_NAME}-${GOOS}-${GOARCH}"

    echo "🔨 Building for $GOOS/$GOARCH → $OUTPUT_DIR/$OUTPUT_NAME"
    GOOS="$GOOS" GOARCH="$GOARCH" go build \
      -tags release \
      -trimpath \
      -ldflags="$LD_FLAGS" \
      -o "../$OUTPUT_DIR/$OUTPUT_NAME" "./$GO_ENTRYPOINT"
  done
)
