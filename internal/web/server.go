// Package web serves FidgetDesk's browser UI. Templates and static assets
// are embedded into the binary so the app ships as a single file.
package web

import (
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/codejenganova/fidgetdesk/internal/store"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

// Server wires HTTP routes to a store.Store.
type Server struct {
	store   store.Store
	log     *slog.Logger
	tmpl    *template.Template
	version string
	mux     *http.ServeMux
}

// New builds a Server. It fails only if the embedded templates are broken,
// which tests catch long before a release.
func New(s store.Store, log *slog.Logger, version string) (*Server, error) {
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"statuses": func() []store.Status {
			return []store.Status{store.StatusOpen, store.StatusWaiting, store.StatusClosed}
		},
	}).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		return nil, err
	}

	srv := &Server{store: s, log: log, tmpl: tmpl, version: version, mux: http.NewServeMux()}
	srv.mux.HandleFunc("GET /healthz", srv.healthz)
	srv.mux.HandleFunc("GET /{$}", srv.index)
	srv.mux.HandleFunc("POST /clients", srv.createClient)
	srv.mux.HandleFunc("POST /tickets", srv.createTicket)
	srv.mux.HandleFunc("POST /tickets/{id}/status", srv.setTicketStatus)
	srv.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static)))
	return srv, nil
}

// ServeHTTP implements http.Handler with security headers and a
// cross-origin check applied to every request.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Referrer-Policy", "same-origin")
	h.Set("Content-Security-Policy", "default-src 'self'; form-action 'self'; frame-ancestors 'none'")

	if !sameOrigin(r) {
		http.Error(w, "cross-origin request rejected", http.StatusForbidden)
		return
	}
	s.mux.ServeHTTP(w, r)
}

// sameOrigin blocks cross-site form posts. Without it, any web page open in
// the user's browser could submit forms to a FidgetDesk running on
// localhost (a CSRF attack). Safe methods are always allowed.
func sameOrigin(r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	// Modern browsers send Sec-Fetch-Site on every request.
	if site := r.Header.Get("Sec-Fetch-Site"); site != "" {
		return site == "same-origin" || site == "none"
	}
	// Fall back to comparing Origin with Host for older browsers.
	if origin := r.Header.Get("Origin"); origin != "" {
		u, err := url.Parse(origin)
		return err == nil && u.Host == r.Host
	}
	// No browser headers at all: a non-browser client such as curl.
	return true
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}

type ticketRow struct {
	store.Ticket
	ClientName string
}

type indexData struct {
	Version string
	Clients []store.Client
	Tickets []ticketRow
	Error   string
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	s.renderIndex(w, r, http.StatusOK, "")
}

func (s *Server) renderIndex(w http.ResponseWriter, r *http.Request, code int, errMsg string) {
	ctx := r.Context()
	clients, err := s.store.ListClients(ctx)
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	tickets, err := s.store.ListTickets(ctx)
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	names := make(map[int64]string, len(clients))
	for _, c := range clients {
		names[c.ID] = c.Name
	}
	rows := make([]ticketRow, len(tickets))
	for i, t := range tickets {
		rows[i] = ticketRow{Ticket: t, ClientName: names[t.ClientID]}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	data := indexData{Version: s.version, Clients: clients, Tickets: rows, Error: errMsg}
	if err := s.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		s.log.Error("render index", "err", err)
	}
}

func (s *Server) createClient(w http.ResponseWriter, r *http.Request) {
	in := store.NewClient{Name: r.FormValue("name"), Notes: r.FormValue("notes")}
	if _, err := s.store.CreateClient(r.Context(), in); err != nil {
		s.handleWriteErr(w, r, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) createTicket(w http.ResponseWriter, r *http.Request) {
	clientID, _ := strconv.ParseInt(r.FormValue("client_id"), 10, 64)
	in := store.NewTicket{ClientID: clientID, Title: r.FormValue("title"), Body: r.FormValue("body")}
	if _, err := s.store.CreateTicket(r.Context(), in); err != nil {
		s.handleWriteErr(w, r, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) setTicketStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if _, err := s.store.SetTicketStatus(r.Context(), id, store.Status(r.FormValue("status"))); err != nil {
		s.handleWriteErr(w, r, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleWriteErr maps store errors to HTTP responses. Validation problems
// re-render the page with a message; anything unexpected is a 500.
func (s *Server) handleWriteErr(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, store.ErrInvalid):
		s.renderIndex(w, r, http.StatusBadRequest, userMessage(err))
	case errors.Is(err, store.ErrNotFound):
		http.NotFound(w, r)
	default:
		s.serverError(w, r, err)
	}
}

// userMessage strips the ErrInvalid sentinel so users see only the detail.
func userMessage(err error) string {
	var joined interface{ Unwrap() []error }
	if errors.As(err, &joined) {
		for _, e := range joined.Unwrap() {
			if !errors.Is(e, store.ErrInvalid) {
				return e.Error()
			}
		}
	}
	return err.Error()
}

func (s *Server) serverError(w http.ResponseWriter, r *http.Request, err error) {
	s.log.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
