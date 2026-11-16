package store

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Memory is a thread-safe, in-memory Store. Data is lost when the process
// exits. It exists so the app runs end to end before the SQLite store lands,
// and it doubles as a fast fake in tests.
type Memory struct {
	mu           sync.RWMutex
	now          func() time.Time
	lastClientID int64
	lastTicketID int64
	clients      map[int64]Client
	tickets      map[int64]Ticket
}

var _ Store = (*Memory)(nil)

// NewMemory returns an empty in-memory store.
func NewMemory() *Memory {
	return &Memory{
		now:     func() time.Time { return time.Now().UTC() },
		clients: make(map[int64]Client),
		tickets: make(map[int64]Ticket),
	}
}

// CreateClient implements Store.
func (m *Memory) CreateClient(_ context.Context, in NewClient) (Client, error) {
	if err := in.Validate(); err != nil {
		return Client{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastClientID++
	c := Client{ID: m.lastClientID, Name: in.Name, Notes: in.Notes, CreatedAt: m.now()}
	m.clients[c.ID] = c
	return c, nil
}

// ListClients implements Store. Clients are sorted by name.
func (m *Memory) ListClients(_ context.Context) ([]Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Client, 0, len(m.clients))
	for _, c := range m.clients {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// GetClient implements Store.
func (m *Memory) GetClient(_ context.Context, id int64) (Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.clients[id]
	if !ok {
		return Client{}, fmt.Errorf("client %d: %w", id, ErrNotFound)
	}
	return c, nil
}

// CreateTicket implements Store.
func (m *Memory) CreateTicket(_ context.Context, in NewTicket) (Ticket, error) {
	if err := in.Validate(); err != nil {
		return Ticket{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.clients[in.ClientID]; !ok {
		return Ticket{}, errors.Join(ErrInvalid, fmt.Errorf("client %d does not exist", in.ClientID))
	}
	now := m.now()
	m.lastTicketID++
	t := Ticket{
		ID:        m.lastTicketID,
		ClientID:  in.ClientID,
		Title:     in.Title,
		Body:      in.Body,
		Status:    StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.tickets[t.ID] = t
	return t, nil
}

// ListTickets implements Store. Newest tickets come first.
func (m *Memory) ListTickets(_ context.Context) ([]Ticket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Ticket, 0, len(m.tickets))
	for _, t := range m.tickets {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out, nil
}

// GetTicket implements Store.
func (m *Memory) GetTicket(_ context.Context, id int64) (Ticket, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tickets[id]
	if !ok {
		return Ticket{}, fmt.Errorf("ticket %d: %w", id, ErrNotFound)
	}
	return t, nil
}

// SetTicketStatus implements Store.
func (m *Memory) SetTicketStatus(_ context.Context, id int64, s Status) (Ticket, error) {
	if !s.Valid() {
		return Ticket{}, errors.Join(ErrInvalid, fmt.Errorf("unknown status %q", s))
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tickets[id]
	if !ok {
		return Ticket{}, fmt.Errorf("ticket %d: %w", id, ErrNotFound)
	}
	t.Status = s
	t.UpdatedAt = m.now()
	m.tickets[id] = t
	return t, nil
}
