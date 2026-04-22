#!/usr/bin/env bash
# scripts/ci/test-python.sh
# Build the Python SDK wheel, install in a fresh venv, and import-smoke-test it.
# Env vars:
#   FRESH_INSTALL (optional)  no-op for Python (pip always resolves fresh); wired for API consistency
set -euo pipefail

FRESH_INSTALL="${FRESH_INSTALL:-false}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SDK_DIR="$REPO_ROOT/sdk/python"

echo "==> Python SDK test: fresh_install=$FRESH_INSTALL (no-op for Python)"

# Prefer `python` (set by CI's setup-python), fall back to `python3`.
PYTHON="$(command -v python || command -v python3)"

cd "$SDK_DIR"

DIST_DIR="$(mktemp -d)"
VENV_DIR="$(mktemp -d)/venv"
trap 'rm -rf "$DIST_DIR" "$(dirname "$VENV_DIR")"' EXIT

echo "==> Building wheel"
"$PYTHON" -m pip install --quiet build
"$PYTHON" -m build --wheel --outdir "$DIST_DIR"

WHEEL="$(ls "$DIST_DIR"/*.whl 2>/dev/null | head -n1)"
[[ -n "$WHEEL" ]] || { echo "no wheel produced"; exit 1; }

echo "==> Creating fresh venv"
"$PYTHON" -m venv "$VENV_DIR"
# shellcheck disable=SC1091
source "$VENV_DIR/bin/activate"

echo "==> Installing wheel + pulumi runtime"
python -m pip install --quiet "$WHEEL" pulumi

echo "==> Import smoke test"
python - <<'PY'
import pulumi_devzero
assert hasattr(pulumi_devzero, "Provider"), "Provider missing from pulumi_devzero"
# Cluster lives in the resources submodule (not re-exported at top level)
assert hasattr(pulumi_devzero.resources, "Cluster"), "Cluster missing from pulumi_devzero.resources"
from pulumi_devzero import config  # exercise a submodule import
print("ok")
PY

deactivate
echo "==> Python SDK test OK"
