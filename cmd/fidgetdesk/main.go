// Command fidgetdesk runs the FidgetDesk helpdesk: a single executable that
// serves a web UI for tracking clients and tickets.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codejenganova/fidgetdesk/internal/store"
	"github.com/codejenganova/fidgetdesk/internal/web"
)

// version is set at build time with -ldflags "-X main.version=v1.2.3".
var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "fidgetdesk:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("fidgetdesk", flag.ContinueOnError)
	addr := fs.String("addr", "127.0.0.1:8080",
		"address to listen on; the default only accepts connections from this computer")
	showVersion := fs.Bool("version", false, "print the version and exit")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *showVersion {
		fmt.Println("fidgetdesk", version)
		return nil
	}

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	handler, err := web.New(store.NewMemory(), log, version)
	if err != nil {
		return fmt.Errorf("init web: %w", err)
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", *addr, err)
	}

	srv := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelWarn),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()
	log.Info("FidgetDesk is running", "url", "http://"+ln.Addr().String(), "version", version)

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		log.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
