package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pretodev/anansi-proxy/internal/parser"
	"github.com/pretodev/anansi-proxy/internal/state"
)

type Server struct {
	state    *state.StateManager
	endpoint parser.EndpointSchema
}

func New(sm *state.StateManager, endpoint *parser.EndpointSchema) *Server {
	return &Server{
		state:    sm,
		endpoint: *endpoint,
	}
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	index := s.state.Index()
	currentResponse := s.endpoint.Responses[index]

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

	addr := fmt.Sprintf(":%d", port)

	http.HandleFunc(s.endpoint.Route, s.handler)

	fmt.Printf("Starting server on port %d...\n", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		return fmt.Errorf("failed to start server on port %d: %w", port, err)
	}

	return nil
}

// TODO: Implement full JSON schema validation
func (s *Server) validateRequestBody(body []byte, schema string) bool {
	if strings.Contains(schema, `"$schema"`) {
		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			return false
		}
		// For now, just validate that it's valid JSON
		// TODO: Implement full JSON schema validation
		return true
	}

	// For XML schemas, just check if body is not empty
	if strings.Contains(schema, "xs:schema") || strings.Contains(schema, "xsd:schema") {
		return len(strings.TrimSpace(string(body))) > 0
	}

	// If no specific schema format detected, assume valid
	return true
}
