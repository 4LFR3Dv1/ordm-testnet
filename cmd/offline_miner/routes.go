package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// setupRoutes configura as rotas HTTP do minerador offline
func setupRoutes() {
	http.HandleFunc("/", handleOfflineDashboard)
	http.HandleFunc("/api/status", handleStatus)
	http.HandleFunc("/api/start-mining", handleStartMining)
	http.HandleFunc("/api/stop-mining", handleStopMining)
	http.HandleFunc("/api/mine-block", handleMineBlock)
	http.HandleFunc("/api/blocks", handleGetBlocks)
	http.HandleFunc("/api/sync", handleSync)
	http.HandleFunc("/api/stats", handleStats)
}

// handleOfflineDashboard serve o dashboard offline
func handleOfflineDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>🏭 Minerador Offline</title>
    <style>
        body {
            font-family: 'Courier New', monospace;
            background: #0a0a0a;
            color: #00ff00;
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
            border-bottom: 2px solid #00ff00;
            padding-bottom: 20px;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .stat-card {
            background: #1a1a1a;
            border: 1px solid #00ff00;
            border-radius: 10px;
            padding: 20px;
            text-align: center;
        }
        .stat-value {
            font-size: 24px;
            font-weight: bold;
            color: #00ff00;
        }
        .stat-label {
            color: #888;
            margin-top: 5px;
        }
        .controls {
            display: flex;
            gap: 10px;
            margin-bottom: 30px;
            justify-content: center;
        }
        .btn {
            padding: 12px 24px;
            background: #00ff00;
            color: #000;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-family: 'Courier New', monospace;
            font-weight: bold;
        }
        .btn:hover {
            background: #00cc00;
        }
        .btn:disabled {
            background: #333;
            color: #666;
            cursor: not-allowed;
        }
        .logs {
            background: #1a1a1a;
            border: 1px solid #00ff00;
            border-radius: 10px;
            padding: 20px;
            height: 300px;
            overflow-y: auto;
            font-family: 'Courier New', monospace;
            font-size: 12px;
        }
        .log-entry {
            margin-bottom: 5px;
            padding: 2px 0;
        }
        .status-connected { color: #00ff00; }
        .status-disconnected { color: #ff4444; }
        .status-syncing { color: #ffaa00; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🏭 Minerador Offline</h1>
            <p>Mineração local independente da rede online</p>
        </div>

        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-value" id="totalBlocks">0</div>
                <div class="stat-label">Blocos Minerados</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="hashRate">0 H/s</div>
                <div class="stat-label">Hash Rate</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="uptime">00:00:00</div>
                <div class="stat-label">Uptime</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="syncStatus" class="status-disconnected">Desconectado</div>
                <div class="stat-label">Status Sincronização</div>
            </div>
        </div>

        <div class="controls">
            <button class="btn" onclick="startMining()">⛏️ Iniciar Mineração</button>
            <button class="btn" onclick="stopMining()">⏸️ Parar Mineração</button>
            <button class="btn" onclick="mineSingleBlock()">🔨 Minerar 1 Bloco</button>
            <button class="btn" onclick="syncBlocks()">🔄 Sincronizar</button>
        </div>

        <div class="logs" id="logs">
            <div class="log-entry">[Sistema] Minerador offline iniciado</div>
            <div class="log-entry">[Sistema] Aguardando comandos...</div>
        </div>
    </div>

    <script>
        let isMining = false;

        function addLog(message) {
            const logs = document.getElementById('logs');
            const timestamp = new Date().toLocaleTimeString();
            const logEntry = document.createElement('div');
            logEntry.className = 'log-entry';
            logEntry.textContent = "[" + timestamp + "] " + message;
            logs.appendChild(logEntry);
            logs.scrollTop = logs.scrollHeight;
        }

        function updateStats() {
            fetch('/api/stats')
                .then(response => response.json())
                .then(data => {
                    document.getElementById('totalBlocks').textContent = data.total_blocks;
                    document.getElementById('hashRate').textContent = data.hash_rate.toFixed(2) + ' H/s';
                    document.getElementById('uptime').textContent = formatUptime(data.uptime);
                    document.getElementById('syncStatus').textContent = data.sync_status;
                    document.getElementById('syncStatus').className = 'status-' + data.sync_status.toLowerCase();
                });
        }

        function formatUptime(seconds) {
            const hours = Math.floor(seconds / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            const secs = Math.floor(seconds % 60);
            return hours.toString().padStart(2, '0') + ":" + minutes.toString().padStart(2, '0') + ":" + secs.toString().padStart(2, '0');
        }

        function startMining() {
            if (isMining) return;
            
            fetch('/api/start-mining', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        isMining = true;
                        addLog('✅ Mineração iniciada');
                        updateStats();
                    } else {
                        addLog('❌ Erro ao iniciar mineração: ' + data.error);
                    }
                });
        }

        function stopMining() {
            if (!isMining) return;
            
            fetch('/api/stop-mining', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        isMining = false;
                        addLog('⏸️ Mineração parada');
                        updateStats();
                    }
                });
        }

        function mineSingleBlock() {
            addLog('🔨 Minerando bloco único...');
            fetch('/api/mine-block', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        addLog("✅ Bloco #" + data.block_number + " minerado! Hash: " + data.block_hash.substring(0, 16) + "...");
                        updateStats();
                    } else {
                        addLog('❌ Erro ao minerar bloco: ' + data.error);
                    }
                });
        }

        function syncBlocks() {
            addLog('🔄 Sincronizando blocos...');
            fetch('/api/sync', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        addLog("✅ " + data.synced_blocks + " blocos sincronizados");
                        updateStats();
                    } else {
                        addLog('❌ Erro na sincronização: ' + data.error);
                    }
                });
        }

        // Atualizar estatísticas a cada 5 segundos
        setInterval(updateStats, 5000);
        updateStats();
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleStatus retorna o status do minerador
func handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := GetMiningStats()
	
	response := map[string]interface{}{
		"miner_id":     offlineMiner.Identity.MinerID,
		"is_mining":    miningTicker != nil,
		"total_blocks": stats.TotalBlocks,
		"hash_rate":    stats.HashRate,
		"uptime":       stats.Uptime.Seconds(),
		"sync_status":  offlineMiner.SyncManager.Status,
		"difficulty":   offlineMiner.LocalChain.Difficulty,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStartMining inicia a mineração contínua
func handleStartMining(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if miningTicker != nil {
		http.Error(w, "Mineração já está ativa", http.StatusBadRequest)
		return
	}

	// Iniciar mineração contínua
	miningTicker = time.NewTicker(30 * time.Second) // Minerar a cada 30 segundos
	miningStop = make(chan bool)

	go func() {
		for {
			select {
			case <-miningTicker.C:
				if offlineMiner != nil {
					block, err := offlineMiner.LocalChain.MineNextBlock(offlineMiner.Identity.MinerID)
					if err != nil {
						fmt.Printf("❌ Erro na mineração: %v\n", err)
					} else {
						fmt.Printf("✅ Bloco #%d minerado: %s\n", block.Header.Number, block.GetBlockHashString()[:16])
					}
				}
			case <-miningStop:
				return
			}
		}
	}()

	response := map[string]interface{}{
		"success": true,
		"message": "Mineração iniciada",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStopMining para a mineração contínua
func handleStopMining(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if miningTicker != nil {
		miningTicker.Stop()
		miningTicker = nil
	}
	if miningStop != nil {
		close(miningStop)
		miningStop = nil
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Mineração parada",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMineBlock minera um bloco único
func handleMineBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	block, err := offlineMiner.LocalChain.MineNextBlock(offlineMiner.Identity.MinerID)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"block_number": block.Header.Number,
		"block_hash":   block.GetBlockHashString(),
		"miner_id":     block.Header.MinerID,
		"timestamp":    block.Header.Timestamp,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetBlocks retorna os blocos minerados
func handleGetBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	blocks := []map[string]interface{}{}
	
	offlineMiner.LocalChain.mu.RLock()
	for _, block := range offlineMiner.LocalChain.Blocks {
		blocks = append(blocks, block.GetBlockInfo())
	}
	offlineMiner.LocalChain.mu.RUnlock()

	response := map[string]interface{}{
		"blocks": blocks,
		"count":  len(blocks),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSync sincroniza blocos com a rede online
func handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Por enquanto, apenas simular sincronização
	// Na Fase 3, será implementada a sincronização real
	offlineMiner.SyncManager.Status = "syncing"
	
	// Simular delay de sincronização
	time.Sleep(2 * time.Second)
	
	offlineMiner.SyncManager.Status = "connected"
	offlineMiner.SyncManager.LastSync = time.Now()

	response := map[string]interface{}{
		"success":       true,
		"synced_blocks": len(offlineMiner.LocalChain.Blocks),
		"status":        "connected",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStats retorna estatísticas detalhadas
func handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := GetMiningStats()
	
	response := map[string]interface{}{
		"total_blocks":   stats.TotalBlocks,
		"valid_blocks":   stats.ValidBlocks,
		"invalid_blocks": stats.InvalidBlocks,
		"hash_rate":      stats.HashRate,
		"uptime":         stats.Uptime.Seconds(),
		"sync_status":    offlineMiner.SyncManager.Status,
		"miner_id":       offlineMiner.Identity.MinerID,
		"difficulty":     offlineMiner.LocalChain.Difficulty,
		"last_sync":      offlineMiner.SyncManager.LastSync,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
