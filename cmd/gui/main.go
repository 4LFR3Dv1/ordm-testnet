package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"ordm-main/pkg/api"
	"ordm-main/pkg/auth"
	"ordm-main/pkg/ledger"
	"ordm-main/pkg/network"
	"ordm-main/pkg/storage"
	"ordm-main/pkg/wallet"
)

type MiningStats struct {
	TotalBlocks    int64   `json:"total_blocks"`
	TotalRewards   int64   `json:"total_rewards"`
	EnergyCost     float64 `json:"energy_cost"`
	Profitability  float64 `json:"profitability"`
	HashRate       float64 `json:"hash_rate"`
	Uptime         int64   `json:"uptime"`
	StakeAmount    int64   `json:"stake_amount"`
	ValidatorLevel string  `json:"validator_level"`
}

type NodeInfo struct {
	Name          string           `json:"name"`
	Port          int              `json:"port"`
	Status        string           `json:"status"`
	IsRunning     bool             `json:"is_running"`
	IsMining      bool             `json:"is_mining"`
	MiningStats   MiningStats      `json:"mining_stats"`
	Balance       map[string]int64 `json:"balance"`
	Peers         []string         `json:"peers"`
	Difficulty    int              `json:"difficulty"`
	EnergyPrice   float64          `json:"energy_price"`
	WalletName    string           `json:"wallet_name"`
	WalletAddress string           `json:"wallet_address"`
}

type TwoFactorAuth struct {
	CurrentPIN      string    `json:"current_pin"`
	GeneratedAt     time.Time `json:"generated_at"`
	ExpiresAt       time.Time `json:"expires_at"`
	Attempts        int       `json:"attempts"`
	MaxAttempts     int       `json:"max_attempts"`
	LockedUntil     time.Time `json:"locked_until"`
	SessionToken    string    `json:"session_token"`
	IsAuthenticated bool      `json:"is_authenticated"`
}

type BlockchainGUI struct {
	Node          NodeInfo               `json:"node"`
	Logs          []string               `json:"logs"`
	IsRunning     bool                   `json:"is_running"`
	WalletManager *wallet.WalletManager  `json:"-"`
	GlobalLedger  *ledger.GlobalLedger   `json:"-"`
	TwoFactorAuth *TwoFactorAuth         `json:"-"`
	UserManager   *auth.UserManager      `json:"-"`
	RenderStorage *storage.RenderStorage `json:"-"`
}

var gui BlockchainGUI
var miningTicker *time.Ticker
var miningStop chan bool

func NewTwoFactorAuth() *TwoFactorAuth {
	return &TwoFactorAuth{
		MaxAttempts: 3,
		Attempts:    0,
	}
}

func (tfa *TwoFactorAuth) GeneratePIN() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)

	num := int(bytes[0])<<16 | int(bytes[1])<<8 | int(bytes[2])
	pin := fmt.Sprintf("%06d", num%1000000)

	tfa.CurrentPIN = pin
	tfa.GeneratedAt = time.Now()
	tfa.ExpiresAt = time.Now().Add(10 * time.Second) // 10 segundos
	tfa.Attempts = 0
	tfa.IsAuthenticated = false
	tfa.SessionToken = hex.EncodeToString(bytes)

	return pin
}

func (tfa *TwoFactorAuth) ValidatePIN(pin string) (bool, string) {
	if time.Now().Before(tfa.LockedUntil) {
		return false, "Sistema bloqueado. Tente novamente em alguns minutos."
	}

	if time.Now().After(tfa.ExpiresAt) {
		return false, "PIN expirado. Gere um novo PIN."
	}

	if pin == tfa.CurrentPIN {
		tfa.IsAuthenticated = true
		tfa.Attempts = 0
		return true, "Login realizado com sucesso!"
	}

	tfa.Attempts++
	if tfa.Attempts >= tfa.MaxAttempts {
		tfa.LockedUntil = time.Now().Add(15 * time.Minute)
		return false, "Muitas tentativas. Sistema bloqueado por 15 minutos."
	}

	return false, fmt.Sprintf("PIN incorreto. Tentativas restantes: %d", tfa.MaxAttempts-tfa.Attempts)
}

func (tfa *TwoFactorAuth) IsUserAuthenticated() bool {
	return tfa.IsAuthenticated
}

func (tfa *TwoFactorAuth) Logout() {
	tfa.IsAuthenticated = false
	tfa.SessionToken = ""
}

// loadMiningState carrega o estado de mineração
func loadMiningState() (int, error) {
	// Em produção, usar diretório temporário
	dataPath := "./data"
	if os.Getenv("PORT") != "" || os.Getenv("NODE_ENV") == "production" {
		dataPath = "/tmp/ordm-data"
	}

	// Criar diretório se não existir
	os.MkdirAll(dataPath, 0755)

	filePath := filepath.Join(dataPath, "mining_state.json")

	// Se o arquivo não existir, retornar 0
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return 0, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	var state struct {
		BlocksMined int `json:"blocks_mined"`
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return 0, err
	}

	return state.BlocksMined, nil
}

// saveMiningState salva o estado de mineração
func saveMiningState(blocksMined int) error {
	// Em produção, usar diretório temporário
	dataPath := "./data"
	if os.Getenv("PORT") != "" || os.Getenv("NODE_ENV") == "production" {
		dataPath = "/tmp/ordm-data"
	}

	// Criar diretório se não existir
	os.MkdirAll(dataPath, 0755)

	filePath := filepath.Join(dataPath, "mining_state.json")

	state := struct {
		BlocksMined int `json:"blocks_mined"`
	}{
		BlocksMined: blocksMined,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func loadExistingWallets() {
	// Carregar wallets existentes se houver
	fmt.Println("🔑 Carregando wallets existentes...")
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if gui.UserManager.GetActiveUser() == nil {
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Login - Blockchain 2-Layer</title>
			<style>
				body { font-family: Arial, sans-serif; background: #1a1a1a; color: #fff; margin: 0; padding: 20px; }
				.login-container { max-width: 600px; margin: 50px auto; background: #2a2a2a; padding: 30px; border-radius: 10px; }
				.form-group { margin-bottom: 20px; }
				label { display: block; margin-bottom: 5px; font-weight: bold; }
				input { width: 100%; padding: 12px; border: none; border-radius: 5px; background: #333; color: #fff; font-size: 16px; }
				button { width: 100%; padding: 15px; background: #007bff; color: #fff; border: none; border-radius: 5px; cursor: pointer; font-size: 16px; font-weight: bold; margin-bottom: 10px; }
				button:hover { background: #0056b3; }
				.info-box { background: #333; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
				.pin-display { background: #28a745; padding: 10px; border-radius: 5px; text-align: center; font-size: 18px; font-weight: bold; margin: 10px 0; }
				.simple-login { background: #28a745; }
				.simple-login:hover { background: #218838; }
				.advanced-login { background: #17a2b8; }
				.advanced-login:hover { background: #138496; }
				.tabs { display: flex; margin-bottom: 20px; }
				.tab { flex: 1; padding: 10px; background: #333; border: none; color: #fff; cursor: pointer; }
				.tab.active { background: #007bff; }
				.tab-content { display: none; }
				.tab-content.active { display: block; }
				.wallet-info { background: #17a2b8; padding: 15px; border-radius: 5px; margin: 10px 0; font-family: monospace; }
			</style>
		</head>
		<body>
			<div class="login-container">
				<h2>🔐 Login Blockchain 2-Layer</h2>
				
				<div class="tabs">
					<button class="tab active" onclick="showTab('simple')">🚀 Login Simples</button>
					<button class="tab" onclick="showTab('advanced')">🔐 Login Avançado</button>
				</div>
				
				<div id="simple-tab" class="tab-content active">
					<div class="info-box">
						<h3>📋 Login Simples (Recomendado):</h3>
						<p>Para testes e uso rápido. Não requer public key.</p>
					</div>
					
					<form action="/user-login" method="POST">
						<div class="form-group">
							<label>Usuário:</label>
							<input type="text" name="username" value="admin" required>
						</div>
						<div class="form-group">
							<label>Senha:</label>
							<input type="password" name="password" value="admin123" required>
						</div>
						<button type="submit" class="simple-login">🚀 Entrar na Blockchain</button>
					</form>
				</div>
				
				<div id="advanced-tab" class="tab-content">
					<div class="info-box">
						<h3>🔐 Login Seguro por Wallet:</h3>
						<p><strong>Public Key:</strong> Chave pública da sua wallet</p>
						<p><strong>PIN Único:</strong> PIN secreto gerado quando a wallet foi criada</p>
						<p><strong>⚠️ Segurança:</strong> Cada wallet tem seu próprio PIN único e secreto</p>
					</div>
					
					<form action="/advanced-login" method="POST">
						<div class="form-group">
							<label>🔑 Public Key da Wallet:</label>
							<input type="text" name="publicKey" placeholder="Ex: wallet_1234567890" required minlength="8">
						</div>
						<div class="form-group">
							<label>🔐 PIN Único da Wallet:</label>
							<input type="text" name="pin" placeholder="6 dígitos" required maxlength="6" pattern="[0-9]{6}">
						</div>
						<button type="submit" class="advanced-login">🔐 Acessar Wallet</button>
					</form>
					
					<div class="info-box" style="margin-top: 20px;">
						<h4>ℹ️ Como funciona:</h4>
						<ul>
							<li>✅ Cada wallet tem um PIN único e secreto</li>
							<li>✅ O PIN é gerado apenas uma vez quando a wallet é criada</li>
							<li>✅ Guarde o PIN com segurança - não pode ser recuperado</li>
							<li>✅ Após 3 tentativas incorretas, a wallet é bloqueada por 5 minutos</li>
						</ul>
					</div>
				</div>
			</div>
			
			<script>
				function showTab(tabName) {
					// Esconder todas as tabs
					document.querySelectorAll('.tab-content').forEach(tab => tab.classList.remove('active'));
					document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
					
					// Mostrar tab selecionada
					document.getElementById(tabName + '-tab').classList.add('active');
					event.target.classList.add('active');
				}
				
				// Sistema de login seguro implementado
				// Cada wallet tem seu próprio PIN único
			</script>
		</body>
		</html>`
		w.Write([]byte(html))
		return
	}

	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Blockchain 2-Layer - Node Minerador</title>
		<style>
			body { font-family: 'Courier New', monospace; background: #000; color: #0f0; margin: 0; padding: 10px; font-size: 12px; }
			.container { max-width: 1200px; margin: 0 auto; }
			.header { background: #111; padding: 10px; border: 1px solid #0f0; margin-bottom: 10px; }
			.panel { background: #111; border: 1px solid #0f0; margin-bottom: 10px; padding: 10px; }
			.panel h3 { margin: 0 0 10px 0; color: #0f0; }
			.status { display: inline-block; padding: 5px 10px; margin: 5px; border: 1px solid #0f0; }
			.status.running { background: #0f0; color: #000; }
			.status.stopped { background: #f00; color: #fff; }
			.button { background: #333; color: #0f0; border: 1px solid #0f0; padding: 5px 10px; cursor: pointer; margin: 2px; }
			.button:hover { background: #0f0; color: #000; }
			.log { background: #000; border: 1px solid #0f0; padding: 10px; height: 200px; overflow-y: scroll; font-family: monospace; }
			.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
			.wallet-list { max-height: 150px; overflow-y: scroll; }
			.wallet-item { padding: 5px; border-bottom: 1px solid #333; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h2>🔗 Blockchain 2-Layer - Node Minerador</h2>
				<div>Usuário: ` + gui.UserManager.GetActiveUser().Username + ` | <a href="/user-logout" style="color: #0f0;">Logout</a></div>
			</div>
			
			<div class="grid">
				<div class="panel">
					<h3>🚀 Controle Unificado</h3>
					<div>
						<span class="status ` + getStatusClass() + `">` + gui.Node.Status + `</span>
						<button class="button" onclick="startNodeAndMining()">🚀 Iniciar Node + Mineração</button>
						<button class="button" onclick="stopNodeAndMining()">⏹️ Parar Tudo</button>
					</div>
					<div>
						<button class="button" onclick="startMining()">⛏️ Apenas Mineração</button>
						<button class="button" onclick="stopMining()">⏸️ Parar Mineração</button>
					</div>
				</div>
				
				<div class="panel">
					<h3>💰 Wallet & Transferências</h3>
					<div>
						<button class="button" onclick="createWallet()">Criar Wallet</button>
						<button class="button" onclick="showTransferForm()">Transferir</button>
					</div>
					<div class="wallet-list" id="walletList">
					</div>
				</div>
			</div>
			
			<div class="grid">
				<div class="panel">
					<h3>🎯 Staking & Validação</h3>
					<div>
						<button class="button" onclick="showStakingForm()">💎 Fazer Stake</button>
						<button class="button" onclick="showValidatorForm()">👑 Tornar-se Validator</button>
					</div>
					<div>
						<div>Stake Atual: <span id="currentStake">0</span> tokens</div>
						<div>Status Validator: <span id="validatorStatus">❌ Não é validator</span></div>
						<div>Recompensas Staking: <span id="stakingRewards">0</span> tokens</div>
					</div>
				</div>
				
				<div class="panel">
					<h3>👤 Painel do Usuário</h3>
					<div>
						<div>Usuário: <strong>` + gui.UserManager.GetActiveUser().Username + `</strong></div>
						<div>Wallet Ativa: <span id="activeWallet">Carregando...</span></div>
						<div>Saldo Total: <span id="totalBalance">0</span> tokens</div>
						<div>Nível: <span id="userLevel">Iniciante</span></div>
					</div>
					<div>
						<button class="button" onclick="showUserStats()">📊 Ver Estatísticas</button>
						<button class="button" onclick="showTransactionHistory()">📜 Histórico</button>
					</div>
				</div>
			</div>
			
			<div class="panel">
				<h3>📊 Estatísticas de Mineração</h3>
				<div>Blocos Minerados: <span id="totalBlocks">0</span></div>
				<div>Recompensas: <span id="totalRewards">0</span> tokens</div>
				<div>Custo Energia: $<span id="energyCost">0.00</span></div>
				<div>Lucratividade: $<span id="profitability">0.00</span></div>
				<div>Hash Rate: <span id="hashRate">0</span> H/s</div>
			</div>
			
			<div class="panel">
				<h3>📝 Logs em Tempo Real</h3>
				<div class="log" id="logContainer">
				</div>
			</div>
		</div>
		
		<script>
			function startNodeAndMining() { 
				fetch('/start', {method: 'POST'})
					.then(() => {
						setTimeout(() => {
							fetch('/start-mining', {method: 'POST'});
						}, 1000);
					});
			}
			function stopNodeAndMining() { 
				fetch('/stop-mining', {method: 'POST'})
					.then(() => {
						setTimeout(() => {
							fetch('/stop', {method: 'POST'});
						}, 1000);
					});
			}
			function startNode() { fetch('/start', {method: 'POST'}); }
			function stopNode() { fetch('/stop', {method: 'POST'}); }
			function startMining() { fetch('/start-mining', {method: 'POST'}); }
			function stopMining() { fetch('/stop-mining', {method: 'POST'}); }
			function createWallet() { fetch('/create-wallet', {method: 'POST'}); }
			
			function showStakingForm() {
				const amount = prompt('Quantidade de tokens para stake:');
				if (amount && !isNaN(amount)) {
					fetch('/stake', {
						method: 'POST',
						headers: {'Content-Type': 'application/json'},
						body: JSON.stringify({amount: parseInt(amount)})
					}).then(() => updateInterface());
				}
			}
			
			function showValidatorForm() {
				const stake = prompt('Quantidade mínima para validator (1000 tokens):');
				if (stake && !isNaN(stake) && parseInt(stake) >= 1000) {
					fetch('/become-validator', {
						method: 'POST',
						headers: {'Content-Type': 'application/json'},
						body: JSON.stringify({stake: parseInt(stake)})
					}).then(() => updateInterface());
				} else {
					alert('É necessário pelo menos 1000 tokens para se tornar validator!');
				}
			}
			
			function showUserStats() {
				alert('Estatísticas do usuário serão implementadas em breve!');
			}
			
			function showTransactionHistory() {
				alert('Histórico de transações será implementado em breve!');
			}
			
			function showTransferForm() {
				const amount = prompt('Quantidade de tokens:');
				const to = prompt('Endereço de destino:');
				if (amount && to) {
					fetch('/transfer', {
						method: 'POST',
						headers: {'Content-Type': 'application/json'},
						body: JSON.stringify({amount: amount, to: to})
					});
				}
			}
			
			function updateInterface() {
				fetch('/status')
					.then(response => response.json())
					.then(data => {
						document.getElementById('totalBlocks').textContent = data.node.mining_stats.total_blocks;
						document.getElementById('totalRewards').textContent = data.node.mining_stats.total_rewards;
						document.getElementById('energyCost').textContent = data.node.mining_stats.energy_cost.toFixed(2);
						document.getElementById('profitability').textContent = data.node.mining_stats.profitability.toFixed(2);
						document.getElementById('hashRate').textContent = data.node.mining_stats.hash_rate.toFixed(0);
						
						// Atualizar informações de staking
						document.getElementById('currentStake').textContent = data.node.mining_stats.stake_amount || 0;
						document.getElementById('validatorStatus').textContent = data.node.mining_stats.validator_level === 'Validator' ? '✅ É validator' : '❌ Não é validator';
						document.getElementById('stakingRewards').textContent = Math.floor((data.node.mining_stats.stake_amount || 0) * 0.05); // 5% de recompensa
						
						// Atualizar informações do usuário
						document.getElementById('activeWallet').textContent = data.node.wallet_address || 'Nenhuma wallet ativa';
						document.getElementById('userLevel').textContent = data.node.mining_stats.validator_level || 'Iniciante';
						
						// Calcular saldo total
						let totalBalance = 0;
						for (const [address, balance] of Object.entries(data.node.balance)) {
							totalBalance += balance;
						}
						document.getElementById('totalBalance').textContent = totalBalance;
						
						const walletList = document.getElementById('walletList');
						walletList.innerHTML = '';
						for (const [address, balance] of Object.entries(data.node.balance)) {
							walletList.innerHTML += '<div class="wallet-item">' + address.substring(0, 10) + '... | ' + balance + ' tokens</div>';
						}
					});
			}
			
			function updateLogs() {
				fetch('/ledger')
					.then(response => response.json())
					.then(data => {
						const logContainer = document.getElementById('logContainer');
						logContainer.innerHTML = data.logs.join('<br>');
						logContainer.scrollTop = logContainer.scrollHeight;
					});
			}
			
			setInterval(updateInterface, 2000);
			setInterval(updateLogs, 2000);
			
			updateInterface();
			updateLogs();
		</script>
	</body>
	</html>`
	w.Write([]byte(html))
}

func handleUserLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := gui.UserManager.Login(username, password)
		if err == nil {
			fmt.Printf("✅ Usuário logado: %s\n", user.Username)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

func handleUserLogout(w http.ResponseWriter, r *http.Request) {
	gui.UserManager.Logout()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		pin := r.FormValue("pin")

		valid, _ := gui.TwoFactorAuth.ValidatePIN(pin)
		if valid {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	gui.TwoFactorAuth.Logout()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleAdvancedLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	publicKey := r.FormValue("publicKey")
	pin := r.FormValue("pin")

	// Validar formato da public key (mínimo 8 caracteres)
	if len(publicKey) < 8 {
		fmt.Printf("❌ Public Key inválida: deve ter pelo menos 8 caracteres\n")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Validar formato do PIN (6 dígitos)
	if len(pin) != 6 {
		fmt.Printf("❌ PIN inválido: deve ter 6 dígitos\n")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Tentar fazer login na wallet
	walletAuth, err := gui.UserManager.LoginWallet(publicKey, pin)
	if err != nil {
		// Se a wallet não existe, criar uma nova
		if err.Error() == "wallet não encontrada" {
			fmt.Printf("🔑 Wallet não encontrada, criando nova wallet: %s\n", publicKey)
			walletAuth, err = gui.UserManager.CreateWalletAuth(publicKey, pin)
			if err != nil {
				fmt.Printf("❌ Erro ao criar wallet: %s\n", err.Error())
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		} else {
			fmt.Printf("❌ Login falhou: %s\n", err.Error())
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	// Login bem-sucedido
	fmt.Printf("✅ Login seguro realizado: Public Key %s | PIN: %s\n", publicKey, pin)

	// Definir wallet ativa
	gui.Node.WalletAddress = walletAuth.PublicKey

	// Criar wallet se não existir
	if _, exists := gui.Node.Balance[walletAuth.PublicKey]; !exists {
		gui.Node.Balance[walletAuth.PublicKey] = 0
		fmt.Printf("🔑 Wallet ativada: %s\n", walletAuth.PublicKey)
	}

	// Definir usuário ativo para mostrar o dashboard
	user := &auth.User{
		ID:           walletAuth.PublicKey,
		Username:     fmt.Sprintf("wallet_%s", publicKey[:10]),
		PasswordHash: "secure_wallet",
		WalletIDs:    []string{walletAuth.PublicKey},
		CreatedAt:    time.Now(),
		LastLogin:    time.Now(),
		IsActive:     true,
	}
	gui.UserManager.ActiveUser = user

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleGeneratePIN(w http.ResponseWriter, r *http.Request) {
	pin := gui.TwoFactorAuth.GeneratePIN()
	fmt.Printf("🔐 Novo PIN 2FA gerado: %s (válido por 10 segundos)\n", pin)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"pin": pin,
	})
}

func handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		gui.Node.IsRunning = true
		gui.Node.Status = "Rodando"
		addLog("🚀 Node minerador iniciado")
	}
}

func handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		gui.Node.IsRunning = false
		gui.Node.Status = "Parado"
		addLog("🛑 Node minerador parado")
	}
}

func handleStartMining(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		gui.Node.IsMining = true
		startMiningProcess()
		addLog("⛏️ Mineração iniciada")
	}
}

func handleStopMining(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		gui.Node.IsMining = false
		stopMiningProcess()
		addLog("⏸️ Mineração parada")
	}
}

func handleStake(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Amount int64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Verificar se tem saldo suficiente
	if gui.Node.WalletAddress != "" {
		if gui.Node.Balance[gui.Node.WalletAddress] < request.Amount {
			http.Error(w, "Saldo insuficiente", http.StatusBadRequest)
			return
		}

		// Deduzir do saldo e adicionar ao stake
		gui.Node.Balance[gui.Node.WalletAddress] -= request.Amount
		gui.Node.MiningStats.StakeAmount += request.Amount

		// Atualizar nível de validator
		if gui.Node.MiningStats.StakeAmount >= 1000 {
			gui.Node.MiningStats.ValidatorLevel = "Validator"
		}

		// Registrar no ledger como recompensa de staking
		gui.GlobalLedger.AddMiningReward(gui.Node.WalletAddress, request.Amount, "stake_deposit")

		addLog(fmt.Sprintf("💎 Stake realizado: +%d tokens | Total: %d tokens | Saldo: %d",
			request.Amount, gui.Node.MiningStats.StakeAmount, gui.Node.Balance[gui.Node.WalletAddress]))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"stake_amount":    gui.Node.MiningStats.StakeAmount,
		"validator_level": gui.Node.MiningStats.ValidatorLevel,
	})
}

func handleBecomeValidator(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Stake int64 `json:"stake"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Verificar stake mínimo
	if request.Stake < 1000 {
		http.Error(w, "Stake mínimo de 1000 tokens necessário", http.StatusBadRequest)
		return
	}

	// Verificar se tem saldo suficiente
	if gui.Node.WalletAddress != "" {
		if gui.Node.Balance[gui.Node.WalletAddress] < request.Stake {
			http.Error(w, "Saldo insuficiente", http.StatusBadRequest)
			return
		}

		// Deduzir do saldo e adicionar ao stake
		gui.Node.Balance[gui.Node.WalletAddress] -= request.Stake
		gui.Node.MiningStats.StakeAmount += request.Stake
		gui.Node.MiningStats.ValidatorLevel = "Validator"

		// Registrar no ledger
		gui.GlobalLedger.AddMiningReward(gui.Node.WalletAddress, request.Stake, "validator_stake")

		addLog(fmt.Sprintf("👑 Validator criado! Stake: %d tokens | Total: %d tokens | Saldo: %d",
			request.Stake, gui.Node.MiningStats.StakeAmount, gui.Node.Balance[gui.Node.WalletAddress]))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"stake_amount":    gui.Node.MiningStats.StakeAmount,
		"validator_level": gui.Node.MiningStats.ValidatorLevel,
	})
}

func handleTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data map[string]string
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	amountStr := data["amount"]
	to := data["to"]

	// Converter amount para int64
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	// Verificar se tem wallet ativa
	if gui.Node.WalletAddress == "" {
		http.Error(w, "Nenhuma wallet ativa", http.StatusBadRequest)
		return
	}

	// Verificar se tem saldo suficiente
	if gui.Node.Balance[gui.Node.WalletAddress] < amount {
		http.Error(w, "Saldo insuficiente", http.StatusBadRequest)
		return
	}

	// Processar transferência usando o ledger global
	from := gui.Node.WalletAddress
	err = gui.GlobalLedger.AddTransfer(from, to, amount, fmt.Sprintf("Transferência de %d tokens de %s para %s", amount, from[:10]+"...", to[:10]+"..."))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Atualizar saldos locais para sincronizar com o ledger
	gui.Node.Balance[from] = gui.GlobalLedger.GetBalance(from)
	gui.Node.Balance[to] = gui.GlobalLedger.GetBalance(to)

	addLog(fmt.Sprintf("💸 Transferência: %d tokens de %s para %s | Saldo: %d",
		amount, from[:10]+"...", to[:10]+"...", gui.Node.Balance[from]))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"from":        from,
		"to":          to,
		"amount":      amount,
		"new_balance": gui.Node.Balance[from],
	})
}

func handleCreateWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Gerar chaves únicas para a wallet
		publicKey := fmt.Sprintf("wallet_%s", fmt.Sprintf("%d", time.Now().UnixNano()))
		privateKey := fmt.Sprintf("private_%s", fmt.Sprintf("%d", time.Now().UnixNano()))

		// Criar autenticação segura para a wallet
		walletAuth, err := gui.UserManager.CreateWalletAuth(publicKey, privateKey)
		if err != nil {
			http.Error(w, "Erro ao criar wallet", http.StatusInternalServerError)
			return
		}

		// Criar wallet no manager
		wallet, err := gui.WalletManager.CreateWallet(publicKey, "default")
		if err == nil {
			gui.UserManager.AddWalletToUser(wallet.ID)
			// Usar a primeira conta como principal
			for _, account := range wallet.Accounts {
				gui.Node.Balance[account.Address] = account.Balance
				gui.Node.WalletAddress = account.Address
				addLog(fmt.Sprintf("✅ Nova wallet segura criada: %s | PIN: %s", account.Address, walletAuth.WalletPIN))
				break // Apenas a primeira conta
			}
		}

		// Retornar informações da wallet criada
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"public_key": publicKey,
			"pin":        walletAuth.WalletPIN,
			"message":    "Wallet criada com sucesso. Guarde o PIN com segurança!",
		})
	}
}

func handleWallets(w http.ResponseWriter, r *http.Request) {
	// Retornar todas as wallets do manager
	response := map[string]interface{}{
		"wallets": gui.WalletManager.Wallets,
		"count":   len(gui.WalletManager.Wallets),
	}
	json.NewEncoder(w).Encode(response)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(gui)
}

func handleLedger(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"logs": gui.Logs,
	}
	json.NewEncoder(w).Encode(response)
}

func getStatusClass() string {
	if gui.Node.IsRunning {
		return "running"
	}
	return "stopped"
}

func addLog(message string) {
	timestamp := time.Now().Format("15:04:05")
	logEntry := fmt.Sprintf("[%s] %s", timestamp, message)
	gui.Logs = append(gui.Logs, logEntry)

	if len(gui.Logs) > 100 {
		gui.Logs = gui.Logs[len(gui.Logs)-100:]
	}
}

func startMiningProcess() {
	if miningTicker != nil {
		return
	}

	miningTicker = time.NewTicker(2 * time.Second)
	miningStop = make(chan bool)

	go func() {
		// Usar o número total de blocos + 1 para continuar a contagem
		blockNumber := gui.Node.MiningStats.TotalBlocks + 1
		for {
			select {
			case <-miningTicker.C:
				if gui.Node.IsMining {
					reward := int64(50)
					gui.Node.MiningStats.TotalBlocks++
					gui.Node.MiningStats.TotalRewards += reward

					if gui.Node.WalletAddress != "" {
						gui.Node.Balance[gui.Node.WalletAddress] += reward
						gui.GlobalLedger.AddMiningReward(gui.Node.WalletAddress, reward, "mining_reward")
					}

					hash := fmt.Sprintf("block_%d_%d", blockNumber, time.Now().Unix())
					addLog(fmt.Sprintf("⛏️ Bloco #%d minerado! Hash: %s | Timestamp: %d | +%d tokens | Saldo: %d",
						blockNumber, hash, time.Now().Unix(), reward, gui.Node.Balance[gui.Node.WalletAddress]))

					// Salvar estado de mineração após cada bloco
					saveMiningState(int(gui.Node.MiningStats.TotalBlocks))

					blockNumber++
				}
			case <-miningStop:
				return
			}
		}
	}()
}

func stopMiningProcess() {
	if miningTicker != nil {
		miningTicker.Stop()
		miningTicker = nil
	}
	if miningStop != nil {
		// Verificar se o canal não foi fechado antes
		select {
		case <-miningStop:
			// Canal já foi fechado
		default:
			close(miningStop)
		}
		miningStop = nil
	}
}

func startRealTimeUpdates() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if gui.Node.IsMining {
				gui.Node.MiningStats.HashRate = float64(gui.Node.MiningStats.TotalBlocks) * 0.5
				gui.Node.MiningStats.EnergyCost = float64(gui.Node.MiningStats.TotalBlocks) * 0.1
				gui.Node.MiningStats.Profitability = float64(gui.Node.MiningStats.TotalRewards) - gui.Node.MiningStats.EnergyCost
			}
		}
	}
}

func main() {
	// Inicializar storage do Render
	renderStorage := storage.NewRenderStorage()
	if err := renderStorage.EnsureDirectories(); err != nil {
		log.Fatalf("Erro ao criar diretórios: %v", err)
	}

	// Inicializar gerenciador de usuários com storage persistente
	userManager := auth.NewUserManager(renderStorage.GetWalletsPath())

	// Inicializar wallet manager
	walletManager := wallet.NewWalletManager(renderStorage.GetWalletsPath())

	// Inicializar ledger global com storage persistente
	globalLedger := ledger.NewGlobalLedger(renderStorage.GetLedgerPath(), walletManager)

	// Inicializar seed nodes online (usado no setupRoutes)
	_ = network.NewOnlineSeedNodeManager()

	// Inicializar 2FA
	twoFactorAuth := NewTwoFactorAuth()

	// Configurar GUI
	gui = BlockchainGUI{
		Node: NodeInfo{
			Name:          "MinerNode",
			Port:          8080,
			Status:        "Iniciando",
			IsRunning:     false,
			IsMining:      false,
			MiningStats:   MiningStats{},
			Balance:       make(map[string]int64),
			Peers:         []string{},
			Difficulty:    1,
			EnergyPrice:   0.12,
			WalletName:    "miner_wallet",
			WalletAddress: "",
		},
		Logs:          []string{},
		IsRunning:     false,
		WalletManager: walletManager,
		GlobalLedger:  globalLedger,
		TwoFactorAuth: twoFactorAuth,
		UserManager:   userManager,
		RenderStorage: renderStorage,
	}

	// Carregar estado persistente
	if err := loadPersistentState(); err != nil {
		log.Printf("Aviso: Erro ao carregar estado persistente: %v", err)
	}

	// Configurar rotas
	setupRoutes()

	// Iniciar servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🔗 Blockchain 2-Layer - Node Online")
	log.Printf("📱 Interface disponível em: http://localhost:%s", port)
	log.Printf("💾 Storage persistente: %s", renderStorage.DataDir)

	http.ListenAndServe(":"+port, nil)
}

// loadPersistentState carrega estado persistente do Render
func loadPersistentState() error {
	// Carregar estado de mineração
	miningState, err := loadMiningState()
	if err != nil {
		return fmt.Errorf("erro ao carregar estado de mineração: %v", err)
	}
	gui.Node.MiningStats.TotalBlocks = int64(miningState)

	// Carregar ledger global
	if err := gui.GlobalLedger.LoadLedger(); err != nil {
		return fmt.Errorf("erro ao carregar ledger: %v", err)
	}

	// Carregar usuários
	if err := gui.UserManager.LoadUsers(); err != nil {
		return fmt.Errorf("erro ao carregar usuários: %v", err)
	}

	log.Printf("📊 Estado persistente carregado: %d blocos minerados", miningState)
	return nil
}



func setupRoutes() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/user-login", handleUserLogin)
	http.HandleFunc("/advanced-login", handleAdvancedLogin)
	http.HandleFunc("/user-logout", handleUserLogout)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/logout", handleLogout)
	http.HandleFunc("/generate-pin", handleGeneratePIN)
	http.HandleFunc("/start", handleStart)
	http.HandleFunc("/stop", handleStop)
	http.HandleFunc("/start-mining", handleStartMining)
	http.HandleFunc("/stop-mining", handleStopMining)
	http.HandleFunc("/stake", handleStake)
	http.HandleFunc("/become-validator", handleBecomeValidator)
	http.HandleFunc("/transfer", handleTransfer)
	http.HandleFunc("/create-wallet", handleCreateWallet)
	http.HandleFunc("/wallets", handleWallets)
	http.HandleFunc("/status", handleStatus)
	http.HandleFunc("/ledger", handleLedger)

	// Registrar endpoints da testnet
	testnetEndpoints := api.NewTestnetEndpoints()
	testnetEndpoints.RegisterTestnetEndpoints(http.DefaultServeMux)

	// Iniciar atualizações em tempo real
	go startRealTimeUpdates()
}
