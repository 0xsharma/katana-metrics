package balancemonitor

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	commonpkg "github.com/0xsharma/katana-finality-tracker/common"
)

const (
	// Vault addresses
	BaseFeeVaultAddr      = "0x4200000000000000000000000000000000000019"
	L1FeeVaultAddr        = "0x420000000000000000000000000000000000001a"
	OperatorFeeVaultAddr  = "0x420000000000000000000000000000000000001B"
	SequencerFeeVaultAddr = "0x4200000000000000000000000000000000000011"

	// Monitoring interval
	MonitorInterval = 300 * time.Second
)

type VaultBalance struct {
	Address  string
	Current  *big.Int
	Previous *big.Int
	Delta    *big.Int
}

type BalanceMonitor struct {
	config  commonpkg.Config
	clients *commonpkg.Clients
	vaults  map[string]*VaultBalance
}

func NewBalanceMonitor(config commonpkg.Config, clients *commonpkg.Clients) *BalanceMonitor {
	return &BalanceMonitor{
		config:  config,
		clients: clients,
		vaults: map[string]*VaultBalance{
			BaseFeeVaultAddr: {
				Address:  BaseFeeVaultAddr,
				Current:  big.NewInt(0),
				Previous: big.NewInt(0),
				Delta:    big.NewInt(0),
			},
			L1FeeVaultAddr: {
				Address:  L1FeeVaultAddr,
				Current:  big.NewInt(0),
				Previous: big.NewInt(0),
				Delta:    big.NewInt(0),
			},
			OperatorFeeVaultAddr: {
				Address:  OperatorFeeVaultAddr,
				Current:  big.NewInt(0),
				Previous: big.NewInt(0),
				Delta:    big.NewInt(0),
			},
			SequencerFeeVaultAddr: {
				Address:  SequencerFeeVaultAddr,
				Current:  big.NewInt(0),
				Previous: big.NewInt(0),
				Delta:    big.NewInt(0),
			},
		},
	}
}

func (bm *BalanceMonitor) Start(ctx context.Context) error {
	log.Println("Starting balance monitor...")

	// Initial balance check
	if err := bm.updateBalances(ctx); err != nil {
		log.Printf("Error in initial balance check: %v", err)
	}

	// Set up ticker for periodic monitoring
	ticker := time.NewTicker(MonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := bm.monitorBalances(ctx); err != nil {
				log.Printf("Error monitoring balances: %v", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (bm *BalanceMonitor) monitorBalances(ctx context.Context) error {
	log.Println("Monitoring vault balances...")

	// Update balances
	if err := bm.updateBalances(ctx); err != nil {
		return fmt.Errorf("failed to update balances: %v", err)
	}

	// Calculate deltas and send metrics
	for _, vault := range bm.vaults {
		if err := bm.calculateDeltaAndSendMetrics(vault); err != nil {
			log.Printf("Error processing vault %s: %v", vault.Address, err)
		}
	}

	return nil
}

func (bm *BalanceMonitor) updateBalances(ctx context.Context) error {
	for _, vault := range bm.vaults {
		// Move current balance to previous
		vault.Previous.Set(vault.Current)

		// Get current balance
		balance, err := bm.getBalance(ctx, vault.Address)
		if err != nil {
			return fmt.Errorf("failed to get balance for %s: %v", vault.Address, err)
		}

		vault.Current.Set(balance)
		log.Printf("Vault %s: Previous=%s, Current=%s",
			vault.Address, vault.Previous.String(), vault.Current.String())
	}

	return nil
}

func (bm *BalanceMonitor) getBalance(ctx context.Context, address string) (*big.Int, error) {
	addr := common.HexToAddress(address)
	balance, err := bm.clients.L2Client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance for %s: %v", address, err)
	}
	return balance, nil
}

func (bm *BalanceMonitor) calculateDeltaAndSendMetrics(vault *VaultBalance) error {
	// Calculate delta (current - previous)
	vault.Delta.Sub(vault.Current, vault.Previous)

	// Send current balance metrics
	if err := bm.sendCurrentBalanceMetric(vault); err != nil {
		return fmt.Errorf("failed to send current balance metric: %v", err)
	}

	// Send delta metrics
	if err := bm.sendDeltaMetric(vault); err != nil {
		return fmt.Errorf("failed to send delta metric: %v", err)
	}

	log.Printf("Vault %s: Delta=%s (in 300s)", vault.Address, vault.Delta.String())
	return nil
}

func (bm *BalanceMonitor) sendCurrentBalanceMetric(vault *VaultBalance) error {
	var metricName string
	var vaultType string

	switch vault.Address {
	case BaseFeeVaultAddr:
		metricName = "katana_balance_monitor.basefee_vault_balance"
		vaultType = "basefee"
	case L1FeeVaultAddr:
		metricName = "katana_balance_monitor.l1fee_vault_balance"
		vaultType = "l1fee"
	case OperatorFeeVaultAddr:
		metricName = "katana_balance_monitor.operator_fee_vault_balance"
		vaultType = "operator_fee"
	case SequencerFeeVaultAddr:
		metricName = "katana_balance_monitor.sequencer_fee_vault_balance"
		vaultType = "sequencer_fee"
	default:
		return fmt.Errorf("unknown vault address: %s", vault.Address)
	}

	tags := []string{
		fmt.Sprintf("vault_type:%s", vaultType),
		fmt.Sprintf("vault_address:%s", vault.Address),
		fmt.Sprintf("rollup_id:%d", bm.config.RollupID),
	}

	// Convert balance to float64 (in wei)
	balanceFloat := new(big.Float).SetInt(vault.Current)
	balanceFloat64, _ := balanceFloat.Float64()

	err := bm.clients.DDClient.Gauge(metricName, balanceFloat64, tags, 1)
	if err != nil {
		return fmt.Errorf("failed to send gauge metric: %v", err)
	}

	log.Printf("Sent metric to DataDog: %s=%f, tags=%v", metricName, balanceFloat64, tags)
	return nil
}

func (bm *BalanceMonitor) sendDeltaMetric(vault *VaultBalance) error {
	var metricName string
	var vaultType string

	switch vault.Address {
	case BaseFeeVaultAddr:
		metricName = "katana_balance_monitor.basefee_vault_delta_300s"
		vaultType = "basefee"
	case L1FeeVaultAddr:
		metricName = "katana_balance_monitor.l1fee_vault_delta_300s"
		vaultType = "l1fee"
	case OperatorFeeVaultAddr:
		metricName = "katana_balance_monitor.operator_fee_vault_delta_300s"
		vaultType = "operator_fee"
	case SequencerFeeVaultAddr:
		metricName = "katana_balance_monitor.sequencer_fee_vault_delta_300s"
		vaultType = "sequencer_fee"
	default:
		return fmt.Errorf("unknown vault address: %s", vault.Address)
	}

	tags := []string{
		fmt.Sprintf("vault_type:%s", vaultType),
		fmt.Sprintf("vault_address:%s", vault.Address),
		fmt.Sprintf("rollup_id:%d", bm.config.RollupID),
	}

	// Convert delta to float64 (in wei)
	deltaFloat := new(big.Float).SetInt(vault.Delta)
	deltaFloat64, _ := deltaFloat.Float64()

	err := bm.clients.DDClient.Gauge(metricName, deltaFloat64, tags, 1)
	if err != nil {
		return fmt.Errorf("failed to send gauge metric: %v", err)
	}

	log.Printf("Sent metric to DataDog: %s=%f, tags=%v", metricName, deltaFloat64, tags)
	return nil
}
