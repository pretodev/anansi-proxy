package server

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pretodev/anansi-proxy/internal/endpoint"
	"github.com/pretodev/anansi-proxy/internal/state"
	"github.com/pretodev/anansi-proxy/pkg/apimock"
)

type Server struct {
	endpoints         []*endpoint.EndpointWithFile
	specificEndpoints []*endpoint.EndpointWithFile // endpoints with specific routes (not "/")
	fallbackEndpoints []*endpoint.EndpointWithFile // endpoints with "/" route
	state             *state.StateManager
	astCache          map[string]*apimock.APIMockFile // Cache parsed AST by file path
}

func New(endpoints []*endpoint.EndpointWithFile) *Server {
	s := &Server{
		endpoints:         endpoints,
		specificEndpoints: make([]*endpoint.EndpointWithFile, 0),
		fallbackEndpoints: make([]*endpoint.EndpointWithFile, 0),
		state:             state.New(0), // max not used for call tracking
		astCache:          make(map[string]*apimock.APIMockFile),
	}

	// Separate specific routes from fallback routes
	for _, ep := range endpoints {
		if ep.Schema.Route == "/" || ep.Schema.Route == "" {
			s.fallbackEndpoints = append(s.fallbackEndpoints, ep)
		} else {
			s.specificEndpoints = append(s.specificEndpoints, ep)
		}

		// Pre-load AST for each endpoint
		if ast, err := parseAPIMockFile(ep.FilePath); err == nil {
			s.astCache[ep.FilePath] = ast
		}
	}

	return s
}

// buildRequestContext creates a RequestContext from an HTTP request
func buildRequestContext(r *http.Request, bodyBytes []byte) *apimock.RequestContext {
	// Parse query parameters
	query := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			query[key] = values[0] // Take first value
		}
	}

	// Parse headers
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[strings.ToLower(key)] = values[0] // Normalize to lowercase
		}
	}

	// TODO: Extract path parameters from route pattern matching
	// For now, just use the full path

	return &apimock.RequestContext{
		Method:  r.Method,
		Path:    r.URL.Path,
		Query:   query,
		Headers: headers,
		Body:    string(bodyBytes),
	}
}

// selectResponseWithConditions evaluates conditions and selects the appropriate response
func (s *Server) selectResponseWithConditions(responses []endpoint.Response, reqCtx *apimock.RequestContext, callCount int, ast *apimock.APIMockFile) endpoint.Response {
	if ast == nil || len(ast.Responses) == 0 {
		// Fallback: return first response if no AST available
		if len(responses) > 0 {
			return responses[0]
		}
		return endpoint.EmptyResponse()
	}

	// Create execution context
	ctx := apimock.NewExecutionContext(callCount, reqCtx)

	// Match responses by evaluating conditions
	for i, astResp := range ast.Responses {
		if i >= len(responses) {
			break
		}

		// If no conditions, this response always matches
		if len(astResp.Conditions) == 0 {
			resp := responses[i]
			resp.Body = interpolateVariables(resp.Body, ctx)
			return resp
		}

		// Create evaluator for this response
		evaluator := apimock.NewEvaluator(ctx)

		// Evaluate all conditions
		conditionsMet, err := evaluator.EvaluateConditions(astResp.Conditions)
		if err != nil {
			// If condition evaluation fails, skip this response
			continue
		}

		if conditionsMet {
			resp := responses[i]
			resp.Body = interpolateVariables(resp.Body, ctx)
			return resp
		}
	}

	// No response matched conditions, return first response as fallback
	if len(responses) > 0 {
		resp := responses[0]
		resp.Body = interpolateVariables(resp.Body, ctx)
		return resp
	}
	return endpoint.EmptyResponse()
}

// interpolateVariables replaces {{variable}} placeholders in the response body
func interpolateVariables(body string, ctx *apimock.ExecutionContext) string {
	result := body

	// Replace {{call_count}}
	result = strings.ReplaceAll(result, "{{call_count}}", fmt.Sprintf("%d", ctx.CallCount))

	// Replace other variables from context
	for key, value := range ctx.Variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result
}

// parseAPIMockFile loads and parses an APIMock file
func parseAPIMockFile(filePath string) (*apimock.APIMockFile, error) {
	parser, err := apimock.NewParser(filePath)
	if err != nil {
		return nil, err
	}

	ast, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	return ast, nil
}

func (s *Server) createHandlerFromEndpoint(ep *endpoint.EndpointWithFile) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read request body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			bodyBytes = []byte{}
		}
		defer r.Body.Close()

		// Increment call count for this endpoint
		callCount := s.state.IncrementCallCount(ep.Schema.Route)

		// Build request context
		reqCtx := buildRequestContext(r, bodyBytes)

		// Get AST from cache
		ast := s.astCache[ep.FilePath]

		// Validate request body if validator is present
		if ep.Schema.Validator != nil {
			if err := ep.Schema.Validator.Validate(string(bodyBytes)); err != nil {
				badResp, hasBadResp := ep.Schema.GetResponseByStatusCode(http.StatusBadRequest)
				if hasBadResp {
					if badResp.ContentType != "" {
						w.Header().Set("Content-Type", badResp.ContentType)
					}
					w.WriteHeader(badResp.StatusCode)
					fmt.Fprint(w, badResp.Body)
					return
				}
				http.Error(w, fmt.Sprintf("Request validation failed: %v", err), http.StatusBadRequest)
				return
			}
		}

		// Get all responses for the endpoint
		responses := ep.Schema.SliceResponses()

		// Select response based on conditions
		resp := s.selectResponseWithConditions(responses, reqCtx, callCount, ast)

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

			// Read request body
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				bodyBytes = []byte{}
			}
			defer r.Body.Close()

			// Increment call count for fallback endpoint
			callCount := s.state.IncrementCallCount(ep.Schema.Route)

			// Build request context
			reqCtx := buildRequestContext(r, bodyBytes)

			// Get AST from cache
			ast := s.astCache[ep.FilePath]

			// Get all responses for the endpoint
			responses := ep.Schema.SliceResponses()

			// Select response based on conditions
			resp := s.selectResponseWithConditions(responses, reqCtx, callCount, ast)

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
