#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Variables
SERVICE_NAME="singbox-api"
BUILD_DIR="$HOME/singbox-api/src"
INSTALL_DIR="$HOME/singbox-api/bin"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

# Functions
function log {
    echo "[$(date +"%Y-%m-%d %H:%M:%S")] $*"
}

# Step 1: Stop and Disable the Service
log "Stopping and disabling the service..."
if systemctl list-units --full --all | grep -Fq "${SERVICE_NAME}.service"; then
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        sudo systemctl stop "$SERVICE_NAME"
        log "Service stopped."
    fi
    sudo systemctl disable "$SERVICE_NAME"
    log "Service disabled."
else
    log "Service not found. Skipping stop and disable steps."
fi

# Step 2: Remove the Systemd Service File
if [ -f "$SERVICE_FILE" ]; then
    log "Removing systemd service file..."
    sudo rm "$SERVICE_FILE"
	sudo systemctl reset-failed "${SERVICE_NAME}.service"
    sudo systemctl daemon-reload
    log "Service file removed and systemd reloaded."
else
    log "Service file not found. Skipping removal."
fi

# Step 3: Remove the Binary and Source Directories
if [ -d "$INSTALL_DIR" ]; then
    log "Removing binary directory: $INSTALL_DIR"
    rm -rf "$INSTALL_DIR"
else
    log "Binary directory not found. Skipping removal."
fi

if [ -d "$BUILD_DIR" ]; then
    log "Removing source directory: $BUILD_DIR"
    rm -rf "$BUILD_DIR"
else
    log "Source directory not found. Skipping removal."
fi

log "Uninstallation complete! Singbox API has been fully removed."