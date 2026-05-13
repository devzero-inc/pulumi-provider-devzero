#!/usr/bin/env bash
set -euo pipefail

# ── Environment ──────────────────────────────────────────────────────────────
export PULUMI_CONFIG_PASSPHRASE=1
export PULUMI_BACKEND_URL=file://~

# DevZero credentials – must be supplied before running the script.
# Export them in your shell or pass them inline:
# export  DEVZERO_TEAM_ID=team-xxx 
# export DEVZERO_TOKEN=dzu-xxx ./scripts/test-examples.sh
# export  DEVZERO_URL is optional – defaults to https://dakr.devzero.io
#   ./scripts/test-packages.sh
: "${DEVZERO_TEAM_ID:?DEVZERO_TEAM_ID is required}"
: "${DEVZERO_TOKEN:?DEVZERO_TOKEN is required}"
#better to test against dev
DEVZERO_URL="${DEVZERO_URL:-https://dakr.devzero.dev}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# ── Colors ────────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'
RED='\033[0;31m'
BOLD='\033[1m'
RESET='\033[0m'

# ── Helpers ───────────────────────────────────────────────────────────────────
log() {
  echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

log_pass() { log "${GREEN}PASS${RESET} $*"; }
log_fail() { log "${RED}FAIL${RESET} $*"; }

# ── Per-language test function ────────────────────────────────────────────────
# Returns 0 on success, 1 on failure.
run_language_test() {
  local lang="$1"
  local stack="test-${lang}"
  local dir="${REPO_ROOT}/examples/${lang}"

  log "──────────────────────────────────────────"
  log "Testing ${BOLD}${lang}${RESET} (stack: ${stack})"
  log "──────────────────────────────────────────"

  cd "${dir}"

  # Language-specific pre-build step
  if [[ "${lang}" == "go" ]]; then
    log "[${lang}] Building binary..."
    if ! go build -o devzero-example . ; then
      log_fail "[${lang}] go build failed"
      return 1
    fi
  elif [[ "${lang}" == "python" ]]; then
    log "[${lang}] Installing Python dependencies..."
    if ! python3 -m pip install -r requirements.txt -q ; then
      log_fail "[${lang}] pip install failed"
      return 1
    fi
  elif [[ "${lang}" == "typescript" ]]; then
    log "[${lang}] Installing Node dependencies..."
    if ! npm install ; then
      log_fail "[${lang}] npm install failed"
      return 1
    fi
    log "[${lang}] Building TypeScript..."
    if ! npm run build ; then
      log_fail "[${lang}] npm run build failed"
      return 1
    fi
  fi

  # 1. Stack init
  log "[${lang}] Initializing stack ${stack}..."
  if ! pulumi stack init "${stack}" ; then
    log_fail "[${lang}] stack init failed"
    return 1
  fi

  # Set required config after stack is created
  log "[${lang}] Setting devzero config..."
  pulumi config set devzero:teamId "${DEVZERO_TEAM_ID}"
  pulumi config set devzero:token  "${DEVZERO_TOKEN}" --secret
  pulumi config set devzero:url    "${DEVZERO_URL}"

  # 2. pulumi up
  log "[${lang}] Running pulumi up..."
  if ! pulumi up --yes --skip-preview ; then
    log_fail "[${lang}] pulumi up failed – cleaning up..."
    pulumi destroy --yes 2>/dev/null || true
    pulumi stack rm "${stack}" --yes 2>/dev/null || true
    return 1
  fi

  # 3. pulumi destroy
  log "[${lang}] Running pulumi destroy..."
  if ! pulumi destroy --yes ; then
    log_fail "[${lang}] pulumi destroy failed – cleaning up..."
    pulumi stack rm "${stack}" --yes 2>/dev/null || true
    return 1
  fi

  # 4. Stack removal
  log "[${lang}] Removing stack ${stack}..."
  if ! pulumi stack rm "${stack}" --yes ; then
    log_fail "[${lang}] stack rm failed"
    return 1
  fi

  log_pass "[${lang}] All steps completed successfully."
  return 0
}

# ── Main ──────────────────────────────────────────────────────────────────────
LANGUAGES=("typescript" "go" "python")
RESULTS=()

for lang in "${LANGUAGES[@]}"; do
  if run_language_test "${lang}"; then
    RESULTS+=("PASSED")
  else
    RESULTS+=("FAILED")
  fi
done

# ── Summary table ─────────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}Results:${RESET}"
OVERALL=0
for i in "${!LANGUAGES[@]}"; do
  lang="${LANGUAGES[$i]}"
  status="${RESULTS[$i]}"
  if [[ "${status}" == "PASSED" ]]; then
    printf "  %-12s ${GREEN}%s${RESET}\n" "${lang}" "${status}"
  else
    printf "  %-12s ${RED}%s${RESET}\n" "${lang}" "${status}"
    OVERALL=1
  fi
done
echo ""

exit "${OVERALL}"
