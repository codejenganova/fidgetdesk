// Package store defines FidgetDesk's data model and the storage interface
// the rest of the app depends on. Implementations live alongside it
// (an in-memory store today; SQLite next).
package store

import (
	"context"
	"errors"
	"strings"
	"time"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("not found")

// ErrInvalid is returned when input fails validation.
var ErrInvalid = errors.New("invalid input")

// Status is the lifecycle state of a ticket.
type Status string

// Ticket statuses.
const (
	StatusOpen    Status = "open"
	StatusWaiting Status = "waiting"
	StatusClosed  Status = "closed"
)

// Valid reports whether s is a known status.
func (s Status) Valid() bool {
	switch s {
	case StatusOpen, StatusWaiting, StatusClosed:
		return true
	}
	return false
}

// Client is a place or person you help: a dentist's office, a neighbor
// in 4B, or your own homelab.
type Client struct {
	ID        int64
	Name      string
	Notes     string
	CreatedAt time.Time
}

// Ticket is a single piece of work for a client.
type Ticket struct {
	ID        int64
	ClientID  int64
	Title     string
	Body      string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewClient holds the fields needed to create a client.
type NewClient struct {
	Name  string
	Notes string
}

// Validate trims input and checks required fields.
func (n *NewClient) Validate() error {
	n.Name = strings.TrimSpace(n.Name)
	if n.Name == "" {
		return errors.Join(ErrInvalid, errors.New("client name is required"))
	}
	return nil
}

// NewTicket holds the fields needed to create a ticket.
type NewTicket struct {
	ClientID int64
	Title    string
	Body     string
}

// Validate trims input and checks required fields.
func (n *NewTicket) Validate() error {
	n.Title = strings.TrimSpace(n.Title)
	n.Body = strings.TrimSpace(n.Body)
	if n.Title == "" {
		return errors.Join(ErrInvalid, errors.New("ticket title is required"))
	}
	if n.ClientID <= 0 {
		return errors.Join(ErrInvalid, errors.New("ticket must belong to a client"))
	}
	return nil
}

// Store is the persistence boundary. Everything above it (HTTP handlers,
// email intake, future AI features) talks to this interface only, so the
// backing database can change without touching them.
type Store interface {
	CreateClient(ctx context.Context, in NewClient) (Client, error)
	ListClients(ctx context.Context) ([]Client, error)
	GetClient(ctx context.Context, id int64) (Client, error)

	CreateTicket(ctx context.Context, in NewTicket) (Ticket, error)
	ListTickets(ctx context.Context) ([]Ticket, error)
	GetTicket(ctx context.Context, id int64) (Ticket, error)
	SetTicketStatus(ctx context.Context, id int64, s Status) (Ticket, error)
}
