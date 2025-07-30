#!/bin/bash

# Katana Metrics Tracker Uninstallation Script
# This script removes the metrics service and all related files

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

print_status "Starting Katana Metrics uninstallation..."

# Stop and disable the service
print_status "Stopping service..."
systemctl stop katana-metrics 2>/dev/null || true

print_status "Disabling service..."
systemctl disable katana-metrics 2>/dev/null || true

# Remove the service file
print_status "Removing service file..."
rm -f /etc/systemd/system/katana-metrics.service

# Reload systemd
print_status "Reloading systemd..."
systemctl daemon-reload

# Remove the application directory
print_status "Removing application directory..."
rm -rf /opt/katana-metrics


print_status "Uninstallation completed successfully!"
print_status "All files and service have been removed." 
