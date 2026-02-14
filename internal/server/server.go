package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rhuss/readwise-mcp-server/internal/api"
	"github.com/rhuss/readwise-mcp-server/internal/cache"
	"github.com/rhuss/readwise-mcp-server/internal/tools"
	"github.com/rhuss/readwise-mcp-server/internal/types"
)

// Server wraps the MCP server and HTTP infrastructure.
type Server struct {
	MCPServer  *mcp.Server
	Config     types.Config
	Logger     *slog.Logger
	handler    *mcp.StreamableHTTPHandler
	mux        *http.ServeMux
	healthMux  *http.ServeMux
	httpServer *http.Server
	tlsServer  *http.Server
	httpLn     net.Listener
	tlsLn      net.Listener
}

// New creates a new Server with the given configuration.
// Returns an error if profile resolution fails.
func New(cfg types.Config, logger *slog.Logger) (*Server, error) {
	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "readwise-mcp-server",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{
			Instructions: "Readwise MCP Server provides access to Readwise and Reader APIs. " +
				"Pass your Readwise API key via the Authorization header (Token <key>).",
			Logger: logger,
		},
	)

	s := &Server{
		MCPServer: mcpServer,
		Config:    cfg,
		Logger:    logger,
	}

	// Register tools based on active profiles
	apiClient := api.NewClient()
	cm := cache.NewManager(cfg.CacheMaxSizeMB, cfg.CacheTTLSeconds, cfg.CacheEnabled)
	if err := tools.RegisterAllTools(mcpServer, apiClient, cm, cfg.Profiles); err != nil {
		return nil, fmt.Errorf("failed to resolve profiles: %w", err)
	}

	s.handler = mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return mcpServer
		},
		&mcp.StreamableHTTPOptions{
			Logger: logger,
		},
	)

	// Full mux serves all endpoints
	s.mux = http.NewServeMux()
	s.mux.Handle("/mcp", s.handler)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/ready", s.handleReady)

	// Health-only mux for HTTP listener in TLS mode
	s.healthMux = http.NewServeMux()
	s.healthMux.HandleFunc("/health", s.handleHealth)
	s.healthMux.HandleFunc("/ready", s.handleReady)

	return s, nil
}

// ListenAndServe starts the server. When TLS is configured, it starts two
// listeners: HTTPS for MCP traffic and HTTP for health/readiness probes.
// When TLS is not configured, it starts a single HTTP listener for all endpoints.
// The context controls graceful shutdown.
func (s *Server) ListenAndServe(ctx context.Context) error {
	if s.Config.TLSEnabled() {
		return s.listenDual(ctx)
	}
	return s.listenHTTP(ctx)
}

// listenHTTP starts a single HTTP listener serving all endpoints (non-TLS mode).
func (s *Server) listenHTTP(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.Config.Addr())
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.Config.Addr(), err)
	}
	s.httpLn = ln

	s.httpServer = &http.Server{Handler: s.mux}

	s.Logger.Info("starting server", "addr", ln.Addr().String(), "profiles", s.Config.Profiles)

	return s.serveWithShutdown(ctx, s.httpServer, ln)
}

// listenDual starts two listeners: HTTPS for MCP, HTTP for probes.
func (s *Server) listenDual(ctx context.Context) error {
	// Check certificate expiry
	s.checkCertExpiry()

	// Load TLS certificate
	cert, err := tls.LoadX509KeyPair(s.Config.TLSCertFile, s.Config.TLSKeyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Start HTTP listener for health probes
	httpLn, err := net.Listen("tcp", s.Config.Addr())
	if err != nil {
		return fmt.Errorf("failed to listen on HTTP %s: %w", s.Config.Addr(), err)
	}
	s.httpLn = httpLn

	// Start TLS listener for MCP
	tlsLn, err := tls.Listen("tcp", s.Config.TLSAddr(), tlsConfig)
	if err != nil {
		httpLn.Close()
		return fmt.Errorf("failed to listen on HTTPS %s: %w", s.Config.TLSAddr(), err)
	}
	s.tlsLn = tlsLn

	s.httpServer = &http.Server{Handler: s.healthMux}
	s.tlsServer = &http.Server{Handler: s.mux}

	s.Logger.Info("starting server with TLS",
		"http_addr", httpLn.Addr().String(),
		"https_addr", tlsLn.Addr().String(),
		"profiles", s.Config.Profiles,
	)

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := s.httpServer.Serve(httpLn); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := s.tlsServer.Serve(tlsLn); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("HTTPS server error: %w", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	s.Logger.Info("shutting down servers")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.httpServer.Shutdown(shutdownCtx)
	s.tlsServer.Shutdown(shutdownCtx)

	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

// serveWithShutdown serves on the listener and handles graceful shutdown via context.
func (s *Server) serveWithShutdown(ctx context.Context, srv *http.Server, ln net.Listener) error {
	errCh := make(chan error, 1)
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	<-ctx.Done()
	s.Logger.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	srv.Shutdown(shutdownCtx)

	if err, ok := <-errCh; ok {
		return err
	}
	return nil
}

// checkCertExpiry logs a warning if the TLS certificate expires within 30 days.
func (s *Server) checkCertExpiry() {
	cert, err := tls.LoadX509KeyPair(s.Config.TLSCertFile, s.Config.TLSKeyFile)
	if err != nil {
		return
	}

	if len(cert.Certificate) == 0 {
		return
	}

	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return
	}

	daysUntilExpiry := time.Until(leaf.NotAfter).Hours() / 24
	if daysUntilExpiry < 30 {
		s.Logger.Warn("TLS certificate expires soon",
			"expires_in_days", int(daysUntilExpiry),
			"not_after", leaf.NotAfter.Format(time.RFC3339),
		)
	}
}

// httpPort returns the actual HTTP listener port (useful when port 0 is used in tests).
func (s *Server) httpPort() int {
	if s.httpLn != nil {
		return s.httpLn.Addr().(*net.TCPAddr).Port
	}
	return s.Config.Port
}

// tlsPort returns the actual TLS listener port (useful when port 0 is used in tests).
func (s *Server) tlsPort() int {
	if s.tlsLn != nil {
		return s.tlsLn.Addr().(*net.TCPAddr).Port
	}
	return s.Config.TLSPort
}

// Handler returns the HTTP handler (for testing).
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok"}`)
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ready"}`)
}
