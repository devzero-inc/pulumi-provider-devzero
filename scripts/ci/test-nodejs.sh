#!/usr/bin/env bash
# scripts/ci/test-nodejs.sh
# Build the NodeJS SDK, pack it, and verify a downstream consumer can tsc --noEmit against it.
# Env vars:
#   MATRIX_TYPESCRIPT (required)  e.g. "5.9"
#   FRESH_INSTALL     (optional)  "true" drops package-lock.json before install
set -euo pipefail

TS_VERSION="${MATRIX_TYPESCRIPT:?MATRIX_TYPESCRIPT required}"
FRESH_INSTALL="${FRESH_INSTALL:-false}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SDK_DIR="$REPO_ROOT/sdk/nodejs"

echo "==> NodeJS SDK test: typescript=$TS_VERSION fresh_install=$FRESH_INSTALL"

cd "$SDK_DIR"

# package.json has "${VERSION}" placeholder which npm rejects; stamp a test version.
jq '.version = "0.0.0-test"' package.json > package.json.tmp
mv package.json.tmp package.json

if [[ "$FRESH_INSTALL" == "true" ]]; then
    rm -f package-lock.json
fi

npm install
npm install --save-dev "typescript@$TS_VERSION"

echo "==> Building SDK with TS $TS_VERSION"
npm run build

echo "==> Packing tarball"
TARBALL="$(npm pack --silent)"
TARBALL_PATH="$SDK_DIR/$TARBALL"

echo "==> Scaffolding consumer project"
CONSUMER_DIR="$(mktemp -d)"
trap 'rm -rf "$CONSUMER_DIR"; git -C "$REPO_ROOT" checkout -- sdk/nodejs/package.json sdk/nodejs/package-lock.json 2>/dev/null || true' EXIT
cd "$CONSUMER_DIR"

cat > package.json <<'PKG'
{
  "name": "consumer-test",
  "version": "1.0.0",
  "private": true
}
PKG

cat > tsconfig.json <<'TSC'
{
  "compilerOptions": {
    "strict": true,
    "target": "ES2020",
    "module": "commonjs",
    "moduleResolution": "node",
    "esModuleInterop": true,
    "noEmit": true
  },
  "files": ["index.ts"]
}
TSC

cat > index.ts <<'TS'
import * as devzero from "@devzero/pulumi-devzero";
import { Provider } from "@devzero/pulumi-devzero";

type Cluster = devzero.resources.Cluster;

declare const _ns: typeof devzero;
declare const _p: Provider;
declare const _c: Cluster;
void _ns; void _p; void _c;
TS

echo "==> Installing SDK tarball + typescript@$TS_VERSION"
npm install "$TARBALL_PATH" "typescript@$TS_VERSION"

echo "==> Running consumer tsc --noEmit"
./node_modules/.bin/tsc --noEmit

echo "==> NodeJS SDK test OK (TS $TS_VERSION)"
