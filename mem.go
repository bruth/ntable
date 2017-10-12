package ntable

import "sync"

type memStore struct {
	store map[string][]byte
	mux   sync.RWMutex
}

func (s *memStore) Set(key, val []byte) error {
	s.mux.Lock()
	s.store[string(key)] = val
	s.mux.Unlock()
	return nil
}

func (s *memStore) Get(key []byte) ([]byte, error) {
	s.mux.RLock()
	val, ok := s.store[string(key)]
	s.mux.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}
	return val, nil
}

func (s *memStore) Del(key []byte) error {
	s.mux.Lock()
	if _, ok := s.store[string(key)]; !ok {
		return ErrNotFound
	}
	delete(s.store, string(key))
	s.mux.Unlock()
	return nil
}

// NewMemStore initialize an in-memory store.
func NewMemStore() Store {
	return &memStore{
		store: make(map[string][]byte),
	}
}
