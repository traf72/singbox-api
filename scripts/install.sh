#!/usr/bin/env bash
set -euo pipefail
umask 077

OWNER="traf72"
REPO="singbox-api"
BRANCH="main"
APP_NAME="singbox-api"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

usage() {
  echo "Usage:" >&2
  echo "  $0 -i <service-user> [additional installer args...]   install / upgrade" >&2
  echo "  $0 -u                                             uninstall" >&2
  exit 2
}

[[ $# -ge 1 ]] || usage
action="$1"
shift

case "$action" in
  -i)
    [[ $# -ge 1 ]] || usage
    SERVICE_USER="$1"
    shift
    SCRIPT_PATH="scripts/install.sh"
    ;;
  -u)
    [[ $# -eq 0 ]] || usage
    SCRIPT_PATH="scripts/uninstall.sh"
    ;;
  *) usage ;;
esac

if [[ "$EUID" -ne 0 ]]; then
  echo "Error: run this wrapper with sudo (or as root)." >&2
  exit 1
fi

# Make Go visible even when launched via sudo.
export PATH="/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:${PATH:-}"

# Token priority:
# 1. GITHUB_TOKEN explicitly passed through sudo
# 2. Original sudo user's token (NOT root's token)
# 3. Current HOME's token as a fallback
PAT="${GITHUB_TOKEN:-}"
PAT_SOURCE="GITHUB_TOKEN"

if [[ -z "$PAT" && -n "${SUDO_USER:-}" && "$SUDO_USER" != "root" ]]; then
  SUDO_HOME="$(getent passwd "$SUDO_USER" 2>/dev/null | cut -d: -f6 || true)"
  if [[ -n "$SUDO_HOME" && -r "$SUDO_HOME/github-singbox-repos-pat" ]]; then
    IFS= read -r PAT < "$SUDO_HOME/github-singbox-repos-pat" || true
    PAT_SOURCE="$SUDO_HOME/github-singbox-repos-pat"
  fi
fi

if [[ -z "$PAT" && -r "${HOME:-}/github-singbox-repos-pat" ]]; then
  IFS= read -r PAT < "${HOME}/github-singbox-repos-pat" || true
  PAT_SOURCE="${HOME}/github-singbox-repos-pat"
fi

PAT="${PAT%$'\r'}"  # Allow a PAT file saved with CRLF endings.
if [[ -z "$PAT" ]]; then
  echo "Error: GitHub PAT missing. Set GITHUB_TOKEN or create ~/github-singbox-repos-pat." >&2
  exit 1
fi

echo "[run-installation] Using GitHub token from: ${PAT_SOURCE}"

# Validate local arguments before downloading anything.
if [[ "$action" == "-i" ]]; then
  PASSWD_ENTRY="$(getent passwd "$SERVICE_USER" || true)"
  if [[ -z "$PASSWD_ENTRY" ]]; then
    echo "Error: service user '$SERVICE_USER' does not exist." >&2
    exit 1
  fi
  SERVICE_HOME="$(printf '%s\n' "$PASSWD_ENTRY" | cut -d: -f6)"
  if [[ -z "$SERVICE_HOME" ]]; then
    echo "Error: cannot determine home directory for '$SERVICE_USER'." >&2
    exit 1
  fi
  if [[ ! -f "$SCRIPT_DIR/env_stub" ]]; then
    echo "Error: missing $SCRIPT_DIR/env_stub" >&2
    exit 1
  fi
fi

# GitHub REST Contents API supports authenticated raw downloads for private repos.
URL="https://api.github.com/repos/${OWNER}/${REPO}/contents/${SCRIPT_PATH}?ref=${BRANCH}"
TMP_SCRIPT="$(mktemp)"
trap 'rm -f "$TMP_SCRIPT"' EXIT

echo "[run-installation] Fetching ${SCRIPT_PATH} from ${OWNER}/${REPO}@${BRANCH} ..."
if ! HTTP_STATUS="$(curl -4 -sS -L \
    --connect-timeout 15 --max-time 120 --retry 2 \
    -o "$TMP_SCRIPT" -w '%{http_code}' \
    -H "Authorization: Bearer $PAT" \
    -H 'Accept: application/vnd.github.raw+json' \
    "$URL")"; then
  echo "Error: could not download ${SCRIPT_PATH} (network/TLS failure)." >&2
  exit 1
fi

if [[ "$HTTP_STATUS" != "200" ]]; then
  echo "Error: GitHub returned HTTP ${HTTP_STATUS} for ${OWNER}/${REPO}/${SCRIPT_PATH} (branch ${BRANCH})." >&2
  case "$HTTP_STATUS" in
    401) echo "The PAT was rejected by GitHub." >&2 ;;
    403) echo "Check GitHub permissions, rate limiting, or organization access restrictions." >&2 ;;
    404) echo "Check repository access (fine-grained PAT: Contents/Read), owner, branch and file path." >&2 ;;
  esac
  exit 1
fi

if [[ ! -s "$TMP_SCRIPT" ]] || ! bash -n "$TMP_SCRIPT"; then
  echo "Error: downloaded file is empty or is not a valid Bash script." >&2
  exit 1
fi

# Prepare .env only after the download succeeds; never overwrite on upgrade.
if [[ "$action" == "-i" ]]; then
  INSTALL_DIR="${SERVICE_HOME}/${APP_NAME}"
  TARGET_ENV="${INSTALL_DIR}/bin/.env"
  SERVICE_GROUP="$(id -gn "$SERVICE_USER")"

  if [[ ! -d "$INSTALL_DIR" ]]; then
    install -d -m 750 -o "$SERVICE_USER" -g "$SERVICE_GROUP" "$INSTALL_DIR"
  fi
  if [[ ! -d "$INSTALL_DIR/bin" ]]; then
    install -d -m 750 -o "$SERVICE_USER" -g "$SERVICE_GROUP" "$INSTALL_DIR/bin"
  fi

  if [[ -e "$TARGET_ENV" || -L "$TARGET_ENV" ]]; then
    echo "[run-installation] Preserving existing $TARGET_ENV"
  else
    echo "[run-installation] Creating $TARGET_ENV"
    install -m 600 -o "$SERVICE_USER" -g "$SERVICE_GROUP" \
      "$SCRIPT_DIR/env_stub" "$TARGET_ENV"
  fi

  # The child script can reuse the same PAT for git clone or API requests.
  GITHUB_TOKEN="$PAT" bash "$TMP_SCRIPT" "$SERVICE_USER" "$@"
else
  GITHUB_TOKEN="$PAT" bash "$TMP_SCRIPT"
fi
