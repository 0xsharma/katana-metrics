#!/bin/bash

# Katana Finality Tracker Installation Script
# This script automates the installation of the finality tracker as a systemd service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   print_error "This script must be run as root (use sudo)"
   exit 1
fi

print_status "Starting Katana Finality Tracker installation..."

# Check if Go binary exists
if [[ ! -f "./katana-finality-tracker" ]]; then
    print_error "katana-finality-tracker binary not found in current directory"
    print_status "Please build the application first: go build -o katana-finality-tracker"
    exit 1
fi

# Check if .env file exists
if [[ ! -f "./.env" ]]; then
    print_warning ".env file not found. Please create it from .env.example"
    print_status "cp .env.example .env"
    print_status "Then edit .env with your actual values"
    exit 1
fi

# Check if service file exists
if [[ ! -f "./katana-finality-tracker.service" ]]; then
    print_error "katana-finality-tracker.service file not found"
    exit 1
fi


print_status "Creating application directory..."
# Create application directory
mkdir -p /opt/katana-finality-tracker

print_status "Copying application files..."
# Copy application binary
cp katana-finality-tracker /opt/katana-finality-tracker/
cp .env /opt/katana-finality-tracker/

print_status "Setting proper permissions..."
# Set permissions (no ownership change needed)
chmod +x /opt/katana-finality-tracker/katana-finality-tracker
chmod 644 /opt/katana-finality-tracker/.env

print_status "Installing systemd service..."
# Copy service file
cp katana-finality-tracker.service /etc/systemd/system/

print_status "Reloading systemd..."
# Reload systemd
systemctl daemon-reload

print_status "Enabling service..."
# Enable service
systemctl enable katana-finality-tracker

print_status "Starting service..."
# Start service
systemctl start katana-finality-tracker

# Wait a moment for service to start
sleep 2

# Check service status
if systemctl is-active --quiet katana-finality-tracker; then
    print_status "Service started successfully!"
    print_status "Service status:"
    systemctl status katana-finality-tracker --no-pager -l
else
    print_error "Service failed to start. Check logs with:"
    print_status "journalctl -u katana-finality-tracker -f"
    exit 1
fi

print_status "Installation completed successfully!"
echo ""
print_status "Useful commands:"
echo "  View logs: sudo journalctl -u katana-finality-tracker -f"
echo "  Check status: sudo systemctl status katana-finality-tracker"
echo "  Restart service: sudo systemctl restart katana-finality-tracker"
echo "  Stop service: sudo systemctl stop katana-finality-tracker"
echo ""
print_status "The service will automatically start on boot and restart if it crashes." 
