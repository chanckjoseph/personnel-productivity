package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SessionManager handles session persistence and retrieval
type SessionManager struct {
	sessionsDir string
	mu          sync.RWMutex
	activeSessions map[string]*SessionContext
}

// NewSessionManager creates and initializes a new SessionManager
func NewSessionManager(workDir string) (*SessionManager, error) {
	sessionsDir := filepath.Join(workDir, ".devtools-mcp", "sessions")
	
	// Create sessions directory
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
	}

	return &SessionManager{
		sessionsDir: sessionsDir,
		activeSessions: make(map[string]*SessionContext),
	}, nil
}

// CreateSession creates a new debugging session
func (sm *SessionManager) CreateSession(bugDescription string) (*SessionContext, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	
	session := &SessionContext{
		ID:              sessionID,
		StartTime:       time.Now(),
		BugDescription:  bugDescription,
		Environment:     make(map[string]string),
		Requests:        make(map[string]*Request),
		Results:         make(map[string]*Result),
		Hypotheses:      []Hypothesis{},
		Experiments:     []Experiment{},
		IterationCount:  0,
		BugFixed:        false,
		Metadata:        make(map[string]interface{}),
	}

	sm.activeSessions[sessionID] = session

	// Persist to disk immediately
	if err := sm.saveSessionToDisk(session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}

// GetSession retrieves an existing session by ID
func (sm *SessionManager) GetSession(sessionID string) (*SessionContext, error) {
	sm.mu.RLock()

	// Check if session is in memory
	if session, ok := sm.activeSessions[sessionID]; ok {
		sm.mu.RUnlock()
		return session, nil
	}
	sm.mu.RUnlock()

	// Try to load from disk
	session, err := sm.loadSessionFromDisk(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Cache in memory
	sm.mu.Lock()
	sm.activeSessions[sessionID] = session
	sm.mu.Unlock()

	return session, nil
}

// UpdateSession updates an existing session
func (sm *SessionManager) UpdateSession(session *SessionContext) error {
	if session.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	sm.mu.Lock()
	sm.activeSessions[session.ID] = session
	sm.mu.Unlock()

	// Persist to disk
	if err := sm.saveSessionToDisk(session); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// CloseSession marks a session as completed
func (sm *SessionManager) CloseSession(sessionID string) error {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return err
	}

	now := time.Now()
	session.EndTime = &now

	return sm.UpdateSession(session)
}

// ListSessions returns all available session IDs
func (sm *SessionManager) ListSessions() ([]string, error) {
	entries, err := ioutil.ReadDir(sm.sessionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var sessionIDs []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			// Remove .json extension
			sessionID := entry.Name()[:len(entry.Name())-5]
			sessionIDs = append(sessionIDs, sessionID)
		}
	}

	return sessionIDs, nil
}

// saveSessionToDisk persists a session to JSON file
func (sm *SessionManager) saveSessionToDisk(session *SessionContext) error {
	if session.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	filePath := filepath.Join(sm.sessionsDir, fmt.Sprintf("%s.json", session.ID))

	// Marshal to JSON
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Write to file (atomic: write to temp file first, then rename)
	tempFile := filePath + ".tmp"
	if err := ioutil.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	if err := os.Rename(tempFile, filePath); err != nil {
		_ = os.Remove(tempFile) // Clean up temp file on error
		return fmt.Errorf("failed to finalize session file: %w", err)
	}

	return nil
}

// loadSessionFromDisk loads a session from JSON file
func (sm *SessionManager) loadSessionFromDisk(sessionID string) (*SessionContext, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}

	filePath := filepath.Join(sm.sessionsDir, fmt.Sprintf("%s.json", sessionID))

	// Read file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	// Unmarshal from JSON
	var session SessionContext
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// DeleteSession removes a session (soft delete: just remove from active, file remains)
func (sm *SessionManager) DeleteSession(sessionID string) error {
	sm.mu.Lock()
	delete(sm.activeSessions, sessionID)
	sm.mu.Unlock()

	// Note: Session file is NOT deleted, only removed from active cache
	// This maintains audit trail
	return nil
}
