package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// LogLevel representa o nível de log
type LogLevel string

const (
	DEBUG    LogLevel = "DEBUG"
	INFO     LogLevel = "INFO"
	WARNING  LogLevel = "WARNING"
	ERROR    LogLevel = "ERROR"
	CRITICAL LogLevel = "CRITICAL"
)

// LogEntry representa uma entrada de log estruturada
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Component string                 `json:"component"`
	NodeID    string                 `json:"node_id,omitempty"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	IP        string                 `json:"ip,omitempty"`
	Duration  int64                  `json:"duration_ms,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// Logger gerencia logs estruturados
type Logger struct {
	filePath     string
	logFile      *os.File
	encoder      *json.Encoder
	level        LogLevel
	component    string
	nodeID       string
	auditFile    *os.File
	auditEncoder *json.Encoder
}

// NewLogger cria um novo logger
func NewLogger(logDir, component, nodeID string, level LogLevel) (*Logger, error) {
	// Criar diretório se não existir
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de logs: %v", err)
	}

	// Arquivo de log principal
	logPath := filepath.Join(logDir, fmt.Sprintf("%s_%s.log", component, time.Now().Format("2006-01-02")))
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo de log: %v", err)
	}

	// Arquivo de auditoria
	auditPath := filepath.Join(logDir, fmt.Sprintf("audit_%s_%s.log", component, time.Now().Format("2006-01-02")))
	auditFile, err := os.OpenFile(auditPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logFile.Close()
		return nil, fmt.Errorf("erro ao abrir arquivo de auditoria: %v", err)
	}

	return &Logger{
		filePath:     logPath,
		logFile:      logFile,
		encoder:      json.NewEncoder(logFile),
		level:        level,
		component:    component,
		nodeID:       nodeID,
		auditFile:    auditFile,
		auditEncoder: json.NewEncoder(auditFile),
	}, nil
}

// log escreve uma entrada de log
func (l *Logger) log(level LogLevel, message string, data map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Component: l.component,
		NodeID:    l.nodeID,
		Message:   message,
		Data:      data,
	}

	// Escrever no arquivo de log
	l.encoder.Encode(entry)

	// Se for nível de auditoria, escrever também no arquivo de auditoria
	if level == WARNING || level == ERROR || level == CRITICAL {
		l.auditEncoder.Encode(entry)
	}

	// Log no console para níveis críticos
	if level == CRITICAL {
		fmt.Printf("[%s] %s: %s\n", level, l.component, message)
	}
}

// Debug log de debug
func (l *Logger) Debug(message string, data map[string]interface{}) {
	if l.level == DEBUG {
		l.log(DEBUG, message, data)
	}
}

// Info log de informação
func (l *Logger) Info(message string, data map[string]interface{}) {
	if l.level == DEBUG || l.level == INFO {
		l.log(INFO, message, data)
	}
}

// Warning log de aviso
func (l *Logger) Warning(message string, data map[string]interface{}) {
	if l.level == DEBUG || l.level == INFO || l.level == WARNING {
		l.log(WARNING, message, data)
	}
}

// Error log de erro
func (l *Logger) Error(message string, err error, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	if err != nil {
		data["error"] = err.Error()
	}

	if l.level == DEBUG || l.level == INFO || l.level == WARNING || l.level == ERROR {
		l.log(ERROR, message, data)
	}
}

// Critical log crítico
func (l *Logger) Critical(message string, err error, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	if err != nil {
		data["error"] = err.Error()
	}

	l.log(CRITICAL, message, data)
}

// Audit log de auditoria específico
func (l *Logger) Audit(action, userID, ip string, data map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     INFO,
		Component: l.component,
		NodeID:    l.nodeID,
		Message:   fmt.Sprintf("AUDIT: %s", action),
		Data:      data,
		UserID:    userID,
		IP:        ip,
	}

	l.auditEncoder.Encode(entry)
}

// Security log de segurança
func (l *Logger) Security(event, ip string, data map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     WARNING,
		Component: l.component,
		NodeID:    l.nodeID,
		Message:   fmt.Sprintf("SECURITY: %s", event),
		Data:      data,
		IP:        ip,
	}

	l.auditEncoder.Encode(entry)
	l.encoder.Encode(entry)
}

// Mining log de mineração
func (l *Logger) Mining(blockHash, minerID string, difficulty int, duration int64, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["block_hash"] = blockHash
	data["miner_id"] = minerID
	data["difficulty"] = difficulty
	data["duration_ms"] = duration

	l.log(INFO, "Block mined", data)
}

// Transaction log de transação
func (l *Logger) Transaction(txID, from, to string, amount int64, fee int64, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["tx_id"] = txID
	data["from"] = from
	data["to"] = to
	data["amount"] = amount
	data["fee"] = fee

	l.log(INFO, "Transaction processed", data)
}

// Network log de rede
func (l *Logger) Network(event, peerID string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["peer_id"] = peerID

	l.log(INFO, fmt.Sprintf("NETWORK: %s", event), data)
}

// Close fecha os arquivos de log
func (l *Logger) Close() error {
	if l.logFile != nil {
		l.logFile.Close()
	}
	if l.auditFile != nil {
		l.auditFile.Close()
	}
	return nil
}

// RotateLogs rotaciona os logs diariamente
func (l *Logger) RotateLogs() error {
	l.Close()

	logDir := filepath.Dir(l.filePath)
	component := l.component
	nodeID := l.nodeID
	level := l.level

	newLogger, err := NewLogger(logDir, component, nodeID, level)
	if err != nil {
		return err
	}

	*l = *newLogger
	return nil
}
