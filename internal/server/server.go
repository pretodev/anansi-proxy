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
	currentResponse := s.res[s.state.Index()]

	if currentResponse.ContentType != "" {
		w.Header().Set("Content-Type", currentResponse.ContentType)
	}

	w.WriteHeader(currentResponse.StatusCode)
	fmt.Fprint(w, currentResponse.Body)
}

func (s *Server) Serve(port int) error {
	addr := fmt.Sprintf(":%d", port)

	http.HandleFunc("/", s.handler)
	return http.ListenAndServe(addr, nil)
}
