package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rhuss/readwise-mcp-server/internal/types"
)

// Server wraps the MCP server and HTTP infrastructure.
type Server struct {
	MCPServer *mcp.Server
	Config    types.Config
	Logger    *slog.Logger
	handler   *mcp.StreamableHTTPHandler
	mux       *http.ServeMux
}

// New creates a new Server with the given configuration.
func New(cfg types.Config, logger *slog.Logger) *Server {
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

	s.handler = mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return mcpServer
		},
		&mcp.StreamableHTTPOptions{
			Logger: logger,
		},
	)

	s.mux = http.NewServeMux()
	s.mux.Handle("/mcp", s.handler)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/ready", s.handleReady)

	return s
}

// ListenAndServe starts the HTTP server on the configured port.
func (s *Server) ListenAndServe() error {
	addr := s.Config.Addr()
	s.Logger.Info("starting server", "addr", addr, "profiles", s.Config.Profiles)
	return http.ListenAndServe(addr, s.mux)
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
