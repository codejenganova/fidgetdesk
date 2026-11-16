package web

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/codejenganova/fidgetdesk/internal/store"
)

func newTestServer(t *testing.T) (*Server, *store.Memory) {
	t.Helper()
	mem := store.NewMemory()
	srv, err := New(mem, slog.New(slog.NewTextHandler(io.Discard, nil)), "test")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return srv, mem
}

func postForm(srv http.Handler, path string, form url.Values, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestHealthz(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "ok" {
		t.Fatalf("got %d %q", rec.Code, rec.Body.String())
	}
}

func TestIndex_RendersAndSetsSecurityHeaders(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "FidgetDesk") {
		t.Error("page missing title")
	}
	for _, h := range []string{"Content-Security-Policy", "X-Content-Type-Options", "X-Frame-Options"} {
		if rec.Header().Get(h) == "" {
			t.Errorf("missing header %s", h)
		}
	}
}

func TestCreateClientAndTicket(t *testing.T) {
	srv, mem := newTestServer(t)

	rec := postForm(srv, "/clients", url.Values{"name": {"Smile Dental"}}, nil)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("create client status = %d", rec.Code)
	}
	clients, _ := mem.ListClients(context.Background())
	if len(clients) != 1 {
		t.Fatalf("clients = %d, want 1", len(clients))
	}

	rec = postForm(srv, "/tickets", url.Values{
		"client_id": {"1"},
		"title":     {"<script>alert(1)</script>"},
	}, nil)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("create ticket status = %d", rec.Code)
	}

	// The title must be HTML-escaped on render.
	page := httptest.NewRecorder()
	srv.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/", nil))
	body := page.Body.String()
	if strings.Contains(body, "<script>alert(1)</script>") {
		t.Error("ticket title was not escaped")
	}
	if !strings.Contains(body, "Smile Dental") {
		t.Error("client name missing from ticket list")
	}
}

func TestSetTicketStatus(t *testing.T) {
	srv, mem := newTestServer(t)
	ctx := context.Background()
	c, _ := mem.CreateClient(ctx, store.NewClient{Name: "Homelab"})
	tk, _ := mem.CreateTicket(ctx, store.NewTicket{ClientID: c.ID, Title: "Replace UPS battery"})

	rec := postForm(srv, fmt.Sprintf("/tickets/%d/status", tk.ID), url.Values{"status": {"closed"}}, nil)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d", rec.Code)
	}
	got, _ := mem.GetTicket(ctx, tk.ID)
	if got.Status != store.StatusClosed {
		t.Errorf("ticket status = %q, want closed", got.Status)
	}

	if rec := postForm(srv, "/tickets/999/status", url.Values{"status": {"closed"}}, nil); rec.Code != http.StatusNotFound {
		t.Errorf("missing ticket: status = %d, want 404", rec.Code)
	}
	if rec := postForm(srv, "/tickets/abc/status", url.Values{"status": {"closed"}}, nil); rec.Code != http.StatusNotFound {
		t.Errorf("bad id: status = %d, want 404", rec.Code)
	}
}

func TestValidationErrorRendersMessage(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := postForm(srv, "/clients", url.Values{"name": {"  "}}, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "client name is required") {
		t.Errorf("error message not shown: %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), store.ErrInvalid.Error()) {
		t.Error("internal sentinel text leaked to the page")
	}
}

func TestCrossOriginPostsAreRejected(t *testing.T) {
	srv, mem := newTestServer(t)
	tests := []struct {
		name    string
		headers map[string]string
		want    int
	}{
		{"cross-site fetch metadata", map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusForbidden},
		{"same-origin fetch metadata", map[string]string{"Sec-Fetch-Site": "same-origin"}, http.StatusSeeOther},
		{"foreign origin header", map[string]string{"Origin": "https://evil.example"}, http.StatusForbidden},
		{"matching origin header", map[string]string{"Origin": "http://example.com"}, http.StatusSeeOther},
		{"no browser headers", nil, http.StatusSeeOther},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := postForm(srv, "/clients", url.Values{"name": {tt.name}}, tt.headers)
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
	clients, _ := mem.ListClients(context.Background())
	if len(clients) != 3 {
		t.Errorf("clients created = %d, want 3 (only allowed requests)", len(clients))
	}
}

func TestStaticAssetsServed(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/static/style.css", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}
