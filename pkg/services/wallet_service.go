package services

import (
	"fmt"
	"ordm-main/pkg/wallet"
)

type WalletService struct {
	walletManager *wallet.WalletManager
}

func NewWalletService(walletManager *wallet.WalletManager) *WalletService {
	return &WalletService{
		walletManager: walletManager,
	}
}

func (ws *WalletService) CreateWallet() (*wallet.BIP39Wallet, error) {
	return ws.walletManager.CreateWallet("default", "")
}

func (ws *WalletService) GetBalance(userID string) (map[string]interface{}, error) {
	// Buscar wallet por ID
	wallet, exists := ws.walletManager.GetWallet(userID)
	if !exists {
		return nil, fmt.Errorf("wallet não encontrada")
	}

	// Pegar primeira conta
	accounts := wallet.GetAccounts()
	if len(accounts) == 0 {
		return nil, fmt.Errorf("wallet sem contas")
	}

	account := accounts[0]
	return map[string]interface{}{
		"balance": account.Balance,
		"address": account.Address,
		"staked":  0, // Não implementado ainda
	}, nil
}

func (ws *WalletService) StakeTokens(userID string, amount int64) error {
	// Implementação futura
	return fmt.Errorf("stake não implementado ainda")
}

func (ws *WalletService) GetWallet(userID string) (*wallet.BIP39Wallet, error) {
	wallet, exists := ws.walletManager.GetWallet(userID)
	if !exists {
		return nil, fmt.Errorf("wallet não encontrada")
	}
	return wallet, nil
}
