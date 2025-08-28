package security

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// IDSMonitor implementa sistema de monitoramento de segurança
type IDSMonitor struct {
	SuspiciousPatterns []*SecurityPattern
	AlertThreshold     int
	BlockedIPs         map[string]*BlockedIP
	AlertHistory       []*SecurityAlert
	mu                 sync.RWMutex
	AlertCallback      func(*SecurityAlert)
}

// SecurityPattern representa um padrão suspeito
type SecurityPattern struct {
	Name        string
	Pattern     string
	Regex       *regexp.Regexp
	Severity    string // low, medium, high, critical
	Description string
	Action      string // log, alert, block
}

// BlockedIP representa um IP bloqueado
type BlockedIP struct {
	IP           string
	Reason       string
	BlockedAt    time.Time
	BlockedUntil time.Time
	Attempts     int
	Severity     string
}

// SecurityAlert representa um alerta de segurança
type SecurityAlert struct {
	ID          string
	Timestamp   time.Time
	Type        string
	Severity    string
	IP          string
	UserAgent   string
	Description string
	Details     map[string]interface{}
	Action      string
	Resolved    bool
}

// IDSConfig configuração do monitor
type IDSConfig struct {
	AlertThreshold int           `json:"alert_threshold"`
	BlockDuration  time.Duration `json:"block_duration"`
	MaxBlockedIPs  int           `json:"max_blocked_ips"`
	EnableIPS      bool          `json:"enable_ips"`
	EnableIDS      bool          `json:"enable_ids"`
}

// NewIDSMonitor cria novo monitor de segurança
func NewIDSMonitor(config *IDSConfig) *IDSMonitor {
	if config == nil {
		config = &IDSConfig{
			AlertThreshold: 5,
			BlockDuration:  30 * time.Minute,
			MaxBlockedIPs:  1000,
			EnableIPS:      true,
			EnableIDS:      true,
		}
	}

	monitor := &IDSMonitor{
		AlertThreshold: config.AlertThreshold,
		BlockedIPs:     make(map[string]*BlockedIP),
		AlertHistory:   make([]*SecurityAlert, 0),
	}

	// Adicionar padrões suspeitos padrão
	monitor.addDefaultPatterns()

	return monitor
}

// addDefaultPatterns adiciona padrões suspeitos padrão
func (im *IDSMonitor) addDefaultPatterns() {
	patterns := []*SecurityPattern{
		{
			Name:        "SQL Injection",
			Pattern:     `(?i)(union|select|insert|update|delete|drop|create|exec|execute|script|javascript|vbscript|onload|onerror|onclick)`,
			Severity:    "high",
			Description: "Possível tentativa de SQL injection",
			Action:      "block",
		},
		{
			Name:        "XSS Attack",
			Pattern:     `(?i)(<script|javascript:|vbscript:|onload=|onerror=|onclick=|<iframe|<object)`,
			Severity:    "high",
			Description: "Possível tentativa de XSS",
			Action:      "block",
		},
		{
			Name:        "Path Traversal",
			Pattern:     `(\.\.\/|\.\.\\|%2e%2e%2f|%2e%2e%5c)`,
			Severity:    "medium",
			Description: "Possível tentativa de path traversal",
			Action:      "alert",
		},
		{
			Name:        "Command Injection",
			Pattern:     `(?i)(cmd|command|exec|system|eval|shell|bash|powershell)`,
			Severity:    "critical",
			Description: "Possível tentativa de command injection",
			Action:      "block",
		},
		{
			Name:        "File Upload",
			Pattern:     `(?i)\.(php|asp|aspx|jsp|exe|bat|cmd|sh|pl|py|rb)$`,
			Severity:    "medium",
			Description: "Tentativa de upload de arquivo suspeito",
			Action:      "alert",
		},
		{
			Name:        "Brute Force",
			Pattern:     `(login|auth|signin|admin)`,
			Severity:    "medium",
			Description: "Possível tentativa de brute force",
			Action:      "alert",
		},
	}

	for _, pattern := range patterns {
		pattern.Regex = regexp.MustCompile(pattern.Pattern)
		im.SuspiciousPatterns = append(im.SuspiciousPatterns, pattern)
	}
}

// AnalyzeRequest analisa requisição HTTP em busca de padrões suspeitos
func (im *IDSMonitor) AnalyzeRequest(r *http.Request) (bool, []*SecurityAlert) {
	alerts := []*SecurityAlert{}
	ip := extractIP(r)
	userAgent := r.UserAgent()

	// Verificar se IP está bloqueado
	if im.isIPBlocked(ip) {
		alert := &SecurityAlert{
			ID:          generateAlertID(),
			Timestamp:   time.Now(),
			Type:        "blocked_ip",
			Severity:    "high",
			IP:          maskIP(ip),
			UserAgent:   maskUserAgent(userAgent),
			Description: "Acesso bloqueado - IP em lista negra",
			Action:      "block",
			Resolved:    false,
		}
		alerts = append(alerts, alert)
		return false, alerts
	}

	// Analisar URL
	if urlAlerts := im.analyzeURL(r.URL.String(), ip, userAgent); len(urlAlerts) > 0 {
		alerts = append(alerts, urlAlerts...)
	}

	// Analisar headers
	if headerAlerts := im.analyzeHeaders(r.Header, ip, userAgent); len(headerAlerts) > 0 {
		alerts = append(alerts, headerAlerts...)
	}

	// Analisar User-Agent
	if uaAlerts := im.analyzeUserAgent(userAgent, ip); len(uaAlerts) > 0 {
		alerts = append(alerts, uaAlerts...)
	}

	// Verificar rate limiting
	if rateAlerts := im.checkRateLimit(ip); len(rateAlerts) > 0 {
		alerts = append(alerts, rateAlerts...)
	}

	// Processar alertas
	for _, alert := range alerts {
		im.processAlert(alert)
	}

	return len(alerts) == 0, alerts
}

// analyzeURL analisa URL em busca de padrões suspeitos
func (im *IDSMonitor) analyzeURL(url, ip, userAgent string) []*SecurityAlert {
	alerts := []*SecurityAlert{}

	for _, pattern := range im.SuspiciousPatterns {
		if pattern.Regex.MatchString(url) {
			alert := &SecurityAlert{
				ID:          generateAlertID(),
				Timestamp:   time.Now(),
				Type:        pattern.Name,
				Severity:    pattern.Severity,
				IP:          maskIP(ip),
				UserAgent:   maskUserAgent(userAgent),
				Description: pattern.Description,
				Details: map[string]interface{}{
					"pattern": pattern.Pattern,
					"url":     url,
				},
				Action:   pattern.Action,
				Resolved: false,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// analyzeHeaders analisa headers HTTP
func (im *IDSMonitor) analyzeHeaders(headers http.Header, ip, userAgent string) []*SecurityAlert {
	alerts := []*SecurityAlert{}

	// Verificar headers suspeitos
	suspiciousHeaders := map[string]string{
		"X-Forwarded-For":  "Possível spoofing de IP",
		"X-Real-IP":        "Possível spoofing de IP",
		"X-Originating-IP": "Possível spoofing de IP",
	}

	for header, description := range suspiciousHeaders {
		if value := headers.Get(header); value != "" {
			alert := &SecurityAlert{
				ID:          generateAlertID(),
				Timestamp:   time.Now(),
				Type:        "suspicious_header",
				Severity:    "medium",
				IP:          maskIP(ip),
				UserAgent:   maskUserAgent(userAgent),
				Description: description,
				Details: map[string]interface{}{
					"header": header,
					"value":  value,
				},
				Action:   "alert",
				Resolved: false,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// analyzeUserAgent analisa User-Agent
func (im *IDSMonitor) analyzeUserAgent(userAgent, ip string) []*SecurityAlert {
	alerts := []*SecurityAlert{}

	// Padrões suspeitos de User-Agent
	suspiciousUAs := []string{
		"sqlmap",
		"nikto",
		"nmap",
		"wget",
		"curl",
		"python-requests",
		"scanner",
		"bot",
		"crawler",
	}

	userAgentLower := strings.ToLower(userAgent)
	for _, suspicious := range suspiciousUAs {
		if strings.Contains(userAgentLower, suspicious) {
			alert := &SecurityAlert{
				ID:          generateAlertID(),
				Timestamp:   time.Now(),
				Type:        "suspicious_user_agent",
				Severity:    "medium",
				IP:          maskIP(ip),
				UserAgent:   maskUserAgent(userAgent),
				Description: fmt.Sprintf("User-Agent suspeito: %s", suspicious),
				Details: map[string]interface{}{
					"suspicious": suspicious,
					"user_agent": userAgent,
				},
				Action:   "alert",
				Resolved: false,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// checkRateLimit verifica rate limiting
func (im *IDSMonitor) checkRateLimit(ip string) []*SecurityAlert {
	alerts := []*SecurityAlert{}

	// Implementação simplificada de rate limiting
	// Em produção, usar sistema mais robusto
	im.mu.Lock()
	defer im.mu.Unlock()

	// Contar requisições recentes (últimos 5 minutos)
	recentAlerts := 0
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	for _, alert := range im.AlertHistory {
		if alert.IP == maskIP(ip) && alert.Timestamp.After(fiveMinutesAgo) {
			recentAlerts++
		}
	}

	if recentAlerts >= im.AlertThreshold {
		alert := &SecurityAlert{
			ID:          generateAlertID(),
			Timestamp:   time.Now(),
			Type:        "rate_limit_exceeded",
			Severity:    "high",
			IP:          maskIP(ip),
			Description: fmt.Sprintf("Rate limit excedido: %d alertas em 5 minutos", recentAlerts),
			Details: map[string]interface{}{
				"recent_alerts": recentAlerts,
				"threshold":     im.AlertThreshold,
			},
			Action:   "block",
			Resolved: false,
		}
		alerts = append(alerts, alert)

		// Bloquear IP
		im.blockIP(ip, "Rate limit excedido", 30*time.Minute)
	}

	return alerts
}

// processAlert processa alerta de segurança
func (im *IDSMonitor) processAlert(alert *SecurityAlert) {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Adicionar à história
	im.AlertHistory = append(im.AlertHistory, alert)

	// Manter apenas últimos 1000 alertas
	if len(im.AlertHistory) > 1000 {
		im.AlertHistory = im.AlertHistory[1:]
	}

	// Executar ação baseada no tipo de alerta
	switch alert.Action {
	case "block":
		im.blockIP(alert.IP, alert.Description, 30*time.Minute)
	case "alert":
		// Apenas logar
		fmt.Printf("🚨 ALERTA DE SEGURANÇA: %s - %s\n", alert.Type, alert.Description)
	}

	// Chamar callback se configurado
	if im.AlertCallback != nil {
		im.AlertCallback(alert)
	}
}

// blockIP bloqueia IP
func (im *IDSMonitor) blockIP(ip, reason string, duration time.Duration) {
	// Limpar IPs expirados primeiro
	im.cleanupExpiredBlocks()

	// Verificar limite de IPs bloqueados
	if len(im.BlockedIPs) >= 1000 {
		// Remover IP mais antigo
		var oldestIP string
		var oldestTime time.Time
		for blockedIP, blocked := range im.BlockedIPs {
			if oldestTime.IsZero() || blocked.BlockedAt.Before(oldestTime) {
				oldestTime = blocked.BlockedAt
				oldestIP = blockedIP
			}
		}
		if oldestIP != "" {
			delete(im.BlockedIPs, oldestIP)
		}
	}

	// Bloquear IP
	im.BlockedIPs[ip] = &BlockedIP{
		IP:           ip,
		Reason:       reason,
		BlockedAt:    time.Now(),
		BlockedUntil: time.Now().Add(duration),
		Attempts:     1,
		Severity:     "high",
	}

	fmt.Printf("🚫 IP bloqueado: %s - %s\n", maskIP(ip), reason)
}

// isIPBlocked verifica se IP está bloqueado
func (im *IDSMonitor) isIPBlocked(ip string) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	blocked, exists := im.BlockedIPs[ip]
	if !exists {
		return false
	}

	// Verificar se bloqueio expirou
	if time.Now().After(blocked.BlockedUntil) {
		delete(im.BlockedIPs, ip)
		return false
	}

	return true
}

// cleanupExpiredBlocks remove bloqueios expirados
func (im *IDSMonitor) cleanupExpiredBlocks() {
	now := time.Now()
	expiredIPs := []string{}

	for ip, blocked := range im.BlockedIPs {
		if now.After(blocked.BlockedUntil) {
			expiredIPs = append(expiredIPs, ip)
		}
	}

	for _, ip := range expiredIPs {
		delete(im.BlockedIPs, ip)
	}
}

// generateAlertID gera ID único para alerta
func generateAlertID() string {
	timestamp := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", timestamp)))
	return base64.URLEncoding.EncodeToString(hash[:8])
}

// GetSecurityStats retorna estatísticas de segurança
func (im *IDSMonitor) GetSecurityStats() map[string]interface{} {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Contar alertas por severidade
	severityCount := map[string]int{
		"low":      0,
		"medium":   0,
		"high":     0,
		"critical": 0,
	}

	for _, alert := range im.AlertHistory {
		severityCount[alert.Severity]++
	}

	return map[string]interface{}{
		"total_alerts":    len(im.AlertHistory),
		"blocked_ips":     len(im.BlockedIPs),
		"severity_count":  severityCount,
		"patterns_count":  len(im.SuspiciousPatterns),
		"alert_threshold": im.AlertThreshold,
	}
}

// SetAlertCallback define callback para alertas
func (im *IDSMonitor) SetAlertCallback(callback func(*SecurityAlert)) {
	im.AlertCallback = callback
}
