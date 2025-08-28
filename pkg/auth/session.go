package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"ordm-main/pkg/config"
)

type Session struct {
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (sm *SessionManager) CreateSession(userID, ip, userAgent string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	session := &Session{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(config.AppConfig.Auth.SessionTTL),
		IP:        ip,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	}

	sm.sessions[token] = session
	return session, nil
}

func (sm *SessionManager) ValidateSession(token string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[token]
	if !exists {
		return nil, false
	}

	if time.Now().After(session.ExpiresAt) {
		sm.mu.RUnlock()
		sm.mu.Lock()
		delete(sm.sessions, token)
		sm.mu.Unlock()
		sm.mu.RLock()
		return nil, false
	}

	return session, true
}

func (sm *SessionManager) InvalidateSession(token string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, token)
}
