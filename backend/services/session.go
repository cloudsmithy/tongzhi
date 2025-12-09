package services

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

// Session represents a user session
type Session struct {
	ID        string
	UserID    string
	Email     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionManager manages user sessions
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	ttl      time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(ttl time.Duration) *SessionManager {
	if ttl == 0 {
		ttl = 24 * time.Hour // Default 24 hours
	}
	return &SessionManager{
		sessions: make(map[string]*Session),
		ttl:      ttl,
	}
}

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(userID, email string) (*Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		Email:     email,
		CreatedAt: now,
		ExpiresAt: now.Add(sm.ttl),
	}

	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.mu.Unlock()

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) *Session {
	sm.mu.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return nil
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		sm.DeleteSession(sessionID)
		return nil
	}

	return session
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mu.Lock()
	delete(sm.sessions, sessionID)
	sm.mu.Unlock()
}

// ValidateSession checks if a session is valid
func (sm *SessionManager) ValidateSession(sessionID string) bool {
	return sm.GetSession(sessionID) != nil
}

// generateSessionID generates a cryptographically secure session ID
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateState generates a random state for OIDC flow
func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
