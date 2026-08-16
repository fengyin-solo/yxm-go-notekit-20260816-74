package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/notekit/internal/config"
	"github.com/example/notekit/internal/logger"
	"github.com/example/notekit/internal/middleware"
	"github.com/example/notekit/internal/service"
	"github.com/example/notekit/internal/store"
)

// Server is the top-level HTTP server that wires all layers together.
type Server struct {
	httpServer *http.Server
	nbStore    store.NotebookStore
	noteStore  store.NoteStore
}

// NewServer creates a new Server with all routes and middleware wired up.
func NewServer(cfg config.Config) (*Server, error) {
	log := logger.Default()

	nbStore, err := store.NewMemoryNotebookStore(cfg.DataFile+"notebooks.json", log, cfg.SaveInterval)
	if err != nil {
		return nil, fmt.Errorf("notebook store: %w", err)
	}
	noteStore, err := store.NewMemoryNoteStore(cfg.DataFile+"notes.json", log, cfg.SaveInterval)
	if err != nil {
		return nil, fmt.Errorf("note store: %w", err)
	}

	nbSvc := service.NewNotebookService(nbStore, noteStore)
	noteSvc := service.NewNoteService(noteStore, cfg)

	notebookHandler := NewNotebookHandler(nbSvc, cfg.MaxBody)
	noteHandler := NewNoteHandler(noteSvc, cfg.MaxBody)

	router := NewRouter(notebookHandler, noteHandler)

	var handler http.Handler = router.Handler()
	handler = middleware.RequestID(handler)
	handler = middleware.Logging(handler)
	handler = middleware.Recovery(handler)
	handler = middleware.RequireAuth(cfg.AuthToken, handler)
	handler = middleware.LimitFromConfig(cfg.RateLimit)(handler)
	handler = middleware.Timeout(30 * time.Second)(handler)

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.Addr,
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		nbStore:   nbStore,
		noteStore: noteStore,
	}, nil
}

// Start starts the HTTP server and blocks until the context is cancelled.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-ctx.Done():
		return s.Shutdown(context.Background())
	}
}

// Shutdown gracefully shuts down the server, flushing stores and closing connections.
func (s *Server) Shutdown(ctx context.Context) error {
	_ = s.noteStore.Close()
	_ = s.nbStore.Close()
	return s.httpServer.Shutdown(ctx)
}

// ListenAndRun creates a server, listens for connections, and handles graceful shutdown.
func ListenAndRun(cfg config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	srv, err := NewServer(cfg)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	go func() {
		<-sigCh
		cancel()
	}()

	return srv.Start(ctx)
}
