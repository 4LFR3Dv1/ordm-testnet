package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"ordm-main/pkg/security"
)

var monitor *security.SecurityMonitor

func main() {
	fmt.Println("📊 Iniciando Dashboard de Monitoramento da Testnet")
	fmt.Println("==================================================")

	// Inicializar monitor de segurança
	config := security.MonitorConfig{
		LogFilePath:      "./logs/testnet_security.log",
		MaxEvents:        10000,
		MaxMetrics:       1000,
		AlertThreshold:   100,
		RateLimitWindow:  time.Hour,
		BlockDuration:    24 * time.Hour,
		BackupInterval:   time.Hour,
		PerformanceCheck: 5 * time.Minute,
		EnableAlerts:     true,
		EnableMetrics:    true,
		EnableRateLimit:  true,
	}

	monitor = security.NewSecurityMonitor(config)

	// Configurar rotas
	http.HandleFunc("/", handleDashboard)
	http.HandleFunc("/api/metrics", handleMetrics)
	http.HandleFunc("/api/security", handleSecurity)
	http.HandleFunc("/api/alerts", handleAlerts)
	http.HandleFunc("/api/events", handleEvents)

	// Iniciar servidor
	port := ":9090"
	fmt.Printf("📊 Dashboard disponível em: http://localhost%s\n", port)
	fmt.Printf("📈 Métricas em tempo real\n")
	fmt.Printf("🔒 Monitoramento de segurança ativo\n")
	fmt.Printf("💾 Backup automático configurado\n")

	log.Fatal(http.ListenAndServe(port, nil))
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "cmd/monitor/dashboard.html")
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Simular métricas (em produção, usar bibliotecas reais)
	metrics := map[string]interface{}{
		"cpu_usage":    float64(time.Now().UnixNano() % 100),
		"memory_usage": float64(time.Now().UnixNano() % 80),
		"disk_usage":   float64(time.Now().UnixNano() % 60),
		"network_io":   float64(time.Now().UnixNano() % 50),
		"connections":  int(time.Now().UnixNano() % 100),
		"request_rate": float64(time.Now().UnixNano()%1000) / 100.0,
		"error_rate":   float64(time.Now().UnixNano() % 10),
		"timestamp":    time.Now(),
	}

	json.NewEncoder(w).Encode(metrics)
}

func handleSecurity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	report := monitor.GetSecurityReport()
	json.NewEncoder(w).Encode(report)
}

func handleAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	report := monitor.GetSecurityReport()
	alerts := map[string]interface{}{
		"alerts": report["recent_alerts"],
		"total":  report["total_alerts"],
	}

	json.NewEncoder(w).Encode(alerts)
}

func handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	report := monitor.GetSecurityReport()
	events := map[string]interface{}{
		"recent_events": report["recent_events"],
		"total_events":  report["total_events"],
	}

	json.NewEncoder(w).Encode(events)
}
