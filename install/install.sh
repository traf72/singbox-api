#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Check if a user parameter is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <user>"
    exit 1
fi

# Variables
SERVICE_USER="$1"
REPO_URL="git@github.com:traf72/singbox-api.git"
BUILD_DIR="$HOME/singbox-api/src"
INSTALL_DIR="$HOME/singbox-api/bin"
SERVICE_NAME="singbox-api"
EXECUTABLE_NAME="singbox-api"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

# Functions
function log {
    echo "[$(date +"%Y-%m-%d %H:%M:%S")] $*"
}

# Step 1: Check if systemd service exists
if systemctl list-units --full --all | grep -Fq "${SERVICE_NAME}.service"; then
    log "Service already exists. Proceeding with update..."

    # Stop the service if running
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        log "Stopping the current service..."
        sudo systemctl stop "$SERVICE_NAME"
    fi

    # Update the code
    if [ -d "$BUILD_DIR" ]; then
        log "Pulling the latest code..."
        cd "$BUILD_DIR"
        git reset --hard HEAD
        git clean -fd
        git pull
    else
        log "Cloning the repository..."
        git clone "$REPO_URL" "$BUILD_DIR"
        cd "$BUILD_DIR"
    fi

    # Build the new version
    log "Building the new version..."
    go build -o "$EXECUTABLE_NAME" ./cmd/api

    # Replace the existing binary
    log "Replacing the existing binary..."
    mv "$EXECUTABLE_NAME" "$INSTALL_DIR"

    # Restart the service
    log "Restarting the service..."
    sudo systemctl start "$SERVICE_NAME"

    log "Update complete! The web server is running the latest version."
else
    log "No existing service found. Proceeding with fresh installation..."

    # Install prerequisites
    log "Installing prerequisites..."
    sudo apt-get update
    sudo apt-get install -y git

    # Clone the repository
    log "Cloning the repository..."
    rm -rf "$BUILD_DIR"
    git clone "$REPO_URL" "$BUILD_DIR"

    # Build the binary
    log "Building the project..."
    cd "$BUILD_DIR"
    go build -o "$EXECUTABLE_NAME" ./cmd/api

    # Install the binary
    log "Installing the binary..."
    mkdir -p "$INSTALL_DIR"
    mv "$EXECUTABLE_NAME" "$INSTALL_DIR"

    # Create the systemd service file
    log "Creating the systemd service file..."
    sudo bash -c "cat > $SERVICE_FILE" <<EOL
[Unit]
Description=Singbox API Service
After=network.target

[Service]
ExecStart=${INSTALL_DIR}/${EXECUTABLE_NAME}
WorkingDirectory=${INSTALL_DIR}
Restart=always
User=${SERVICE_USER}
Group=${SERVICE_USER}

[Install]
WantedBy=multi-user.target
EOL

    # Reload systemd and start the service
    log "Reloading systemd and enabling the service..."
    sudo systemctl daemon-reload
    sudo systemctl enable "$SERVICE_NAME"
    sudo systemctl start "$SERVICE_NAME"

    log "Installation complete! The web server is now running as a systemd service."
fi
