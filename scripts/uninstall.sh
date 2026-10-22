#!/usr/bin/env bash
set -euo pipefail

SERVICE_NAME="singbox-api"
UNIT="${SERVICE_NAME}.service"
UNIT_FILE="/etc/systemd/system/${UNIT}"

log() {
  printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$*"
}

usage() {
  echo "Usage: $0 [service-user]   (default: root)" >&2
  exit 2
}

[[ $# -le 1 ]] || usage
SERVICE_USER="${1:-root}"

if [[ "$EUID" -ne 0 ]]; then
  echo "Error: run the uninstaller as root (e.g., through sudo)." >&2
  exit 1
fi

# Resolve the installation directory from the account used at install time.
PASSWD_ENTRY="$(getent passwd "$SERVICE_USER" || true)"
if [[ -z "$PASSWD_ENTRY" ]]; then
  echo "Error: service user '$SERVICE_USER' does not exist." >&2
  exit 1
fi
SERVICE_HOME="$(printf '%s\n' "$PASSWD_ENTRY" | cut -d: -f6)"
if [[ "$SERVICE_HOME" != /* || "$SERVICE_HOME" == "/" ]]; then
  echo "Error: unsafe home directory for '$SERVICE_USER': '$SERVICE_HOME'." >&2
  exit 1
fi

APP_DIR="${SERVICE_HOME%/}/${SERVICE_NAME}"
if [[ -L "$APP_DIR" ]]; then
  echo "Error: refusing to delete through a symlink: $APP_DIR" >&2
  exit 1
fi
if [[ -e "$APP_DIR" && ! -d "$APP_DIR" ]]; then
  echo "Error: expected an installation directory, found something else: $APP_DIR" >&2
  exit 1
fi

log "Uninstalling ${SERVICE_NAME} for service user '${SERVICE_USER}'"
log "Installation directory: ${APP_DIR}"

# Stop first; if the stop fails, do not delete a potentially running binary.
if systemctl is-active --quiet "$UNIT"; then
  log "Stopping ${UNIT}"
  systemctl stop "$UNIT"
fi

# Disabling doesn't stop a unit, so it is handled separately.
if systemctl is-enabled --quiet "$UNIT"; then
  log "Disabling ${UNIT}"
  systemctl disable "$UNIT"
fi

if systemctl is-active --quiet "$UNIT"; then
  echo "Error: ${UNIT} is still running. Refusing to delete application files." >&2
  exit 1
fi

if [[ -e "$UNIT_FILE" || -L "$UNIT_FILE" ]]; then
  log "Removing systemd unit file: $UNIT_FILE"
  rm -f -- "$UNIT_FILE"
fi

# Refresh even if the unit file was already missing.
systemctl daemon-reload
systemctl reset-failed "$UNIT" >/dev/null 2>&1 || true

# Remove only the paths the installer owns; leave any unrelated files alone.
for dir in "${APP_DIR}/bin" "${APP_DIR}/src"; do
  if [[ -e "$dir" || -L "$dir" ]]; then
    log "Removing $dir"
    rm -rf -- "$dir"
  fi
done

if [[ -d "$APP_DIR" ]]; then
  if rmdir -- "$APP_DIR" 2>/dev/null; then
    log "Removed empty application directory: $APP_DIR"
  else
    log "Keeping application directory because other files remain: $APP_DIR"
  fi
fi

log "Uninstallation finished."
