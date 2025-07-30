# Katana Metrics Tracker

A modular Go application for monitoring various metrics in the Katana rollup ecosystem. The application is designed to support multiple metrics scripts with shared infrastructure for Ethereum connections, L2 queries, and DataDog integration.

## Architecture

The application follows a modular structure:

```
finality-tracker/
├── main.go                    # Entry point for the application
├── common/                    # Shared functionality
│   ├── config.go             # Configuration management
│   ├── clients.go            # Client connections (Ethereum, L2, DataDog)
│   ├── constants.go          # Shared constants and event signatures
│   ├── types.go              # Shared data structures
│   └── README.md             # Common package documentation
├── scripts/                   # Individual metrics scripts
│   ├── finality-tracker/     # L1-L2 finality timing metrics
│   │   ├── finality_tracker.go
│   │   └── README.md
│   ├── balance-monitor/       # Vault balance monitoring
│   │   ├── balance_monitor.go
│   │   └── README.md
│   └── README.md             # Scripts documentation
└── README.md                 # This file
```

## Available Scripts

### Finality Tracker
Monitors the timing between L1 and L2 block finality by:
- Subscribing to `VerifyBatchesTrustedAggregator` events
- Calculating time deltas between L1 and L2 timestamps
- Sending metrics to DataDog for monitoring

### Balance Monitor
Monitors vault balances and tracks revenue generation by:
- Monitoring four vault addresses every 300 seconds
- Tracking current balances of BaseFeeVault, L1FeeVault, OperatorFeeVault, and SequencerFeeVault
- Calculating delta increases over 300-second intervals
- Sending both current balance and delta metrics to DataDog

## Prerequisites

- Go 1.24.4 or higher
- Linux system with systemd
- Ethereum RPC endpoint (Alchemy, Infura, etc.)
- Polygon zkEVM RPC endpoint
- Datadog-Agent installed and running with DogStatsD enabled

## Quick Start

### 1. Build the Application

```bash
# Clone the repository
git clone <repository-url>
cd finality-tracker

# Build the application
go build -o katana-metrics
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
# Ethereum RPC endpoint (websocket only)
ETH_RPC=https://eth-mainnet.g.alchemy.com/v2/your-api-key

# Polygon zkEVM proxy contract address
POLYGON_ZKEVM_PROXY_ADDR=0x5132A183E9F3CB7C848b0AAC5Ae0c4f0491B7aB2

# Rollup ID to monitor (20 for this case)
ROLLUP_ID=20

# Rollup RPC endpoint for L2 queries
ROLLUP_RPC=https://polygon-zkevm-mainnet.g.alchemy.com/v2/your-api-key
```

### 3. Test the Application

```bash
# Run the application locally to test
./katana-metrics
```

## Adding New Scripts

To add a new metrics script:

1. Create a new directory under `scripts/` with a descriptive name
2. Implement the script following the pattern in `scripts/README.md`
3. Update `main.go` to include the new script if needed

See `scripts/README.md` for detailed instructions on creating new scripts.

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
sudo mkdir -p /opt/katana-metrics

# Copy the application binary
sudo cp katana-metrics /opt/katana-metrics/

# Copy the environment file
sudo cp .env /opt/katana-metrics/

# Set proper permissions
sudo chmod +x /opt/katana-metrics/katana-metrics
```

#### 3. Install Systemd Service

```bash
# Copy the service file to systemd directory
sudo cp katana-metrics.service /etc/systemd/system/

# Reload systemd to recognize the new service
sudo systemctl daemon-reload

# Enable the service to start on boot
sudo systemctl enable katana-metrics

# Start the service
sudo systemctl start katana-metrics

# Check the service status
sudo systemctl status katana-metrics
```

## Service Management

### Check Service Status

```bash
sudo systemctl status katana-metrics
```

### View Logs

```bash
# View real-time logs
sudo journalctl -u katana-metrics -f

# View recent logs
sudo journalctl -u katana-metrics --since "1 hour ago"
```

### Stop/Start/Restart Service

```bash
# Stop the service
sudo systemctl stop katana-metrics

# Start the service
sudo systemctl start katana-metrics

# Restart the service
sudo systemctl restart katana-metrics
```

## Uninstallation

### Option 1: Automated Uninstallation

```bash
# Make the uninstallation script executable
chmod +x uninstall.sh

# Run the automated uninstallation
sudo ./uninstall.sh
```

### Option 2: Manual Uninstallation

```bash
# Stop and disable the service
sudo systemctl stop katana-metrics
sudo systemctl disable katana-metrics

# Remove the service file
sudo rm /etc/systemd/system/katana-metrics.service

# Reload systemd
sudo systemctl daemon-reload

# Remove the application directory
sudo rm -rf /opt/katana-metrics
```

## Metrics

The application sends the following metrics to DataDog:

### Finality Tracker Metrics

- **Metric Name**: `katana_finality_tracker.l1_l2_time_delta`
- **Type**: Gauge
- **Description**: Time difference between L1 timestamp and L2 block timestamp in seconds
- **Tags**:
  - `l2_block_number`: The L2 block number
  - `rollup_id`: The rollup ID being monitored

### Balance Monitor Metrics

#### Current Balance Metrics
- **Metric Name**: `katana_balance_monitor.basefee_vault_balance`
- **Metric Name**: `katana_balance_monitor.l1fee_vault_balance`
- **Metric Name**: `katana_balance_monitor.operator_fee_vault_balance`
- **Metric Name**: `katana_balance_monitor.sequencer_fee_vault_balance`
- **Type**: Gauge
- **Description**: Current balance of each vault in wei
- **Tags**:
  - `vault_type`: Type of vault (basefee, l1fee, operator_fee, sequencer_fee)
  - `vault_address`: The vault contract address
  - `rollup_id`: The rollup ID being monitored

#### Delta Metrics
- **Metric Name**: `katana_balance_monitor.basefee_vault_delta_300s`
- **Metric Name**: `katana_balance_monitor.l1fee_vault_delta_300s`
- **Metric Name**: `katana_balance_monitor.operator_fee_vault_delta_300s`
- **Metric Name**: `katana_balance_monitor.sequencer_fee_vault_delta_300s`
- **Type**: Gauge
- **Description**: Balance increase over the last 300 seconds in wei
- **Tags**:
  - `vault_type`: Type of vault (basefee, l1fee, operator_fee, sequencer_fee)
  - `vault_address`: The vault contract address
  - `rollup_id`: The rollup ID being monitored

## Development

### Project Structure

- `main.go`: Application entry point
- `common/`: Shared functionality for all scripts
- `scripts/`: Individual metrics scripts
- `install.sh`: Automated installation script
- `uninstall.sh`: Automated uninstallation script
- `katana-metrics.service`: Systemd service definition

### Adding New Scripts

1. Create a new directory under `scripts/`
2. Implement the script following the pattern in `scripts/README.md`
3. Update `main.go` to include the new script
4. Add documentation in the script's directory

### Building

```bash
# Build the application
go build -o katana-metrics

# Build for different architectures
GOOS=linux GOARCH=amd64 go build -o katana-metrics-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o katana-metrics-darwin-amd64
```

## Troubleshooting

### Common Issues

1. **Service fails to start**: Check the logs with `sudo journalctl -u katana-metrics -f`
2. **Connection errors**: Verify your RPC endpoints are correct and accessible
3. **DataDog metrics not appearing**: Ensure the DataDog agent is running and DogStatsD is enabled

### Logs

The application logs to systemd journal. View logs with:

```bash
# Real-time logs
sudo journalctl -u katana-metrics -f

# Recent logs
sudo journalctl -u katana-metrics --since "1 hour ago"

# All logs
sudo journalctl -u katana-metrics
```

## License

[Add your license information here]
