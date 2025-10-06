package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pretodev/anansi-proxy/internal/endpoint"
)

type Server struct {
	endpoints         []*endpoint.EndpointWithFile
	specificEndpoints []*endpoint.EndpointWithFile // endpoints with specific routes (not "/")
	fallbackEndpoints []*endpoint.EndpointWithFile // endpoints with "/" route
}

func New(endpoints []*endpoint.EndpointWithFile) *Server {
	s := &Server{
		endpoints:         endpoints,
		specificEndpoints: make([]*endpoint.EndpointWithFile, 0),
		fallbackEndpoints: make([]*endpoint.EndpointWithFile, 0),
	}

	// Separate specific routes from fallback routes
	for _, ep := range endpoints {
		if ep.Schema.Route == "/" || ep.Schema.Route == "" {
			s.fallbackEndpoints = append(s.fallbackEndpoints, ep)
		} else {
			s.specificEndpoints = append(s.specificEndpoints, ep)
		}
	}

	return s
}

func (s *Server) createHandlerFromEndpoint(ep *endpoint.EndpointWithFile) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the first available response (prioritize 200 OK if available, otherwise use first in map)
		resp := endpoint.EmptyResponse()
		if okResp, ok := ep.Schema.GetResponseByStatusCode(http.StatusOK); ok {
			resp = okResp
		} else {
			// If no 200 OK response, use the first available response
			responses := ep.Schema.SliceResponses()
			if len(responses) > 0 {
				resp = responses[0]
			}
		}

		if ep.Schema.Validator != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			defer r.Body.Close()

			if err != nil {
				badResp, hasBadResp := ep.Schema.GetResponseByStatusCode(http.StatusBadRequest)
				if hasBadResp {
					resp = badResp
				} else {
					http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
					return
				}
			} else if err := ep.Schema.Validator.Validate(string(bodyBytes)); err != nil {
				badResp, hasBadResp := ep.Schema.GetResponseByStatusCode(http.StatusBadRequest)
				if hasBadResp {
					resp = badResp
				} else {
					http.Error(w, fmt.Sprintf("Request validation failed: %v", err), http.StatusBadRequest)
					return
				}
			}
		}

		if resp.ContentType != "" {
			w.Header().Set("Content-Type", resp.ContentType)
		}

		w.WriteHeader(resp.StatusCode)
		fmt.Fprint(w, resp.Body)
	}
}

func (s *Server) fallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(s.fallbackEndpoints) > 0 {
			ep := s.fallbackEndpoints[0]

			// Get the first available response (prioritize 200 OK if available, otherwise use first in map)
			resp := endpoint.EmptyResponse()
			if okResp, ok := ep.Schema.GetResponseByStatusCode(http.StatusOK); ok {
				resp = okResp
			} else {
				// If no 200 OK response, use the first available response
				responses := ep.Schema.SliceResponses()
				if len(responses) > 0 {
					resp = responses[0]
				}
			}

			if resp.ContentType != "" {
				w.Header().Set("Content-Type", resp.ContentType)
			}

			w.WriteHeader(resp.StatusCode)
			fmt.Fprint(w, resp.Body)
			return
		}

		// No fallback endpoint, return 404
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "404 - Not Found")
	}
}

func (s *Server) Serve(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port number %d: must be between 1 and 65535", port)
	}
	addr := fmt.Sprintf(":%d", port)

	mux := http.NewServeMux()

	for _, ep := range s.specificEndpoints {
		route := ep.Schema.Route
		mux.HandleFunc(route, s.createHandlerFromEndpoint(ep))
	}

	mux.HandleFunc("/", s.fallbackHandler())

	fmt.Printf("\nStarting server on port %d...\n", port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		return fmt.Errorf("failed to start server on port %d: %w", port, err)
	}

	return nil
}
