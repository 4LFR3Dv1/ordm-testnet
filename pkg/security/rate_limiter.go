package security

import (
	"sync"
	"time"
)

// RateLimiter implementa rate limiting por IP/Node
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string][]time.Time
	window   time.Duration
	limit    int
}

// NewRateLimiter cria um novo rate limiter
func NewRateLimiter(window time.Duration, limit int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		window:   window,
		limit:    limit,
	}
}

// Allow verifica se uma requisição é permitida
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Limpar requisições antigas
	if times, exists := rl.requests[key]; exists {
		var valid []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		rl.requests[key] = valid
	}

	// Verificar limite
	if len(rl.requests[key]) >= rl.limit {
		return false
	}

	// Adicionar nova requisição
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

// GetRemaining retorna quantas requisições restam
func (rl *RateLimiter) GetRemaining(key string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	if times, exists := rl.requests[key]; exists {
		count := 0
		for _, t := range times {
			if t.After(windowStart) {
				count++
			}
		}
		return rl.limit - count
	}

	return rl.limit
}

// Cleanup remove entradas antigas
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	for key, times := range rl.requests {
		var valid []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = valid
		}
	}
}

// SecurityManager gerencia múltiplos rate limiters
type SecurityManager struct {
	apiLimiter        *RateLimiter
	miningLimiter     *RateLimiter
	connectionLimiter *RateLimiter
	blacklist         map[string]time.Time
	whitelist         map[string]bool
	mu                sync.RWMutex
}

// NewSecurityManager cria um novo gerenciador de segurança
func NewSecurityManager() *SecurityManager {
	return &SecurityManager{
		apiLimiter:        NewRateLimiter(1*time.Minute, 100), // 100 req/min
		miningLimiter:     NewRateLimiter(1*time.Second, 10),  // 10 req/sec
		connectionLimiter: NewRateLimiter(1*time.Minute, 50),  // 50 conexões/min
		blacklist:         make(map[string]time.Time),
		whitelist:         make(map[string]bool),
	}
}

// AllowAPI verifica se requisição API é permitida
func (sm *SecurityManager) AllowAPI(key string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Verificar whitelist
	if sm.whitelist[key] {
		return true
	}

	// Verificar blacklist
	if banTime, banned := sm.blacklist[key]; banned {
		if time.Now().Before(banTime) {
			return false
		}
		// Remover da blacklist se expirou
		delete(sm.blacklist, key)
	}

	return sm.apiLimiter.Allow(key)
}

// AllowMining verifica se mineração é permitida
func (sm *SecurityManager) AllowMining(key string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.whitelist[key] {
		return true
	}

	if banTime, banned := sm.blacklist[key]; banned {
		if time.Now().Before(banTime) {
			return false
		}
		delete(sm.blacklist, key)
	}

	return sm.miningLimiter.Allow(key)
}

// AllowConnection verifica se conexão é permitida
func (sm *SecurityManager) AllowConnection(key string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.whitelist[key] {
		return true
	}

	if banTime, banned := sm.blacklist[key]; banned {
		if time.Now().Before(banTime) {
			return false
		}
		delete(sm.blacklist, key)
	}

	return sm.connectionLimiter.Allow(key)
}

// BanIP bane um IP por um período
func (sm *SecurityManager) BanIP(ip string, duration time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.blacklist[ip] = time.Now().Add(duration)
}

// WhitelistIP adiciona IP à whitelist
func (sm *SecurityManager) WhitelistIP(ip string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.whitelist[ip] = true
}

// RemoveFromWhitelist remove IP da whitelist
func (sm *SecurityManager) RemoveFromWhitelist(ip string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.whitelist, ip)
}

// IsBanned verifica se IP está banido
func (sm *SecurityManager) IsBanned(ip string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if banTime, banned := sm.blacklist[ip]; banned {
		if time.Now().Before(banTime) {
			return true
		}
		// Remover se expirou
		delete(sm.blacklist, ip)
	}
	return false
}

// Cleanup limpa entradas expiradas
func (sm *SecurityManager) Cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for ip, banTime := range sm.blacklist {
		if now.After(banTime) {
			delete(sm.blacklist, ip)
		}
	}

	sm.apiLimiter.Cleanup()
	sm.miningLimiter.Cleanup()
	sm.connectionLimiter.Cleanup()
}
