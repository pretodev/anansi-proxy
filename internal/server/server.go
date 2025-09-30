package server

import (
	"fmt"
	"net/http"

	"github.com/pretodev/anansi-proxy/internal/parser"
	"github.com/pretodev/anansi-proxy/internal/state"
)

type Server struct {
	state *state.StateManager
	res   []parser.Response
}

func New(sm *state.StateManager, res []parser.Response) *Server {
	return &Server{
		state: sm,
		res:   res,
	}
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	index := s.state.Index()

	if index < 0 || index >= len(s.res) {
		http.Error(w, "Internal server error: invalid response index", http.StatusInternalServerError)
		return
	}

	currentResponse := s.res[index]

	if currentResponse.ContentType != "" {
		w.Header().Set("Content-Type", currentResponse.ContentType)
	}

	w.WriteHeader(currentResponse.StatusCode)
	fmt.Fprint(w, currentResponse.Body)
}

func (s *Server) Serve(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port number %d: must be between 1 and 65535", port)
	}

	if len(s.res) == 0 {
		return fmt.Errorf("no responses available to serve")
	}

	addr := fmt.Sprintf(":%d", port)

	http.HandleFunc("/", s.handler)

	fmt.Printf("Starting server on port %d...\n", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		return fmt.Errorf("failed to start server on port %d: %w", port, err)
	}

	return nil
}
