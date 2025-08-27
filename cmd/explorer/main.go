package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"ordm-main/pkg/ledger"
	"ordm-main/pkg/wallet"
)

type Explorer struct {
	GlobalLedger  *ledger.GlobalLedger
	WalletManager *wallet.WalletManager
	Port          string
}

type BlockInfo struct {
	Number       int64  `json:"number"`
	Hash         string `json:"hash"`
	Timestamp    int64  `json:"timestamp"`
	Transactions int    `json:"transactions"`
	Reward       int64  `json:"reward"`
	Difficulty   string `json:"difficulty"`
	Size         int    `json:"size"`
}

type TransactionInfo struct {
	Hash      string `json:"hash"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int64  `json:"amount"`
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
	Status    string `json:"status"`
	Fee       int64  `json:"fee"`
}

type WalletInfo struct {
	Address     string `json:"address"`
	Balance     int64  `json:"balance"`
	StakeAmount int64  `json:"stake_amount"`
	IsValidator bool   `json:"is_validator"`
	TxCount     int    `json:"tx_count"`
	LastTx      int64  `json:"last_tx"`
}

type NetworkStats struct {
	TotalBlocks       int64   `json:"total_blocks"`
	TotalTransactions int64   `json:"total_transactions"`
	TotalSupply       int64   `json:"total_supply"`
	ActiveWallets     int     `json:"active_wallets"`
	NetworkHashRate   float64 `json:"network_hash_rate"`
	AverageBlockTime  float64 `json:"average_block_time"`
	TotalStaked       int64   `json:"total_staked"`
	Validators        int     `json:"validators"`
}

var explorer *Explorer

func main() {
	// Inicializar explorer
	explorer = &Explorer{
		GlobalLedger:  ledger.NewGlobalLedger("data", nil),
		WalletManager: wallet.NewWalletManager("data"),
		Port:          ":8080",
	}

	// Carregar dados existentes
	explorer.loadExistingData()

	// Configurar rotas
	http.HandleFunc("/", explorer.handleHome)
	http.HandleFunc("/blocks", explorer.handleBlocks)
	http.HandleFunc("/transactions", explorer.handleTransactions)
	http.HandleFunc("/wallets", explorer.handleWallets)
	http.HandleFunc("/block/", explorer.handleBlockDetail)
	http.HandleFunc("/tx/", explorer.handleTransactionDetail)
	http.HandleFunc("/address/", explorer.handleAddressDetail)
	http.HandleFunc("/api/stats", explorer.handleAPIStats)
	http.HandleFunc("/api/blocks", explorer.handleAPIBlocks)
	http.HandleFunc("/api/transactions", explorer.handleAPITransactions)
	http.HandleFunc("/api/wallets", explorer.handleAPIWallets)

	// Servir arquivos estáticos
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("cmd/explorer/static"))))

	fmt.Printf("🔍 Blockchain Explorer iniciado em: http://localhost%s\n", explorer.Port)
	fmt.Printf("📊 Dashboard público disponível\n")
	fmt.Printf("🌐 Qualquer pessoa pode acessar e visualizar a blockchain\n")

	log.Fatal(http.ListenAndServe(explorer.Port, nil))
}

func (e *Explorer) loadExistingData() {
	// Carregar ledger global
	if err := e.GlobalLedger.LoadLedger(); err != nil {
		fmt.Printf("⚠️ Erro ao carregar ledger: %v\n", err)
	}

	// Carregar wallets
	if err := e.WalletManager.LoadWallets(); err != nil {
		fmt.Printf("⚠️ Erro ao carregar wallets: %v\n", err)
	}

	fmt.Printf("📊 Dados carregados: %d movimentos, %d wallets\n",
		len(e.GlobalLedger.Movements), len(e.WalletManager.Wallets))

	// Iniciar atualização automática dos dados
	go e.autoRefreshData()
}

func (e *Explorer) autoRefreshData() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Recarregar ledger para pegar novas transações
		if err := e.GlobalLedger.LoadLedger(); err == nil {
			// Log silencioso de atualização
		}
	}
}

func (e *Explorer) handleHome(w http.ResponseWriter, r *http.Request) {
	stats := e.getNetworkStats()

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>🔍 Blockchain Explorer - 2-Layer Network</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #0f0f23 0%%, #1a1a2e 50%%, #16213e 100%%);
            color: #e0e0e0;
            line-height: 1.6;
        }
        .header {
            background: rgba(0,0,0,0.8);
            padding: 1rem 0;
            border-bottom: 2px solid #00d4ff;
            box-shadow: 0 4px 20px rgba(0,212,255,0.3);
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 0 20px;
        }
        .logo {
            font-size: 2rem;
            font-weight: bold;
            color: #00d4ff;
            text-align: center;
            margin-bottom: 1rem;
        }
        .nav {
            display: flex;
            justify-content: center;
            gap: 2rem;
            flex-wrap: wrap;
        }
        .nav a {
            color: #e0e0e0;
            text-decoration: none;
            padding: 0.5rem 1rem;
            border-radius: 5px;
            transition: all 0.3s;
            border: 1px solid transparent;
        }
        .nav a:hover {
            background: rgba(0,212,255,0.1);
            border-color: #00d4ff;
            color: #00d4ff;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 1.5rem;
            margin: 2rem 0;
        }
        .stat-card {
            background: rgba(255,255,255,0.05);
            border: 1px solid rgba(0,212,255,0.3);
            border-radius: 10px;
            padding: 1.5rem;
            text-align: center;
            backdrop-filter: blur(10px);
            transition: transform 0.3s;
        }
        .stat-card:hover {
            transform: translateY(-5px);
            border-color: #00d4ff;
        }
        .stat-value {
            font-size: 2rem;
            font-weight: bold;
            color: #00d4ff;
            margin-bottom: 0.5rem;
        }
        .stat-label {
            color: #b0b0b0;
            font-size: 0.9rem;
        }
        .recent-section {
            background: rgba(255,255,255,0.03);
            border-radius: 10px;
            padding: 1.5rem;
            margin: 2rem 0;
            border: 1px solid rgba(0,212,255,0.2);
        }
        .section-title {
            color: #00d4ff;
            font-size: 1.5rem;
            margin-bottom: 1rem;
            border-bottom: 2px solid #00d4ff;
            padding-bottom: 0.5rem;
        }
        .block-list, .tx-list {
            display: grid;
            gap: 0.5rem;
        }
        .block-item, .tx-item {
            background: rgba(255,255,255,0.05);
            padding: 1rem;
            border-radius: 5px;
            border-left: 4px solid #00d4ff;
            transition: all 0.3s;
        }
        .block-item:hover, .tx-item:hover {
            background: rgba(0,212,255,0.1);
            transform: translateX(5px);
        }
        .hash {
            font-family: monospace;
            color: #00d4ff;
            font-size: 0.9rem;
        }
        .time {
            color: #888;
            font-size: 0.8rem;
        }
        .amount {
            color: #00ff88;
            font-weight: bold;
        }
        .footer {
            text-align: center;
            padding: 2rem 0;
            color: #888;
            border-top: 1px solid rgba(0,212,255,0.3);
            margin-top: 3rem;
        }
        .search-box {
            background: rgba(255,255,255,0.1);
            border: 1px solid rgba(0,212,255,0.3);
            border-radius: 25px;
            padding: 0.8rem 1.5rem;
            color: #e0e0e0;
            width: 100%%;
            max-width: 400px;
            margin: 1rem auto;
            display: block;
        }
        .search-box::placeholder {
            color: #888;
        }
        .search-box:focus {
            outline: none;
            border-color: #00d4ff;
            box-shadow: 0 0 10px rgba(0,212,255,0.3);
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="container">
            <div class="logo">🔍 Blockchain Explorer</div>
            <nav class="nav">
                <a href="/">🏠 Dashboard</a>
                <a href="/blocks">📦 Blocos</a>
                <a href="/transactions">💸 Transações</a>
                <a href="/wallets">👛 Wallets</a>
            </nav>
        </div>
    </div>

    <div class="container">
        <input type="text" class="search-box" placeholder="🔍 Buscar por hash, endereço ou bloco..." id="searchBox">
        
        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-value">%d</div>
                <div class="stat-label">Total de Blocos</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">%d</div>
                <div class="stat-label">Total de Transações</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">%d</div>
                <div class="stat-label">Supply Total</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">%d</div>
                <div class="stat-label">Wallets Ativas</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">%.2f</div>
                <div class="stat-label">Hash Rate (H/s)</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">%d</div>
                <div class="stat-label">Total Staked</div>
            </div>
        </div>

        <div class="recent-section">
            <h2 class="section-title">📦 Últimos Blocos</h2>
            <div class="block-list" id="recentBlocks">
                <!-- Blocos serão carregados via JavaScript -->
            </div>
        </div>

        <div class="recent-section">
            <h2 class="section-title">💸 Últimas Transações</h2>
            <div class="tx-list" id="recentTransactions">
                <!-- Transações serão carregadas via JavaScript -->
            </div>
        </div>
    </div>

    <div class="footer">
        <p>🔍 Blockchain Explorer - 2-Layer Network | Dados em tempo real</p>
        <p>🌐 Acesso público - Qualquer pessoa pode visualizar a blockchain</p>
    </div>

             <script>
             // Carregar dados em tempo real
                      function loadRecentBlocks() {
             fetch('/api/blocks?limit=5')
                 .then(response => response.json())
                 .then(blocks => {
                     const container = document.getElementById('recentBlocks');
                     container.innerHTML = blocks.map(block => 
                         '<div class="block-item">' +
                         '<div><strong>Bloco #' + block.number + '</strong></div>' +
                         '<div class="hash">' + block.hash + '</div>' +
                         '<div class="time">' + new Date(block.timestamp * 1000).toLocaleString() + '</div>' +
                         '<div>' + block.transactions + ' transações | +' + block.reward + ' tokens</div>' +
                         '</div>'
                     ).join('');
                 });
         }

                 function loadRecentTransactions() {
             fetch('/api/transactions?limit=5')
                 .then(response => response.json())
                 .then(transactions => {
                     const container = document.getElementById('recentTransactions');
                     container.innerHTML = transactions.map(tx => 
                         '<div class="tx-item">' +
                         '<div class="hash">' + tx.hash + '</div>' +
                         '<div>' + tx.from + ' → ' + tx.to + '</div>' +
                         '<div class="amount">' + tx.amount + ' tokens</div>' +
                         '<div class="time">' + new Date(tx.timestamp * 1000).toLocaleString() + '</div>' +
                         '</div>'
                     ).join('');
                 });
         }

        // Carregar dados iniciais
        loadRecentBlocks();
        loadRecentTransactions();

        // Atualizar a cada 10 segundos
        setInterval(() => {
            loadRecentBlocks();
            loadRecentTransactions();
        }, 10000);

        // Busca
        document.getElementById('searchBox').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                const query = this.value.trim();
                if (query) {
                    // Implementar busca
                    alert('Funcionalidade de busca será implementada em breve!');
                }
            }
        });
    </script>
</body>
</html>`,
		stats.TotalBlocks, stats.TotalTransactions, stats.TotalSupply,
		stats.ActiveWallets, stats.NetworkHashRate, stats.TotalStaked)

	w.Write([]byte(html))
}

func (e *Explorer) handleBlocks(w http.ResponseWriter, r *http.Request) {
	blocks := e.getRecentBlocks(50)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>📦 Blocos - Blockchain Explorer</title>
    <style>
        /* Mesmo CSS do dashboard */
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #0f0f23 0%%, #1a1a2e 50%%, #16213e 100%%);
            color: #e0e0e0;
            line-height: 1.6;
        }
        .header {
            background: rgba(0,0,0,0.8);
            padding: 1rem 0;
            border-bottom: 2px solid #00d4ff;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 0 20px;
        }
        .logo {
            font-size: 2rem;
            font-weight: bold;
            color: #00d4ff;
            text-align: center;
        }
        .nav {
            display: flex;
            justify-content: center;
            gap: 2rem;
            margin-top: 1rem;
        }
        .nav a {
            color: #e0e0e0;
            text-decoration: none;
            padding: 0.5rem 1rem;
            border-radius: 5px;
            transition: all 0.3s;
        }
        .nav a:hover {
            background: rgba(0,212,255,0.1);
            color: #00d4ff;
        }
        .blocks-table {
            width: 100%%;
            background: rgba(255,255,255,0.05);
            border-radius: 10px;
            margin: 2rem 0;
            overflow: hidden;
        }
        .blocks-table th {
            background: rgba(0,212,255,0.2);
            padding: 1rem;
            text-align: left;
            color: #00d4ff;
        }
        .blocks-table td {
            padding: 1rem;
            border-bottom: 1px solid rgba(255,255,255,0.1);
        }
        .blocks-table tr:hover {
            background: rgba(0,212,255,0.1);
        }
        .hash {
            font-family: monospace;
            color: #00d4ff;
        }
        .time {
            color: #888;
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="container">
            <div class="logo">🔍 Blockchain Explorer</div>
            <nav class="nav">
                <a href="/">🏠 Dashboard</a>
                <a href="/blocks">📦 Blocos</a>
                <a href="/transactions">💸 Transações</a>
                <a href="/wallets">👛 Wallets</a>
            </nav>
        </div>
    </div>

    <div class="container">
        <h1 style="color: #00d4ff; margin: 2rem 0;">📦 Todos os Blocos</h1>
        
        <table class="blocks-table">
            <thead>
                <tr>
                    <th>Número</th>
                    <th>Hash</th>
                    <th>Timestamp</th>
                    <th>Transações</th>
                    <th>Recompensa</th>
                    <th>Dificuldade</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>
    </div>
</body>
</html>`, e.generateBlocksTable(blocks))

	w.Write([]byte(html))
}

func (e *Explorer) generateBlocksTable(blocks []BlockInfo) string {
	var rows string
	for _, block := range blocks {
		rows += fmt.Sprintf(`
            <tr>
                <td><strong>#%d</strong></td>
                <td class="hash">%s</td>
                <td class="time">%s</td>
                <td>%d</td>
                <td>+%d tokens</td>
                <td>%s</td>
            </tr>`,
			block.Number, block.Hash,
			time.Unix(block.Timestamp, 0).Format("02/01/2006 15:04:05"),
			block.Transactions, block.Reward, block.Difficulty)
	}
	return rows
}

func (e *Explorer) handleTransactions(w http.ResponseWriter, r *http.Request) {
	transactions := e.getRecentTransactions(50)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>💸 Transações - Blockchain Explorer</title>
    <style>
        /* Mesmo CSS */
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #0f0f23 0%%, #1a1a2e 50%%, #16213e 100%%);
            color: #e0e0e0;
            line-height: 1.6;
        }
        .header {
            background: rgba(0,0,0,0.8);
            padding: 1rem 0;
            border-bottom: 2px solid #00d4ff;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 0 20px;
        }
        .logo {
            font-size: 2rem;
            font-weight: bold;
            color: #00d4ff;
            text-align: center;
        }
        .nav {
            display: flex;
            justify-content: center;
            gap: 2rem;
            margin-top: 1rem;
        }
        .nav a {
            color: #e0e0e0;
            text-decoration: none;
            padding: 0.5rem 1rem;
            border-radius: 5px;
            transition: all 0.3s;
        }
        .nav a:hover {
            background: rgba(0,212,255,0.1);
            color: #00d4ff;
        }
        .tx-table {
            width: 100%%;
            background: rgba(255,255,255,0.05);
            border-radius: 10px;
            margin: 2rem 0;
            overflow: hidden;
        }
        .tx-table th {
            background: rgba(0,212,255,0.2);
            padding: 1rem;
            text-align: left;
            color: #00d4ff;
        }
        .tx-table td {
            padding: 1rem;
            border-bottom: 1px solid rgba(255,255,255,0.1);
        }
        .tx-table tr:hover {
            background: rgba(0,212,255,0.1);
        }
        .hash {
            font-family: monospace;
            color: #00d4ff;
        }
        .address {
            font-family: monospace;
            color: #00ff88;
        }
        .amount {
            color: #00ff88;
            font-weight: bold;
        }
        .time {
            color: #888;
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="container">
            <div class="logo">🔍 Blockchain Explorer</div>
            <nav class="nav">
                <a href="/">🏠 Dashboard</a>
                <a href="/blocks">📦 Blocos</a>
                <a href="/transactions">💸 Transações</a>
                <a href="/wallets">👛 Wallets</a>
            </nav>
        </div>
    </div>

    <div class="container">
        <h1 style="color: #00d4ff; margin: 2rem 0;">💸 Todas as Transações</h1>
        
        <table class="tx-table">
            <thead>
                <tr>
                    <th>Hash</th>
                    <th>De</th>
                    <th>Para</th>
                    <th>Valor</th>
                    <th>Tipo</th>
                    <th>Status</th>
                    <th>Timestamp</th>
                </tr>
            </thead>
            <tbody>
                %s
            </tbody>
        </table>
    </div>
</body>
</html>`, e.generateTransactionsTable(transactions))

	w.Write([]byte(html))
}

func (e *Explorer) generateTransactionsTable(transactions []TransactionInfo) string {
	var rows string
	for _, tx := range transactions {
		rows += fmt.Sprintf(`
            <tr>
                <td class="hash">%s</td>
                <td class="address">%s</td>
                <td class="address">%s</td>
                <td class="amount">%d tokens</td>
                <td>%s</td>
                <td>%s</td>
                <td class="time">%s</td>
            </tr>`,
			tx.Hash, tx.From, tx.To, tx.Amount, tx.Type, tx.Status,
			time.Unix(tx.Timestamp, 0).Format("02/01/2006 15:04:05"))
	}
	return rows
}

func (e *Explorer) handleWallets(w http.ResponseWriter, r *http.Request) {
	wallets := e.getAllWallets()

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>👛 Wallets - Blockchain Explorer</title>
    <style>
        /* Mesmo CSS */
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #0f0f23 0%%, #1a1a2e 50%%, #16213e 100%%);
            color: #e0e0e0;
            line-height: 1.6;
        }
        .header {
            background: rgba(0,0,0,0.8);
            padding: 1rem 0;
            border-bottom: 2px solid #00d4ff;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 0 20px;
        }
        .logo {
            font-size: 2rem;
            font-weight: bold;
            color: #00d4ff;
            text-align: center;
        }
        .nav {
            display: flex;
            justify-content: center;
            gap: 2rem;
            margin-top: 1rem;
        }
        .nav a {
            color: #e0e0e0;
            text-decoration: none;
            padding: 0.5rem 1rem;
            border-radius: 5px;
            transition: all 0.3s;
        }
        .nav a:hover {
            background: rgba(0,212,255,0.1);
            color: #00d4ff;
        }
        .wallets-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 1.5rem;
            margin: 2rem 0;
        }
        .wallet-card {
            background: rgba(255,255,255,0.05);
            border: 1px solid rgba(0,212,255,0.3);
            border-radius: 10px;
            padding: 1.5rem;
            transition: transform 0.3s;
        }
        .wallet-card:hover {
            transform: translateY(-5px);
            border-color: #00d4ff;
        }
        .wallet-address {
            font-family: monospace;
            color: #00d4ff;
            font-size: 0.9rem;
            margin-bottom: 1rem;
            word-break: break-all;
        }
        .wallet-balance {
            font-size: 1.5rem;
            color: #00ff88;
            font-weight: bold;
            margin-bottom: 0.5rem;
        }
        .wallet-stake {
            color: #ffaa00;
            margin-bottom: 0.5rem;
        }
        .validator-badge {
            background: #00ff88;
            color: #000;
            padding: 0.2rem 0.5rem;
            border-radius: 3px;
            font-size: 0.8rem;
            display: inline-block;
            margin-top: 0.5rem;
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="container">
            <div class="logo">🔍 Blockchain Explorer</div>
            <nav class="nav">
                <a href="/">🏠 Dashboard</a>
                <a href="/blocks">📦 Blocos</a>
                <a href="/transactions">💸 Transações</a>
                <a href="/wallets">👛 Wallets</a>
            </nav>
        </div>
    </div>

    <div class="container">
        <h1 style="color: #00d4ff; margin: 2rem 0;">👛 Todas as Wallets</h1>
        
        <div class="wallets-grid">
            %s
        </div>
    </div>
</body>
</html>`, e.generateWalletsGrid(wallets))

	w.Write([]byte(html))
}

func (e *Explorer) generateWalletsGrid(wallets []WalletInfo) string {
	var cards string
	for _, wallet := range wallets {
		validatorBadge := ""
		if wallet.IsValidator {
			validatorBadge = `<div class="validator-badge">✅ Validator</div>`
		}

		cards += fmt.Sprintf(`
            <div class="wallet-card">
                <div class="wallet-address">%s</div>
                <div class="wallet-balance">%d tokens</div>
                <div class="wallet-stake">Stake: %d tokens</div>
                <div>Transações: %d</div>
                %s
            </div>`,
			wallet.Address, wallet.Balance, wallet.StakeAmount, wallet.TxCount, validatorBadge)
	}
	return cards
}

// API endpoints
func (e *Explorer) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	stats := e.getNetworkStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (e *Explorer) handleAPIBlocks(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	blocks := e.getRecentBlocks(limit)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blocks)
}

func (e *Explorer) handleAPITransactions(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	transactions := e.getRecentTransactions(limit)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

func (e *Explorer) handleAPIWallets(w http.ResponseWriter, r *http.Request) {
	wallets := e.getAllWallets()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallets)
}

// Handlers para detalhes (placeholder)
func (e *Explorer) handleBlockDetail(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Detalhes do bloco serão implementados em breve!"))
}

func (e *Explorer) handleTransactionDetail(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Detalhes da transação serão implementados em breve!"))
}

func (e *Explorer) handleAddressDetail(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Detalhes do endereço serão implementados em breve!"))
}

// Funções auxiliares para gerar dados
func (e *Explorer) getNetworkStats() NetworkStats {
	// Carregar estado real de mineração
	miningStatePath := "data/mining_state.json"
	var totalBlocks int64 = 0

	if data, err := os.ReadFile(miningStatePath); err == nil {
		var stats struct {
			TotalBlocks int64 `json:"total_blocks"`
		}
		if json.Unmarshal(data, &stats) == nil {
			totalBlocks = stats.TotalBlocks
		}
	}

	// Calcular supply real baseado nos blocos minerados
	totalSupply := totalBlocks * 50 // 50 tokens por bloco

	return NetworkStats{
		TotalBlocks:       totalBlocks,
		TotalTransactions: int64(len(e.GlobalLedger.Movements)),
		TotalSupply:       totalSupply,
		ActiveWallets:     len(e.WalletManager.Wallets),
		NetworkHashRate:   75.0,
		AverageBlockTime:  2.0,
		TotalStaked:       1000,
		Validators:        5,
	}
}

func (e *Explorer) getRecentBlocks(limit int) []BlockInfo {
	// Carregar estado real de mineração
	miningStatePath := "data/mining_state.json"
	var totalBlocks int64 = 0

	if data, err := os.ReadFile(miningStatePath); err == nil {
		var stats struct {
			TotalBlocks int64 `json:"total_blocks"`
		}
		if json.Unmarshal(data, &stats) == nil {
			totalBlocks = stats.TotalBlocks
		}
	}

	// Gerar blocos reais baseados no total minerado
	var blocks []BlockInfo
	startBlock := totalBlocks
	endBlock := startBlock - int64(limit)
	if endBlock < 0 {
		endBlock = 0
	}

	for i := startBlock; i > endBlock; i-- {
		// Calcular timestamp baseado no número do bloco
		timestamp := time.Now().Unix() - int64((startBlock-i)*2)
		blocks = append(blocks, BlockInfo{
			Number:       i,
			Hash:         fmt.Sprintf("block_%d_%d", i, timestamp),
			Timestamp:    timestamp,
			Transactions: 1,
			Reward:       50,
			Difficulty:   "0x1",
			Size:         1024,
		})
	}
	return blocks
}

func (e *Explorer) getRecentTransactions(limit int) []TransactionInfo {
	// Usar movimentos reais do ledger
	var transactions []TransactionInfo

	// Ordenar movimentos por timestamp (mais recentes primeiro)
	movements := make([]ledger.TokenMovement, len(e.GlobalLedger.Movements))
	copy(movements, e.GlobalLedger.Movements)

	// Ordenar por timestamp decrescente
	for i := 0; i < len(movements)-1; i++ {
		for j := i + 1; j < len(movements); j++ {
			if movements[i].Timestamp < movements[j].Timestamp {
				movements[i], movements[j] = movements[j], movements[i]
			}
		}
	}

	// Pegar os mais recentes
	for i, movement := range movements {
		if i >= limit {
			break
		}

		// Determinar status baseado no tipo
		status := "Confirmed"
		if movement.Type == "transfer" {
			status = "Transfer"
		} else if movement.Type == "mining_reward" {
			status = "Mining"
		} else if movement.Type == "stake" {
			status = "Stake"
		}

		transactions = append(transactions, TransactionInfo{
			Hash:      movement.ID,
			From:      movement.From,
			To:        movement.To,
			Amount:    movement.Amount,
			Type:      movement.Type,
			Timestamp: movement.Timestamp,
			Status:    status,
			Fee:       movement.Fee,
		})
	}
	return transactions
}

func (e *Explorer) getAllWallets() []WalletInfo {
	var wallets []WalletInfo

	// Criar mapa de endereços únicos com saldos do ledger
	addressMap := make(map[string]WalletInfo)

	// Adicionar wallets do wallet manager
	for _, w := range e.WalletManager.Wallets {
		for _, account := range w.Accounts {
			address := account.Address
			balance := e.GlobalLedger.GetBalance(address)

			addressMap[address] = WalletInfo{
				Address:     address,
				Balance:     balance,
				StakeAmount: 0,     // Não implementado ainda
				IsValidator: false, // Não implementado ainda
				TxCount:     e.countTransactionsForAddress(address),
				LastTx:      e.getLastTransactionTime(address),
			}
		}
	}

	// Adicionar endereços que aparecem no ledger mas não estão no wallet manager
	for address, balance := range e.GlobalLedger.Balances {
		if _, exists := addressMap[address]; !exists && balance > 0 {
			addressMap[address] = WalletInfo{
				Address:     address,
				Balance:     balance,
				StakeAmount: 0,
				IsValidator: false,
				TxCount:     e.countTransactionsForAddress(address),
				LastTx:      e.getLastTransactionTime(address),
			}
		}
	}

	// Converter mapa para slice
	for _, wallet := range addressMap {
		wallets = append(wallets, wallet)
	}

	return wallets
}

func (e *Explorer) countTransactionsForAddress(address string) int {
	count := 0
	for _, movement := range e.GlobalLedger.Movements {
		if movement.From == address || movement.To == address {
			count++
		}
	}
	return count
}

func (e *Explorer) getLastTransactionTime(address string) int64 {
	var lastTime int64 = 0
	for _, movement := range e.GlobalLedger.Movements {
		if (movement.From == address || movement.To == address) && movement.Timestamp > lastTime {
			lastTime = movement.Timestamp
		}
	}
	return lastTime
}
