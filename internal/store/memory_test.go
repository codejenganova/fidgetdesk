package store

import (
	"context"
	"errors"
	"testing"
)

func TestMemory_ClientLifecycle(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()

	c, err := m.CreateClient(ctx, NewClient{Name: "  Smile Dental  "})
	if err != nil {
		t.Fatalf("CreateClient: %v", err)
	}
	if c.Name != "Smile Dental" {
		t.Errorf("name not trimmed: got %q", c.Name)
	}

	got, err := m.GetClient(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetClient: %v", err)
	}
	if got != c {
		t.Errorf("GetClient = %+v, want %+v", got, c)
	}

	if _, err := m.GetClient(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetClient(missing) err = %v, want ErrNotFound", err)
	}
}

func TestMemory_ListClientsSortedByName(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	for _, n := range []string{"Zed's Garage", "Apt 4B", "Homelab"} {
		if _, err := m.CreateClient(ctx, NewClient{Name: n}); err != nil {
			t.Fatal(err)
		}
	}
	list, err := m.ListClients(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Apt 4B", "Homelab", "Zed's Garage"}
	for i, c := range list {
		if c.Name != want[i] {
			t.Errorf("list[%d] = %q, want %q", i, c.Name, want[i])
		}
	}
}

func TestMemory_TicketLifecycle(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	c, err := m.CreateClient(ctx, NewClient{Name: "Smile Dental"})
	if err != nil {
		t.Fatal(err)
	}

	tk, err := m.CreateTicket(ctx, NewTicket{ClientID: c.ID, Title: "Printer jammed again"})
	if err != nil {
		t.Fatalf("CreateTicket: %v", err)
	}
	if tk.Status != StatusOpen {
		t.Errorf("new ticket status = %q, want open", tk.Status)
	}

	closed, err := m.SetTicketStatus(ctx, tk.ID, StatusClosed)
	if err != nil {
		t.Fatalf("SetTicketStatus: %v", err)
	}
	if closed.Status != StatusClosed {
		t.Errorf("status = %q, want closed", closed.Status)
	}
	if closed.UpdatedAt.Before(tk.UpdatedAt) {
		t.Error("UpdatedAt went backwards")
	}
}

func TestMemory_ListTicketsNewestFirst(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	c, _ := m.CreateClient(ctx, NewClient{Name: "Homelab"})
	first, _ := m.CreateTicket(ctx, NewTicket{ClientID: c.ID, Title: "first"})
	second, _ := m.CreateTicket(ctx, NewTicket{ClientID: c.ID, Title: "second"})

	list, err := m.ListTickets(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].ID != second.ID || list[1].ID != first.ID {
		t.Errorf("unexpected order: %+v", list)
	}
}

func TestMemory_Validation(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	c, _ := m.CreateClient(ctx, NewClient{Name: "Homelab"})

	tests := []struct {
		name string
		run  func() error
	}{
		{"empty client name", func() error {
			_, err := m.CreateClient(ctx, NewClient{Name: "   "})
			return err
		}},
		{"empty ticket title", func() error {
			_, err := m.CreateTicket(ctx, NewTicket{ClientID: c.ID, Title: " "})
			return err
		}},
		{"missing client id", func() error {
			_, err := m.CreateTicket(ctx, NewTicket{Title: "x"})
			return err
		}},
		{"unknown client", func() error {
			_, err := m.CreateTicket(ctx, NewTicket{ClientID: 999, Title: "x"})
			return err
		}},
		{"unknown status", func() error {
			tk, _ := m.CreateTicket(ctx, NewTicket{ClientID: c.ID, Title: "x"})
			_, err := m.SetTicketStatus(ctx, tk.ID, Status("bogus"))
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(); !errors.Is(err, ErrInvalid) {
				t.Errorf("err = %v, want ErrInvalid", err)
			}
		})
	}
}
