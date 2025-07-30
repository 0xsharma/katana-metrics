package common

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	EthRPC                string
	PolygonZkEVMProxyAddr string
	RollupID              uint32
	RollupRPC             string
}

func LoadConfig() Config {
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

func ValidateConfig(config Config) error {
	if config.EthRPC == "" {
		return fmt.Errorf("ETH_RPC environment variable is required")
	}
	if config.PolygonZkEVMProxyAddr == "" {
		return fmt.Errorf("POLYGON_ZKEVM_PROXY_ADDR environment variable is required")
	}
	if config.RollupRPC == "" {
		return fmt.Errorf("ROLLUP_RPC environment variable is required")
	}
	return nil
}
