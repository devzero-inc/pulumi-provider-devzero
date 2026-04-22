#!/usr/bin/env bash
# scripts/ci/test-go.sh
# Build the Go SDK, then scaffold a consumer module that imports it via replace directive.
# Env vars:
#   FRESH_INSTALL (optional)  "true" drops go.sum + runs go mod tidy before build
set -euo pipefail

FRESH_INSTALL="${FRESH_INSTALL:-false}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SDK_DIR="$REPO_ROOT/sdk/go/devzero"

echo "==> Go SDK test: fresh_install=$FRESH_INSTALL"

cd "$SDK_DIR"

if [[ "$FRESH_INSTALL" == "true" ]]; then
    rm -f go.sum
    go mod tidy
fi

echo "==> Building SDK (go build ./...)"
go build ./...

MODULE_PATH="$(awk '/^module /{print $2}' go.mod)"
echo "==> SDK module path: $MODULE_PATH"

echo "==> Scaffolding consumer module"
CONSUMER_DIR="$(mktemp -d)"
trap 'rm -rf "$CONSUMER_DIR"; git -C "$REPO_ROOT" checkout -- sdk/go/devzero/go.sum 2>/dev/null || true' EXIT
cd "$CONSUMER_DIR"

cat > go.mod <<EOF
module consumer-test

go 1.22

require $MODULE_PATH v0.0.0-00010101000000-000000000000

replace $MODULE_PATH => $SDK_DIR
EOF

cat > main.go <<EOF
package main

import (
	"fmt"

	_ "$MODULE_PATH"
)

func main() {
	fmt.Println("ok")
}
EOF

echo "==> go mod tidy + go build"
go mod tidy
go build ./...

echo "==> Go SDK test OK"
