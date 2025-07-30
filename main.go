package main

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/0xsharma/katana-finality-tracker/common"
	balancemonitor "github.com/0xsharma/katana-finality-tracker/scripts/balance-monitor"
	finalitytracker "github.com/0xsharma/katana-finality-tracker/scripts/finality-tracker"
)

func main() {
	// Configure logging to show timestamps and ensure immediate output
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stdout)

	config := common.LoadConfig()

	// Validate required environment variables
	if err := common.ValidateConfig(config); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Initialize clients
	clients, err := common.NewClients(config)
	if err != nil {
		log.Fatalf("Failed to create clients: %v", err)
	}
	defer clients.Close()

	ctx := context.Background()

	log.Printf("Monitoring rollup ID: %d", config.RollupID)
	log.Printf("Contract address: %s", config.PolygonZkEVMProxyAddr)

	// Create scripts
	finalityTracker := finalitytracker.NewFinalityTracker(config, clients)
	balanceMonitor := balancemonitor.NewBalanceMonitor(config, clients)

	// Run both scripts concurrently
	var wg sync.WaitGroup
	wg.Add(2)

	// Start finality tracker
	go func() {
		defer wg.Done()
		if err := finalityTracker.Start(ctx); err != nil {
			log.Fatalf("Finality tracker error: %v", err)
		}
	}()

	// Start balance monitor
	go func() {
		defer wg.Done()
		if err := balanceMonitor.Start(ctx); err != nil {
			log.Fatalf("Balance monitor error: %v", err)
		}
	}()

	// Wait for both scripts to complete (they should run indefinitely)
	wg.Wait()
}
