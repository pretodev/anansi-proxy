package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

	if currentResponse.Method != "" && currentResponse.Path != "" {
		if strings.ToUpper(r.Method) != strings.ToUpper(currentResponse.Method) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != currentResponse.Path {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		if currentResponse.RequestSchema != "" && (r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH") {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}

			if !s.validateRequestBody(body, currentResponse.RequestSchema) {
				http.Error(w, "Request body validation failed", http.StatusBadRequest)
				return
			}
		}
	}

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
