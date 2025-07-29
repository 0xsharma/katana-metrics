package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/joho/godotenv"
)

const (
	// Event signatures
	VerifyBatchesTrustedAggregatorSig = "0xd1ec3a1216f08b6eff72e169ceb548b782db18a6614852618d86bb19f3f9b0d3"
	OutputProposedSig                 = "0xa7aaf2512769da4e444e3de247be2564225c2e7a8f74cfe528e46e17d24868e2"
)

type Config struct {
	EthRPC                string
	PolygonZkEVMProxyAddr string
	RollupID              uint32
	RollupRPC             string
}

type FinalityTracker struct {
	config    Config
	ethClient *ethclient.Client
	l2Client  *ethclient.Client
	ddClient  *statsd.Client
}

type BlockResponse struct {
	Timestamp string `json:"timestamp"`
}

func loadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	rollupIDStr := os.Getenv("ROLLUP_ID")
	rollupID, err := strconv.ParseUint(rollupIDStr, 10, 32)
	if err != nil {
		log.Fatalf("Invalid ROLLUP_ID: %v", err)
	}

	return Config{
		EthRPC:                os.Getenv("ETH_RPC"),
		PolygonZkEVMProxyAddr: os.Getenv("POLYGON_ZKEVM_PROXY_ADDR"),
		RollupID:              uint32(rollupID),
		RollupRPC:             os.Getenv("ROLLUP_RPC"),
	}
}

func NewFinalityTracker(config Config) (*FinalityTracker, error) {
	// Convert HTTP RPC to WebSocket if needed
	ethRPC := config.EthRPC
	if !strings.HasPrefix(ethRPC, "ws://") && !strings.HasPrefix(ethRPC, "wss://") {
		// Convert HTTP to WebSocket
		ethRPC = strings.Replace(ethRPC, "http://", "ws://", 1)
		ethRPC = strings.Replace(ethRPC, "https://", "wss://", 1)
		log.Printf("Converting RPC endpoint to WebSocket: %s", ethRPC)
	}

	// Connect to Ethereum client via WebSocket
	ethClient, err := ethclient.Dial(ethRPC)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client via WebSocket: %v", err)
	}
	log.Printf("Successfully connected to Ethereum via WebSocket: %s", ethRPC)

	// Connect to L2 client (keep as HTTP for now since we only query it)
	l2Client, err := ethclient.Dial(config.RollupRPC)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to L2 client: %v", err)
	}

	// Initialize DataDog client
	ddClient, err := statsd.New("127.0.0.1:8125")
	if err != nil {
		return nil, fmt.Errorf("failed to create DataDog client: %v", err)
	}

	return &FinalityTracker{
		config:    config,
		ethClient: ethClient,
		l2Client:  l2Client,
		ddClient:  ddClient,
	}, nil
}

func (ft *FinalityTracker) Start(ctx context.Context) error {
	log.Println("Starting finality tracker...")

	for {
		if err := ft.runEventSubscription(ctx); err != nil {
			log.Printf("WebSocket connection error: %v", err)
			log.Println("Attempting to reconnect in 5 seconds...")

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
				continue
			}
		}
	}
}

func (ft *FinalityTracker) runEventSubscription(ctx context.Context) error {
	// Create filter for VerifyBatchesTrustedAggregator events
	contractAddr := common.HexToAddress(ft.config.PolygonZkEVMProxyAddr)
	eventSig := common.HexToHash(VerifyBatchesTrustedAggregatorSig)

	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr},
		Topics:    [][]common.Hash{{eventSig}},
	}

	// Subscribe to logs via WebSocket
	logs := make(chan types.Log)
	sub, err := ft.ethClient.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to logs: %v", err)
	}
	defer sub.Unsubscribe()

	log.Printf("Subscribed to VerifyBatchesTrustedAggregator events on contract %s via WebSocket", ft.config.PolygonZkEVMProxyAddr)

	// Health check ticker
	healthTicker := time.NewTicker(30 * time.Second)
	defer healthTicker.Stop()

	for {
		select {
		case err := <-sub.Err():
			return fmt.Errorf("WebSocket subscription error: %v", err)
		case vLog := <-logs:
			if err := ft.processVerifyBatchesLog(ctx, vLog); err != nil {
				log.Printf("Error processing log: %v", err)
			}
		case <-healthTicker.C:
			if err := ft.checkWebSocketHealth(ctx); err != nil {
				log.Printf("WebSocket health check failed: %v", err)
				return fmt.Errorf("WebSocket connection unhealthy: %v", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (ft *FinalityTracker) processVerifyBatchesLog(ctx context.Context, vLog types.Log) error {
	// Extract rollupID from the first indexed topic (topic[1])
	if len(vLog.Topics) < 2 {
		return fmt.Errorf("insufficient topics in log")
	}

	rollupID := new(big.Int).SetBytes(vLog.Topics[1].Bytes()).Uint64()
	log.Printf("Received VerifyBatchesTrustedAggregator event for rollupID: %d", rollupID)

	// Check if this is the rollup we're monitoring
	if uint32(rollupID) != ft.config.RollupID {
		log.Printf("Ignoring event for rollupID %d (monitoring %d)", rollupID, ft.config.RollupID)
		return nil
	}

	// Get the full transaction receipt to find OutputProposed events
	receipt, err := ft.ethClient.TransactionReceipt(ctx, vLog.TxHash)
	if err != nil {
		return fmt.Errorf("failed to get transaction receipt: %v", err)
	}

	// Find OutputProposed event in the transaction logs
	outputProposedSig := common.HexToHash(OutputProposedSig)
	for _, txLog := range receipt.Logs {
		if len(txLog.Topics) > 0 && txLog.Topics[0] == outputProposedSig {
			if err := ft.processOutputProposedLog(ctx, *txLog); err != nil {
				log.Printf("Error processing OutputProposed log: %v", err)
			}
		}
	}

	return nil
}

func (ft *FinalityTracker) processOutputProposedLog(ctx context.Context, vLog types.Log) error {
	if len(vLog.Topics) < 4 {
		return fmt.Errorf("insufficient topics in OutputProposed log")
	}

	// Extract L2BlockNumber from topic[3] and l1Timestamp from data
	l2BlockNumber := new(big.Int).SetBytes(vLog.Topics[3].Bytes())

	// l1Timestamp is in the data field (32 bytes)
	if len(vLog.Data) < 32 {
		return fmt.Errorf("insufficient data in OutputProposed log")
	}
	l1Timestamp := new(big.Int).SetBytes(vLog.Data[:32])

	log.Printf("Found OutputProposed: L2BlockNumber=%s, L1Timestamp=%s", l2BlockNumber.String(), l1Timestamp.String())

	// Query L2 block timestamp
	l2BlockTime, err := ft.getL2BlockTimestamp(ctx, l2BlockNumber)
	if err != nil {
		return fmt.Errorf("failed to get L2 block timestamp: %v", err)
	}

	// Calculate delta (L1 timestamp - L2 block timestamp)
	delta := l1Timestamp.Int64() - l2BlockTime

	log.Printf("L2BlockNumber: %s, L1Timestamp: %d, L2BlockTime: %d, Delta: %d seconds",
		l2BlockNumber.String(), l1Timestamp.Int64(), l2BlockTime, delta)

	// Send metric to DataDog
	if err := ft.sendMetricToDataDog(l2BlockNumber.String(), delta); err != nil {
		return fmt.Errorf("failed to send metric to DataDog: %v", err)
	}

	return nil
}

func (ft *FinalityTracker) getL2BlockTimestamp(ctx context.Context, blockNumber *big.Int) (int64, error) {
	// Use RPC client to get block details
	rpcClient, err := rpc.Dial(ft.config.RollupRPC)
	if err != nil {
		return 0, fmt.Errorf("failed to dial L2 RPC: %v", err)
	}
	defer rpcClient.Close()

	var result BlockResponse
	err = rpcClient.CallContext(ctx, &result, "eth_getBlockByNumber", fmt.Sprintf("0x%x", blockNumber), false)
	if err != nil {
		return 0, fmt.Errorf("failed to get block: %v", err)
	}

	// Parse hex timestamp
	timestamp, err := strconv.ParseInt(result.Timestamp[2:], 16, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse timestamp: %v", err)
	}

	return timestamp, nil
}

func (ft *FinalityTracker) sendMetricToDataDog(l2BlockNumber string, delta int64) error {
	tags := []string{
		fmt.Sprintf("l2_block_number:%s", l2BlockNumber),
		fmt.Sprintf("rollup_id:%d", ft.config.RollupID),
	}

	// Send the delta as a gauge metric
	err := ft.ddClient.Gauge("katana_finality_tracker.l1_l2_time_delta", float64(delta), tags, 1)
	if err != nil {
		return fmt.Errorf("failed to send gauge metric: %v", err)
	}

	log.Printf("Sent metric to DataDog: l1_l2_time_delta=%d, tags=%v", delta, tags)
	return nil
}

func (ft *FinalityTracker) checkWebSocketHealth(ctx context.Context) error {
	// Try to get the latest block number to check if WebSocket connection is alive
	_, err := ft.ethClient.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("WebSocket health check failed: %v", err)
	}
	return nil
}

func (ft *FinalityTracker) Close() {
	if ft.ethClient != nil {
		ft.ethClient.Close()
	}
	if ft.l2Client != nil {
		ft.l2Client.Close()
	}
	if ft.ddClient != nil {
		ft.ddClient.Close()
	}
}

func main() {
	// Configure logging to show timestamps and ensure immediate output
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)

	config := loadConfig()

	// Validate required environment variables
	if config.EthRPC == "" {
		log.Fatal("ETH_RPC environment variable is required")
	}
	if config.PolygonZkEVMProxyAddr == "" {
		log.Fatal("POLYGON_ZKEVM_PROXY_ADDR environment variable is required")
	}
	if config.RollupRPC == "" {
		log.Fatal("ROLLUP_RPC environment variable is required")
	}

	tracker, err := NewFinalityTracker(config)
	if err != nil {
		log.Fatalf("Failed to create finality tracker: %v", err)
	}
	defer tracker.Close()

	ctx := context.Background()

	log.Printf("Monitoring rollup ID: %d", config.RollupID)
	log.Printf("Contract address: %s", config.PolygonZkEVMProxyAddr)

	if err := tracker.Start(ctx); err != nil {
		log.Fatalf("Tracker error: %v", err)
	}
}
