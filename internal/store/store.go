// Package store keeps calculations in memory.
package store

import (
	"errors"
	"slices"
	"sync"
	"time"
)

var ErrNotFound = errors.New("calculation not found")

// Calculation is one evaluated operation, the resource served under /calculations.
type Calculation struct {
	ID        int64     `json:"id"`
	Operation string    `json:"operation"`
	Operands  []float64 `json:"operands"`
	AngleUnit string    `json:"angle_unit,omitempty"`
	Result    float64   `json:"result"`
	CreatedAt time.Time `json:"created_at"`
}

// Memory is a concurrency-safe in-memory calculation store.
type Memory struct {
	mu     sync.RWMutex
	nextID int64
	items  map[int64]Calculation
}

func NewMemory() *Memory {
	return &Memory{items: make(map[int64]Calculation)}
}

// Create assigns an ID and creation time to c and stores it.
func (m *Memory) Create(c Calculation) Calculation {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	c.ID = m.nextID
	c.Operands = slices.Clone(c.Operands)
	c.CreatedAt = time.Now().UTC()
	m.items[c.ID] = c
	return c
}

func (m *Memory) Get(id int64) (Calculation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.items[id]
	if !ok {
		return Calculation{}, ErrNotFound
	}
	return c, nil
}

// List returns every calculation, oldest first.
func (m *Memory) List() []Calculation {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]Calculation, 0, len(m.items))
	for _, c := range m.items {
		list = append(list, c)
	}
	slices.SortFunc(list, func(a, b Calculation) int { return int(a.ID - b.ID) })
	return list
}

func (m *Memory) Delete(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[id]; !ok {
		return ErrNotFound
	}
	delete(m.items, id)
	return nil
}
