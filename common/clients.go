package common

import (
	"fmt"
	"log"
	"strings"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Clients struct {
	EthClient *ethclient.Client
	L2Client  *ethclient.Client
	DDClient  *statsd.Client
}

func NewClients(config Config) (*Clients, error) {
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

	return &Clients{
		EthClient: ethClient,
		L2Client:  l2Client,
		DDClient:  ddClient,
	}, nil
}

func (c *Clients) Close() {
	if c.EthClient != nil {
		c.EthClient.Close()
	}
	if c.L2Client != nil {
		c.L2Client.Close()
	}
	if c.DDClient != nil {
		c.DDClient.Close()
	}
}
