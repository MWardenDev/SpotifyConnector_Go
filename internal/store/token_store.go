package store

import "sync"

type Token struct {
	AccessToken string
	RefreshToken	string
	TokenType	string
	ExpiresIn 	int
	Scope	string
}

type TokenStore interface {
	Put(sessionID string, t Token)
	Get(sessionID string) (Token, bool)
	Delete(sessionID string)
}

type MemoryTokenStore struct {
	mu sync.RWMutex
	m map[string]Token
}

func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{m: make(map[string]Token)}
}

func (s *MemoryTokenStore) Put(sessionID string, t Token) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[sessionID] = t
}

func (s *MemoryTokenStore) Get(sessionID string) (Token, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.m[sessionID]
	return t, ok
}

func (s *MemoryTokenStore) Delete(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, sessionID)
}
