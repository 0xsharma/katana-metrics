# Katana Finality Tracker

A Go application that monitors Ethereum events for finality tracking and sends metrics to DataDog. The application subscribes to `VerifyBatchesTrustedAggregator` events and calculates time deltas between L1 and L2 block timestamps.

## Prerequisites

- Go 1.24.4 or higher
- Linux system with systemd
- DataDog account with API key
- Ethereum RPC endpoint (Alchemy, Infura, etc.)
- Polygon zkEVM RPC endpoint

## Quick Start

### 1. Build the Application

```bash
# Clone the repository
git clone <repository-url>
cd finality-tracker

# Build the application
go build -o katana-finality-tracker
```

### 2. Create Environment Configuration

```bash
# Copy the example environment file
cp .env.example .env

# Edit the environment file with your actual values
nano .env
```

Required environment variables:
```bash
# Ethereum RPC endpoint (will be converted to WebSocket automatically)
ETH_RPC=https://eth-mainnet.g.alchemy.com/v2/your-api-key

# Polygon zkEVM proxy contract address
POLYGON_ZKEVM_PROXY_ADDR=0x5132A183E9F3CB7C848b0AAC5Ae0c4f0491B7aB2

# Rollup ID to monitor (20 for this case)
ROLLUP_ID=20

# Rollup RPC endpoint for L2 queries
ROLLUP_RPC=https://polygon-zkevm-mainnet.g.alchemy.com/v2/your-api-key

# DataDog API key for sending metrics
DATADOG_API_KEY=your-datadog-api-key

# DataDog API URL (optional, defaults to https://api.datadoghq.com)
DATADOG_API_URL=https://api.datadoghq.com
```

### 3. Test the Application

```bash
# Run the application locally to test
./katana-finality-tracker
```

## Systemd Service Installation

### Option 1: Automated Installation (Recommended)

```bash
# Make the installation script executable
chmod +x install.sh

# Run the automated installation
sudo ./install.sh
```

The installation script will:
- Set up the application directory
- Install the systemd service
- Start the service automatically
- Verify the installation

### Option 2: Manual Installation

#### 1. Create Application Directory

```bash
# Create the application directory
sudo mkdir -p /opt/katana-finality-tracker

# Copy the application binary
sudo cp katana-finality-tracker /opt/katana-finality-tracker/

# Copy the environment file
sudo cp .env /opt/katana-finality-tracker/

# Set proper permissions
sudo chmod +x /opt/katana-finality-tracker/katana-finality-tracker
```

#### 3. Install Systemd Service

```bash
# Copy the service file to systemd directory
sudo cp katana-finality-tracker.service /etc/systemd/system/

# Reload systemd to recognize the new service
sudo systemctl daemon-reload

# Enable the service to start on boot
sudo systemctl enable katana-finality-tracker

# Start the service
sudo systemctl start katana-finality-tracker
```

#### 4. Verify Service Status

```bash
# Check service status
sudo systemctl status katana-finality-tracker

# View logs
sudo journalctl -u katana-finality-tracker -f

# Check if the service is running
sudo systemctl is-active katana-finality-tracker
```

## Service Management

### Start/Stop/Restart Service

```bash
# Start the service
sudo systemctl start katana-finality-tracker

# Stop the service
sudo systemctl stop katana-finality-tracker

# Restart the service
sudo systemctl restart katana-finality-tracker

# Reload configuration (if you changed .env)
sudo systemctl reload katana-finality-tracker
```

### View Logs

```bash
# View real-time logs
sudo journalctl -u katana-finality-tracker -f

# View logs from the last hour
sudo journalctl -u katana-finality-tracker --since "1 hour ago"

# View logs with timestamps
sudo journalctl -u katana-finality-tracker -o short-iso

# View error logs only
sudo journalctl -u katana-finality-tracker -p err
```

### Service Configuration

The service file includes:

- **Automatic Restart**: Service restarts automatically if it crashes
- **Environment Variables**: Loads from `/opt/katana-finality-tracker/.env`
- **Security**: Runs with limited privileges and system protection
- **Resource Limits**: Configured file descriptor and process limits
- **Logging**: All output goes to systemd journal

## Monitoring and Troubleshooting

### Check Service Health

```bash
# Check if the service is running
sudo systemctl is-active katana-finality-tracker

# Check service status with details
sudo systemctl status katana-finality-tracker

# View recent logs
sudo journalctl -u katana-finality-tracker -n 50
```

### Common Issues

1. **Service won't start**: Check logs with `journalctl -u katana-finality-tracker`
2. **Environment variables not loaded**: Verify `.env` file exists and has correct permissions
3. **Permission denied**: Ensure the `finality-tracker` user owns the application directory
4. **Network connectivity**: Check if the service can reach the RPC endpoints

### Debug Mode

To run the application in debug mode:

```bash
# Stop the service
sudo systemctl stop katana-finality-tracker

# Run manually with debug output
sudo -u finality-tracker /opt/katana-finality-tracker/katana-finality-tracker
```

## DataDog Integration

The application sends the following metrics to DataDog:

- **Metric Name**: `katana_finality_tracker.l1_l2_time_delta`
- **Type**: Gauge
- **Tags**: 
  - `l2_block_number`: The L2 block number
  - `rollup_id`: The rollup ID being monitored
- **Value**: Time delta between L1 and L2 timestamps in seconds

### DataDog Dashboard Setup

1. Go to your DataDog dashboard
2. Create a new dashboard
3. Add a widget for the metric `katana_finality_tracker.l1_l2_time_delta`
4. Set up alerts for unusual finality delays

## Service Configuration

- The service runs with automatic restart on failure
- Environment variables are loaded from a secure file
- All logs are captured in systemd journal
- Resource limits are configured to prevent abuse

## Updating the Application

```bash
# Stop the service
sudo systemctl stop katana-finality-tracker

# Backup the current binary
sudo cp /opt/katana-finality-tracker/katana-finality-tracker /opt/katana-finality-tracker/katana-finality-tracker.backup

# Copy the new binary
sudo cp katana-finality-tracker /opt/katana-finality-tracker/

# Set proper permissions
sudo chown finality-tracker:finality-tracker /opt/katana-finality-tracker/katana-finality-tracker
sudo chmod +x /opt/katana-finality-tracker/katana-finality-tracker

# Start the service
sudo systemctl start katana-finality-tracker

# Verify it's running
sudo systemctl status katana-finality-tracker
```

## Uninstallation

### Option 1: Automated Uninstallation (Recommended)

```bash
# Make the uninstall script executable
chmod +x uninstall.sh

# Run the automated uninstallation
sudo ./uninstall.sh
```

The uninstall script will:
- Stop and disable the service
- Remove all service files
- Clean up the application directory

### Option 2: Manual Uninstallation

```bash
# Stop and disable the service
sudo systemctl stop katana-finality-tracker
sudo systemctl disable katana-finality-tracker

# Remove the service file
sudo rm /etc/systemd/system/katana-finality-tracker.service

# Reload systemd
sudo systemctl daemon-reload

# Remove the application directory
sudo rm -rf /opt/katana-finality-tracker
```

## Support

For issues and questions:
1. Check the logs: `sudo journalctl -u katana-finality-tracker`
2. Verify environment configuration
3. Test network connectivity to RPC endpoints
4. Ensure DataDog API key is valid 
