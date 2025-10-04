package server

import (
	"fmt"
	"net/http"

	"github.com/pretodev/anansi-proxy/internal/endpoint"
	"github.com/pretodev/anansi-proxy/internal/state"
)

type InteractiveServer struct {
	state    *state.StateManager
	endpoint *endpoint.EndpointSchema
}

func NewInteractive(sm *state.StateManager, endpoint *endpoint.EndpointSchema) *InteractiveServer {
	return &InteractiveServer{
		state:    sm,
		endpoint: endpoint,
	}
}

func (s *InteractiveServer) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseIndex := s.state.Index()
		currentResponse := s.endpoint.Responses[responseIndex]

		if currentResponse.ContentType != "" {
			w.Header().Set("Content-Type", currentResponse.ContentType)
		}

		w.WriteHeader(currentResponse.StatusCode)
		fmt.Fprint(w, currentResponse.Body)
	}
}

func (s *InteractiveServer) Serve(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port number %d: must be between 1 and 65535", port)
	}

	addr := fmt.Sprintf(":%d", port)

	http.HandleFunc(s.endpoint.Route, s.handler())

	fmt.Printf("\nStarting server on port %d...\n", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		return fmt.Errorf("failed to start server on port %d: %w", port, err)
	}

	return nil
}
