package security

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// AlertLevel define o nível de alerta
type AlertLevel string

const (
	AlertInfo     AlertLevel = "INFO"
	AlertWarning  AlertLevel = "WARNING"
	AlertError    AlertLevel = "ERROR"
	AlertCritical AlertLevel = "CRITICAL"
)

// SecurityEvent representa um evento de segurança
type SecurityEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     AlertLevel             `json:"level"`
	Category  string                 `json:"category"`
	Message   string                 `json:"message"`
	IP        string                 `json:"ip,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Endpoint  string                 `json:"endpoint,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	EventID   string                 `json:"event_id"`
}

// PerformanceMetrics métricas de performance
type PerformanceMetrics struct {
	Timestamp         time.Time `json:"timestamp"`
	CPUUsage          float64   `json:"cpu_usage"`
	MemoryUsage       float64   `json:"memory_usage"`
	DiskUsage         float64   `json:"disk_usage"`
	NetworkIO         float64   `json:"network_io"`
	ActiveConnections int       `json:"active_connections"`
	RequestRate       float64   `json:"request_rate"`
	ErrorRate         float64   `json:"error_rate"`
}

// SecurityMonitor monitor de segurança
type SecurityMonitor struct {
	Events     []SecurityEvent      `json:"events"`
	Metrics    []PerformanceMetrics `json:"metrics"`
	Alerts     []SecurityEvent      `json:"alerts"`
	RateLimits map[string]int       `json:"rate_limits"`
	BlockedIPs map[string]time.Time `json:"blocked_ips"`
	Config     MonitorConfig        `json:"config"`
	mutex      sync.RWMutex
	logFile    *os.File
}

// MonitorConfig configuração do monitor
type MonitorConfig struct {
	LogFilePath      string        `json:"log_file_path"`
	MaxEvents        int           `json:"max_events"`
	MaxMetrics       int           `json:"max_metrics"`
	AlertThreshold   int           `json:"alert_threshold"`
	RateLimitWindow  time.Duration `json:"rate_limit_window"`
	BlockDuration    time.Duration `json:"block_duration"`
	BackupInterval   time.Duration `json:"backup_interval"`
	PerformanceCheck time.Duration `json:"performance_check"`
	EnableAlerts     bool          `json:"enable_alerts"`
	EnableMetrics    bool          `json:"enable_metrics"`
	EnableRateLimit  bool          `json:"enable_rate_limit"`
}

// NewSecurityMonitor cria um novo monitor de segurança
func NewSecurityMonitor(config MonitorConfig) *SecurityMonitor {
	if config.LogFilePath == "" {
		config.LogFilePath = "./logs/security.log"
	}
	if config.MaxEvents == 0 {
		config.MaxEvents = 10000
	}
	if config.MaxMetrics == 0 {
		config.MaxMetrics = 1000
	}
	if config.AlertThreshold == 0 {
		config.AlertThreshold = 100
	}
	if config.RateLimitWindow == 0 {
		config.RateLimitWindow = time.Hour
	}
	if config.BlockDuration == 0 {
		config.BlockDuration = 24 * time.Hour
	}
	if config.BackupInterval == 0 {
		config.BackupInterval = time.Hour
	}
	if config.PerformanceCheck == 0 {
		config.PerformanceCheck = 5 * time.Minute
	}

	// Criar diretório de logs se não existir
	os.MkdirAll("./logs", 0755)

	logFile, err := os.OpenFile(config.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("❌ Erro ao abrir arquivo de log: %v", err)
	}

	monitor := &SecurityMonitor{
		Events:     []SecurityEvent{},
		Metrics:    []PerformanceMetrics{},
		Alerts:     []SecurityEvent{},
		RateLimits: make(map[string]int),
		BlockedIPs: make(map[string]time.Time),
		Config:     config,
		logFile:    logFile,
	}

	// Iniciar serviços em background
	if config.EnableMetrics {
		go monitor.startPerformanceMonitoring()
	}
	if config.EnableAlerts {
		go monitor.startAlertMonitoring()
	}
	go monitor.startBackupService()

	return monitor
}

// LogEvent registra um evento de segurança
func (sm *SecurityMonitor) LogEvent(level AlertLevel, category, message, ip, userAgent, endpoint string, details map[string]interface{}) {
	event := SecurityEvent{
		Timestamp: time.Now(),
		Level:     level,
		Category:  category,
		Message:   message,
		IP:        ip,
		UserAgent: userAgent,
		Endpoint:  endpoint,
		Details:   details,
		EventID:   generateEventID(),
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Adicionar evento à lista
	sm.Events = append(sm.Events, event)
	if len(sm.Events) > sm.Config.MaxEvents {
		sm.Events = sm.Events[1:]
	}

	// Se for alerta, adicionar à lista de alertas
	if level == AlertWarning || level == AlertError || level == AlertCritical {
		sm.Alerts = append(sm.Alerts, event)
		if len(sm.Alerts) > sm.Config.AlertThreshold {
			sm.Alerts = sm.Alerts[1:]
		}
	}

	// Logar para arquivo
	sm.logToFile(event)

	// Logar para console
	log.Printf("🔒 [%s] %s: %s (IP: %s, Endpoint: %s)",
		level, category, message, ip, endpoint)
}

// CheckRateLimit verifica rate limit para um IP
func (sm *SecurityMonitor) CheckRateLimit(ip string) bool {
	if !sm.Config.EnableRateLimit {
		return true
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Verificar se IP está bloqueado
	if blockTime, exists := sm.BlockedIPs[ip]; exists {
		if time.Since(blockTime) < sm.Config.BlockDuration {
			sm.LogEvent(AlertWarning, "RATE_LIMIT",
				fmt.Sprintf("IP bloqueado tentou acessar: %s", ip),
				ip, "", "", nil)
			return false
		}
		// Remover do bloqueio se expirou
		delete(sm.BlockedIPs, ip)
	}

	// Incrementar contador
	sm.RateLimits[ip]++

	// Verificar se excedeu limite
	if sm.RateLimits[ip] > 100 { // 100 requests por janela
		sm.BlockedIPs[ip] = time.Now()
		sm.LogEvent(AlertWarning, "RATE_LIMIT",
			fmt.Sprintf("IP bloqueado por rate limit: %s", ip),
			ip, "", "", nil)
		return false
	}

	return true
}

// RecordMetric registra uma métrica de performance
func (sm *SecurityMonitor) RecordMetric(cpu, memory, disk, network float64, connections int, requestRate, errorRate float64) {
	if !sm.Config.EnableMetrics {
		return
	}

	metric := PerformanceMetrics{
		Timestamp:         time.Now(),
		CPUUsage:          cpu,
		MemoryUsage:       memory,
		DiskUsage:         disk,
		NetworkIO:         network,
		ActiveConnections: connections,
		RequestRate:       requestRate,
		ErrorRate:         errorRate,
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.Metrics = append(sm.Metrics, metric)
	if len(sm.Metrics) > sm.Config.MaxMetrics {
		sm.Metrics = sm.Metrics[1:]
	}
}

// GetSecurityReport gera relatório de segurança
func (sm *SecurityMonitor) GetSecurityReport() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Contar eventos por nível
	levelCounts := make(map[string]int)
	for _, event := range sm.Events {
		levelCounts[string(event.Level)]++
	}

	// Contar eventos por categoria
	categoryCounts := make(map[string]int)
	for _, event := range sm.Events {
		categoryCounts[event.Category]++
	}

	// Calcular métricas de performance médias
	var avgCPU, avgMemory, avgDisk, avgNetwork float64
	if len(sm.Metrics) > 0 {
		for _, metric := range sm.Metrics {
			avgCPU += metric.CPUUsage
			avgMemory += metric.MemoryUsage
			avgDisk += metric.DiskUsage
			avgNetwork += metric.NetworkIO
		}
		count := float64(len(sm.Metrics))
		avgCPU /= count
		avgMemory /= count
		avgDisk /= count
		avgNetwork /= count
	}

	return map[string]interface{}{
		"total_events":    len(sm.Events),
		"total_alerts":    len(sm.Alerts),
		"total_metrics":   len(sm.Metrics),
		"blocked_ips":     len(sm.BlockedIPs),
		"level_counts":    levelCounts,
		"category_counts": categoryCounts,
		"performance": map[string]interface{}{
			"avg_cpu":     avgCPU,
			"avg_memory":  avgMemory,
			"avg_disk":    avgDisk,
			"avg_network": avgNetwork,
		},
		"recent_events": sm.Events[max(0, len(sm.Events)-10):],
		"recent_alerts": sm.Alerts[max(0, len(sm.Alerts)-5):],
		"config":        sm.Config,
	}
}

// startPerformanceMonitoring inicia monitoramento de performance
func (sm *SecurityMonitor) startPerformanceMonitoring() {
	ticker := time.NewTicker(sm.Config.PerformanceCheck)
	defer ticker.Stop()

	for range ticker.C {
		// Simular métricas de performance (em produção, usar bibliotecas reais)
		cpu := float64(time.Now().UnixNano()%100) / 100.0
		memory := float64(time.Now().UnixNano()%80) / 100.0
		disk := float64(time.Now().UnixNano()%60) / 100.0
		network := float64(time.Now().UnixNano()%50) / 100.0
		connections := int(time.Now().UnixNano() % 100)
		requestRate := float64(time.Now().UnixNano()%1000) / 100.0
		errorRate := float64(time.Now().UnixNano()%10) / 100.0

		sm.RecordMetric(cpu, memory, disk, network, connections, requestRate, errorRate)

		// Alertar se métricas estão altas
		if cpu > 80 || memory > 80 || disk > 90 {
			sm.LogEvent(AlertWarning, "PERFORMANCE",
				fmt.Sprintf("Métricas altas - CPU: %.1f%%, Memory: %.1f%%, Disk: %.1f%%", cpu*100, memory*100, disk*100),
				"", "", "", nil)
		}
	}
}

// startAlertMonitoring inicia monitoramento de alertas
func (sm *SecurityMonitor) startAlertMonitoring() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.mutex.RLock()
		alertCount := len(sm.Alerts)
		sm.mutex.RUnlock()

		if alertCount > sm.Config.AlertThreshold {
			sm.LogEvent(AlertCritical, "SYSTEM",
				fmt.Sprintf("Muitos alertas ativos: %d", alertCount),
				"", "", "", nil)
		}
	}
}

// startBackupService inicia serviço de backup
func (sm *SecurityMonitor) startBackupService() {
	ticker := time.NewTicker(sm.Config.BackupInterval)
	defer ticker.Stop()

	for range ticker.C {
		sm.backupData()
	}
}

// backupData faz backup dos dados de segurança
func (sm *SecurityMonitor) backupData() {
	sm.mutex.RLock()
	data := map[string]interface{}{
		"events":      sm.Events,
		"metrics":     sm.Metrics,
		"alerts":      sm.Alerts,
		"rate_limits": sm.RateLimits,
		"blocked_ips": sm.BlockedIPs,
		"config":      sm.Config,
		"timestamp":   time.Now(),
	}
	sm.mutex.RUnlock()

	// Criar backup
	backupFile := fmt.Sprintf("./backups/security_backup_%s.json",
		time.Now().Format("2006-01-02_15-04-05"))

	os.MkdirAll("./backups", 0755)

	file, err := os.Create(backupFile)
	if err != nil {
		log.Printf("❌ Erro ao criar backup: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		log.Printf("❌ Erro ao codificar backup: %v", err)
		return
	}

	log.Printf("✅ Backup de segurança criado: %s", backupFile)
}

// logToFile loga evento para arquivo
func (sm *SecurityMonitor) logToFile(event SecurityEvent) {
	if sm.logFile == nil {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("❌ Erro ao codificar evento: %v", err)
		return
	}

	sm.logFile.Write(append(data, '\n'))
	sm.logFile.Sync()
}

// generateEventID gera ID único para evento
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

// max retorna o maior de dois inteiros
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
