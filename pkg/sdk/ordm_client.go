package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ORDMClient é o cliente SDK para interagir com a ORDM Blockchain
type ORDMClient struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// NewORDMClient cria um novo cliente ORDM
func NewORDMClient(baseURL string) *ORDMClient {
	return &ORDMClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewORDMClientWithAPIKey cria um novo cliente ORDM com API key
func NewORDMClientWithAPIKey(baseURL, apiKey string) *ORDMClient {
	return &ORDMClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
	}
}

// ===== MÉTODOS DA BLOCKCHAIN =====

// GetBlockchainInfo retorna informações gerais da blockchain
func (c *ORDMClient) GetBlockchainInfo() (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/v1/blockchain/info", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return result, nil
}

// GetBlockchainStatus retorna o status atual da blockchain
func (c *ORDMClient) GetBlockchainStatus() (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/v1/blockchain/status", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return result, nil
}

// ===== MÉTODOS DE TRANSAÇÕES =====

// SendTransaction envia uma nova transação
func (c *ORDMClient) SendTransaction(from, to string, amount, fee int64, data, signature string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"from":      from,
		"to":        to,
		"amount":    amount,
		"fee":       fee,
		"data":      data,
		"signature": signature,
	}

	resp, err := c.makeRequest("POST", "/api/v1/transactions/send", payload)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return result, nil
}

// GetPendingTransactions retorna transações pendentes
func (c *ORDMClient) GetPendingTransactions(limit int) (map[string]interface{}, error) {
	url := fmt.Sprintf("/api/v1/transactions/pending?limit=%d", limit)
	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return result, nil
}

// ===== MÉTODOS DO MEMPOOL =====

// GetMempoolStatus retorna status do mempool
func (c *ORDMClient) GetMempoolStatus() (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/v1/mempool/status", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return result, nil
}

// GetMempoolTransactions retorna transações do mempool
func (c *ORDMClient) GetMempoolTransactions(limit int) (map[string]interface{}, error) {
	url := fmt.Sprintf("/api/v1/mempool/transactions?limit=%d", limit)
	resp, err := c.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return result, nil
}

// ===== MÉTODOS DE CONSENSO =====

// GetConsensusStatus retorna status do consenso
func (c *ORDMClient) GetConsensusStatus() (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/v1/consensus/status", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return result, nil
}

// GetForks retorna forks conhecidos
func (c *ORDMClient) GetForks() (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/v1/consensus/forks", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return result, nil
}

// ===== MÉTODOS DE MINERAÇÃO =====

// GetMiningStatus retorna o status da mineração
func (c *ORDMClient) GetMiningStatus() (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/v1/mining/status", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return result, nil
}

// StartMining inicia mineração
func (c *ORDMClient) StartMining() (map[string]interface{}, error) {
	resp, err := c.makeRequest("POST", "/api/v1/mining/start", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return result, nil
}

// StopMining para mineração
func (c *ORDMClient) StopMining() (map[string]interface{}, error) {
	resp, err := c.makeRequest("POST", "/api/v1/mining/stop", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return result, nil
}

// ===== MÉTODOS DE ESTATÍSTICAS =====

// GetStats retorna estatísticas gerais
func (c *ORDMClient) GetStats() (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/v1/stats", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return result, nil
}

// GetNetworkStats retorna estatísticas da rede
func (c *ORDMClient) GetNetworkStats() (map[string]interface{}, error) {
	resp, err := c.makeRequest("GET", "/api/v1/stats/network", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return result, nil
}

// ===== MÉTODOS DE WALLETS =====

// CreateWallet cria uma nova wallet
func (c *ORDMClient) CreateWallet() (map[string]interface{}, error) {
	resp, err := c.makeRequest("POST", "/api/v1/wallets/create", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return result, nil
}

// ===== MÉTODOS AUXILIARES =====

// makeRequest faz uma requisição HTTP para a API
func (c *ORDMClient) makeRequest(method, endpoint string, payload interface{}) ([]byte, error) {
	url := c.baseURL + endpoint

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar payload: %v", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}

	// Adicionar headers
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	// Fazer requisição
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição HTTP: %v", err)
	}
	defer resp.Body.Close()

	// Ler resposta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	// Verificar status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("erro HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// ===== MÉTODOS DE UTILIDADE =====

// IsConnected verifica se o cliente está conectado à API
func (c *ORDMClient) IsConnected() bool {
	_, err := c.GetBlockchainInfo()
	return err == nil
}

// GetAPIVersion retorna a versão da API
func (c *ORDMClient) GetAPIVersion() (string, error) {
	info, err := c.GetBlockchainInfo()
	if err != nil {
		return "", err
	}

	if data, ok := info["data"].(map[string]interface{}); ok {
		if version, ok := data["version"].(string); ok {
			return version, nil
		}
	}

	return "", fmt.Errorf("versão não encontrada na resposta")
}

// Ping faz um ping para verificar conectividade
func (c *ORDMClient) Ping() (time.Duration, error) {
	start := time.Now()

	_, err := c.GetBlockchainInfo()
	if err != nil {
		return 0, err
	}

	return time.Since(start), nil
}
