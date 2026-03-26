package store

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// Memory is an in-memory Store implementation. It is safe for concurrent
// use and is intended for tests and local development only. Data does not
// persist across process restarts.
type Memory struct {
	mu      sync.RWMutex
	entries map[CID][]byte
}

// NewMemory returns an initialised, empty in-memory store.
func NewMemory() *Memory {
	return &Memory{
		entries: make(map[CID][]byte),
	}
}

// Put stores data in memory. As with all Store implementations, it is
// idempotent: uploading the same bytes twice is a no-op.
func (m *Memory) Put(_ context.Context, data []byte) (CID, error) {
	cid := ComputeCID(data)

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.entries[cid]; !exists {
		// make a defensive copy so the caller cannot mutate stored content
		stored := make([]byte, len(data))
		copy(stored, data)
		m.entries[cid] = stored
	}
	return cid, nil
}

// Get retrieves the bytes stored under cid. Returns ErrNotFound if absent.
func (m *Memory) Get(_ context.Context, cid CID) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.entries[cid]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, cid)
	}
	// return a copy — callers must not mutate store internals
	out := make([]byte, len(data))
	copy(out, data)
	return out, nil
}

// GetWriter implements WriterTo for the memory store.
func (m *Memory) GetWriter(_ context.Context, cid CID, w io.Writer) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.entries[cid]
	if !ok {
		return fmt.Errorf("%w: %s", ErrNotFound, cid)
	}
	_, err := w.Write(data)
	return err
}

// Exists reports whether cid is present in the store.
func (m *Memory) Exists(_ context.Context, cid CID) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.entries[cid]
	return ok, nil
}

// List returns all CIDs currently held in the store.
func (m *Memory) List(_ context.Context) ([]CID, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cids := make([]CID, 0, len(m.entries))
	for cid := range m.entries {
		cids = append(cids, cid)
	}
	return cids, nil
}

// Delete removes the content for cid. Returns ErrNotFound if absent.
func (m *Memory) Delete(_ context.Context, cid CID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.entries[cid]; !ok {
		return fmt.Errorf("%w: %s", ErrNotFound, cid)
	}
	delete(m.entries, cid)
	return nil
}

// Len returns the number of entries currently in the store.
// Intended for tests and diagnostics.
func (m *Memory) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entries)
}
