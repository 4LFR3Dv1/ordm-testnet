package auth

import (
	"log"
	"sync"
	"time"
)

type RateLimiter struct {
	attempts      map[string][]time.Time
	mu            sync.RWMutex
	maxAttempts   int
	window        time.Duration
	lockoutTime   time.Duration
	lockouts      map[string]time.Time
	logSuspicious bool
	suspiciousIPs map[string]int
}

type AttemptResult struct {
	Allowed     bool
	Remaining   int
	Locked      bool
	LockedUntil time.Time
}

func NewRateLimiter(maxAttempts int, window, lockoutTime time.Duration) *RateLimiter {
	return &RateLimiter{
		attempts:      make(map[string][]time.Time),
		maxAttempts:   maxAttempts,
		window:        window,
		lockoutTime:   lockoutTime,
		lockouts:      make(map[string]time.Time),
		logSuspicious: true,
		suspiciousIPs: make(map[string]int),
	}
}

func (rl *RateLimiter) CheckRateLimit(identifier string) AttemptResult {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	if lockedUntil, exists := rl.lockouts[identifier]; exists {
		if now.Before(lockedUntil) {
			return AttemptResult{
				Allowed:     false,
				Remaining:   0,
				Locked:      true,
				LockedUntil: lockedUntil,
			}
		} else {
			delete(rl.lockouts, identifier)
		}
	}

	cutoff := now.Add(-rl.window)
	var validAttempts []time.Time
	for _, attempt := range rl.attempts[identifier] {
		if attempt.After(cutoff) {
			validAttempts = append(validAttempts, attempt)
		}
	}
	rl.attempts[identifier] = validAttempts

	remaining := rl.maxAttempts - len(validAttempts)
	allowed := remaining > 0

	return AttemptResult{
		Allowed:   allowed,
		Remaining: remaining,
		Locked:    false,
	}
}

func (rl *RateLimiter) RecordAttempt(identifier string, success bool) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	if success {
		delete(rl.attempts, identifier)
		delete(rl.lockouts, identifier)
		return
	}

	rl.attempts[identifier] = append(rl.attempts[identifier], now)

	// Log tentativas suspeitas
	if rl.logSuspicious {
		rl.suspiciousIPs[identifier]++
		if rl.suspiciousIPs[identifier] >= 2 {
			log.Printf("🚨 Tentativa suspeita detectada: IP %s, tentativas: %d",
				identifier, rl.suspiciousIPs[identifier])
		}
	}

	if len(rl.attempts[identifier]) >= rl.maxAttempts {
		lockoutUntil := now.Add(rl.lockoutTime)
		rl.lockouts[identifier] = lockoutUntil
		log.Printf("🔒 IP %s bloqueado por %v devido a múltiplas tentativas",
			identifier, rl.lockoutTime)
	}
}

// NewSecureRateLimiter cria rate limiter com configurações seguras padrão
func NewSecureRateLimiter() *RateLimiter {
	return NewRateLimiter(
		3,             // 3 tentativas
		1*time.Hour,   // Janela de 1 hora
		5*time.Minute, // Lockout de 5 minutos
	)
}

// GetSuspiciousIPs retorna IPs com tentativas suspeitas
func (rl *RateLimiter) GetSuspiciousIPs() map[string]int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	result := make(map[string]int)
	for ip, count := range rl.suspiciousIPs {
		result[ip] = count
	}
	return result
}

// CleanupOldEntries remove entradas antigas
func (rl *RateLimiter) CleanupOldEntries() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Limpar tentativas antigas
	for identifier, attempts := range rl.attempts {
		var valid []time.Time
		for _, attempt := range attempts {
			if attempt.After(cutoff) {
				valid = append(valid, attempt)
			}
		}
		if len(valid) == 0 {
			delete(rl.attempts, identifier)
		} else {
			rl.attempts[identifier] = valid
		}
	}

	// Limpar lockouts expirados
	for identifier, lockoutTime := range rl.lockouts {
		if now.After(lockoutTime) {
			delete(rl.lockouts, identifier)
		}
	}
}
