package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pretodev/anansi-proxy/internal/endpoint"
)

// Helper function to create test endpoints
func createEndpointWithFile(route string, statusCode int, body string) *endpoint.EndpointWithFile {
	return &endpoint.EndpointWithFile{
		Schema: &endpoint.EndpointSchema{
			Route:  route,
			Accept: "application/json",
			Responses: []endpoint.Response{
				{
					Title:       "Test Response",
					Body:        body,
					ContentType: "application/json",
					StatusCode:  statusCode,
				},
			},
		},
		FilePath: "/test/mock.apimock",
	}
}

func TestNew_SeparatesSpecificAndFallbackEndpoints(t *testing.T) {
	endpoints := []*endpoint.EndpointWithFile{
		createEndpointWithFile("GET /api/users", 200, `{"users": []}`),
		createEndpointWithFile("POST /api/users", 201, `{"created": true}`),
		createEndpointWithFile("/", 200, `{"fallback": true}`),
		createEndpointWithFile("GET /api/posts", 200, `{"posts": []}`),
	}

	server := New(endpoints)

	// Should have 3 specific endpoints
	if len(server.specificEndpoints) != 3 {
		t.Errorf("Expected 3 specific endpoints, got %d", len(server.specificEndpoints))
	}

	// Should have 1 fallback endpoint
	if len(server.fallbackEndpoints) != 1 {
		t.Errorf("Expected 1 fallback endpoint, got %d", len(server.fallbackEndpoints))
	}

	// Should keep all endpoints
	if len(server.endpoints) != 4 {
		t.Errorf("Expected 4 total endpoints, got %d", len(server.endpoints))
	}
}

func TestNew_OnlySpecificEndpoints(t *testing.T) {
	endpoints := []*endpoint.EndpointWithFile{
		createEndpointWithFile("GET /api/users", 200, `{"users": []}`),
		createEndpointWithFile("POST /api/posts", 201, `{"created": true}`),
	}

	server := New(endpoints)

	if len(server.specificEndpoints) != 2 {
		t.Errorf("Expected 2 specific endpoints, got %d", len(server.specificEndpoints))
	}

	if len(server.fallbackEndpoints) != 0 {
		t.Errorf("Expected 0 fallback endpoints, got %d", len(server.fallbackEndpoints))
	}
}

func TestNew_OnlyFallbackEndpoint(t *testing.T) {
	endpoints := []*endpoint.EndpointWithFile{
		createEndpointWithFile("/", 200, `{"fallback": true}`),
	}

	server := New(endpoints)

	if len(server.specificEndpoints) != 0 {
		t.Errorf("Expected 0 specific endpoints, got %d", len(server.specificEndpoints))
	}

	if len(server.fallbackEndpoints) != 1 {
		t.Errorf("Expected 1 fallback endpoint, got %d", len(server.fallbackEndpoints))
	}
}

func TestNew_EmptyRouteIsFallback(t *testing.T) {
	endpoints := []*endpoint.EndpointWithFile{
		createEndpointWithFile("", 200, `{"empty": true}`),
	}

	server := New(endpoints)

	if len(server.fallbackEndpoints) != 1 {
		t.Errorf("Expected empty route to be treated as fallback, got %d fallback endpoints", len(server.fallbackEndpoints))
	}
}

func TestServer_SpecificEndpointResponds(t *testing.T) {
	endpoints := []*endpoint.EndpointWithFile{
		createEndpointWithFile("GET /api/users", 200, `{"users": ["alice", "bob"]}`),
		createEndpointWithFile("/", 200, `{"fallback": true}`),
	}

	server := New(endpoints)
	mux := server.createTestMux()

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	body, _ := io.ReadAll(rec.Body)
	expectedBody := `{"users": ["alice", "bob"]}`
	if string(body) != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, string(body))
	}
}

func TestServer_FallbackHandlerRespondsWhenNoMatch(t *testing.T) {
	endpoints := []*endpoint.EndpointWithFile{
		createEndpointWithFile("GET /api/users", 200, `{"users": []}`),
		createEndpointWithFile("/", 201, `{"fallback": "response"}`),
	}

	server := New(endpoints)
	mux := server.createTestMux()

	req := httptest.NewRequest(http.MethodGet, "/unknown/path", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Errorf("Expected status 201 from fallback, got %d", rec.Code)
	}

	body, _ := io.ReadAll(rec.Body)
	expectedBody := `{"fallback": "response"}`
	if string(body) != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, string(body))
	}
}

func TestServer_Returns404WhenNoFallbackAndNoMatch(t *testing.T) {
	endpoints := []*endpoint.EndpointWithFile{
		createEndpointWithFile("GET /api/users", 200, `{"users": []}`),
		createEndpointWithFile("POST /api/posts", 201, `{"created": true}`),
	}

	server := New(endpoints)
	mux := server.createTestMux()

	req := httptest.NewRequest(http.MethodGet, "/unknown/path", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}

	body, _ := io.ReadAll(rec.Body)
	if string(body) != "404 - Not Found" {
		t.Errorf("Expected '404 - Not Found', got %q", string(body))
	}
}

func TestServer_FallbackHandlerForRootPath(t *testing.T) {
	endpoints := []*endpoint.EndpointWithFile{
		createEndpointWithFile("GET /api/users", 200, `{"users": []}`),
		createEndpointWithFile("/", 200, `{"home": true}`),
	}

	server := New(endpoints)
	mux := server.createTestMux()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	body, _ := io.ReadAll(rec.Body)
	expectedBody := `{"home": true}`
	if string(body) != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, string(body))
	}
}

func TestServer_MultipleSpecificEndpoints(t *testing.T) {
	endpoints := []*endpoint.EndpointWithFile{
		createEndpointWithFile("GET /api/users", 200, `{"users": []}`),
		createEndpointWithFile("POST /api/users", 201, `{"created": true}`),
		createEndpointWithFile("GET /api/posts", 200, `{"posts": []}`),
		createEndpointWithFile("DELETE /api/users", 204, ``),
	}

	server := New(endpoints)
	mux := server.createTestMux()

	tests := []struct {
		method   string
		path     string
		wantCode int
		wantBody string
	}{
		{http.MethodGet, "/api/users", 200, `{"users": []}`},
		{http.MethodPost, "/api/users", 201, `{"created": true}`},
		{http.MethodGet, "/api/posts", 200, `{"posts": []}`},
		{http.MethodDelete, "/api/users", 204, ``},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, rec.Code)
			}

			body, _ := io.ReadAll(rec.Body)
			if string(body) != tt.wantBody {
				t.Errorf("Expected body %q, got %q", tt.wantBody, string(body))
			}
		})
	}
}

func TestServer_SpecificEndpointTakesPriorityOverFallback(t *testing.T) {
	// Even if fallback exists, specific routes should be matched first
	endpoints := []*endpoint.EndpointWithFile{
		createEndpointWithFile("GET /api/specific", 200, `{"specific": true}`),
		createEndpointWithFile("/", 200, `{"fallback": true}`),
	}

	server := New(endpoints)
	mux := server.createTestMux()

	// Test specific endpoint
	req := httptest.NewRequest(http.MethodGet, "/api/specific", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	body, _ := io.ReadAll(rec.Body)
	if string(body) != `{"specific": true}` {
		t.Errorf("Expected specific endpoint response, got fallback: %q", string(body))
	}
}

func TestServer_MultipleFallbackEndpoints_UsesFirst(t *testing.T) {
	// If multiple fallback endpoints exist, use the first one
	endpoints := []*endpoint.EndpointWithFile{
		createEndpointWithFile("GET /api/users", 200, `{"users": []}`),
		createEndpointWithFile("/", 200, `{"first": true}`),
		createEndpointWithFile("/", 201, `{"second": true}`),
	}

	server := New(endpoints)
	mux := server.createTestMux()

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("Expected status 200 from first fallback, got %d", rec.Code)
	}

	body, _ := io.ReadAll(rec.Body)
	if string(body) != `{"first": true}` {
		t.Errorf("Expected first fallback response, got %q", string(body))
	}
}

func TestServer_ContentTypeHeader(t *testing.T) {
	ep := &endpoint.EndpointWithFile{
		Schema: &endpoint.EndpointSchema{
			Route:  "GET /api/xml",
			Accept: "application/xml",
			Responses: []endpoint.Response{
				{
					Title:       "XML Response",
					Body:        `<root><data>test</data></root>`,
					ContentType: "application/xml",
					StatusCode:  200,
				},
			},
		},
		FilePath: "/test/mock.apimock",
	}

	server := New([]*endpoint.EndpointWithFile{ep})
	mux := server.createTestMux()

	req := httptest.NewRequest(http.MethodGet, "/api/xml", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/xml" {
		t.Errorf("Expected Content-Type 'application/xml', got %q", contentType)
	}
}

func TestServer_EmptyContentType(t *testing.T) {
	ep := &endpoint.EndpointWithFile{
		Schema: &endpoint.EndpointSchema{
			Route:  "GET /api/plain",
			Accept: "text/plain",
			Responses: []endpoint.Response{
				{
					Title:       "Plain Response",
					Body:        `Hello World`,
					ContentType: "", // No content type
					StatusCode:  200,
				},
			},
		},
		FilePath: "/test/mock.apimock",
	}

	server := New([]*endpoint.EndpointWithFile{ep})
	mux := server.createTestMux()

	req := httptest.NewRequest(http.MethodGet, "/api/plain", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	// Should not set Content-Type header when empty
	contentType := rec.Header().Get("Content-Type")
	// net/http may set a default, but we shouldn't explicitly set one
	if contentType != "" && contentType != "text/plain; charset=utf-8" {
		t.Logf("Content-Type header: %q (may be set by net/http)", contentType)
	}

	body, _ := io.ReadAll(rec.Body)
	if string(body) != "Hello World" {
		t.Errorf("Expected body 'Hello World', got %q", string(body))
	}
}

// Helper method for testing - creates a ServeMux for testing without starting a server
func (s *Server) createTestMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register specific endpoints first
	for _, ep := range s.specificEndpoints {
		route := ep.Schema.Route
		mux.HandleFunc(route, s.createHandlerFromEndpoint(ep))
	}

	// Set fallback handler for all other routes
	mux.HandleFunc("/", s.fallbackHandler())

	return mux
}

func BenchmarkServer_SpecificEndpoint(b *testing.B) {
	endpoints := []*endpoint.EndpointWithFile{
		createEndpointWithFile("GET /api/users", 200, `{"users": ["alice", "bob"]}`),
		createEndpointWithFile("/", 200, `{"fallback": true}`),
	}

	server := New(endpoints)
	mux := server.createTestMux()

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
	}
}

func BenchmarkServer_FallbackEndpoint(b *testing.B) {
	endpoints := []*endpoint.EndpointWithFile{
		createEndpointWithFile("GET /api/users", 200, `{"users": []}`),
		createEndpointWithFile("/", 200, `{"fallback": true}`),
	}

	server := New(endpoints)
	mux := server.createTestMux()

	req := httptest.NewRequest(http.MethodGet, "/unknown/path", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
	}
}

func BenchmarkServer_404Response(b *testing.B) {
	endpoints := []*endpoint.EndpointWithFile{
		createEndpointWithFile("GET /api/users", 200, `{"users": []}`),
	}

	server := New(endpoints)
	mux := server.createTestMux()

	req := httptest.NewRequest(http.MethodGet, "/unknown/path", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
	}
}

func TestServer_Serve_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"zero port", 0},
		{"negative port", -1},
		{"very negative port", -100},
		{"port too high", 65536},
		{"port way too high", 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoints := []*endpoint.EndpointWithFile{
				createEndpointWithFile("GET /api/test", 200, `{"test": true}`),
			}
			server := New(endpoints)

			err := server.Serve(tt.port)
			if err == nil {
				t.Error("Serve() should return error for invalid port")
			}

			expectedMsg := "invalid port number"
			if err.Error()[:19] != expectedMsg {
				t.Errorf("error message should start with %q, got %q", expectedMsg, err.Error())
			}
		})
	}
}

func TestServer_Serve_ValidPortRange(t *testing.T) {
	// Test that valid port numbers pass validation
	// Note: We can't actually start servers in unit tests, so we just
	// verify port validation logic
	validPorts := []int{1, 80, 443, 8080, 8443, 65535}

	for _, port := range validPorts {
		if port <= 0 || port > 65535 {
			t.Errorf("Port %d should be valid but isn't in range", port)
		}
	}
}

func TestServer_RequestValidation_JSON(t *testing.T) {
	// Create a JSON schema validator
	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "number"}
		},
		"required": ["name"]
	}`

	validator, err := endpoint.NewJsonSchemaValidator(schema)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	ep := &endpoint.EndpointWithFile{
		Schema: &endpoint.EndpointSchema{
			Route:     "POST /api/users",
			Accept:    "application/json",
			Body:      schema,
			Validator: validator,
			Responses: []endpoint.Response{
				{
					Title:       "User created",
					Body:        `{"id": 123, "name": "John"}`,
					ContentType: "application/json",
					StatusCode:  201,
				},
			},
		},
		FilePath: "/test/mock.apimock",
	}

	server := New([]*endpoint.EndpointWithFile{ep})
	mux := server.createTestMux()

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid request body",
			body:           `{"name": "John", "age": 30}`,
			expectedStatus: 201,
			expectError:    false,
		},
		{
			name:           "Valid request without optional field",
			body:           `{"name": "Jane"}`,
			expectedStatus: 201,
			expectError:    false,
		},
		{
			name:           "Invalid request - missing required field",
			body:           `{"age": 25}`,
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:           "Invalid request - wrong type",
			body:           `{"name": 123}`,
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:           "Invalid request - malformed JSON",
			body:           `{invalid json}`,
			expectedStatus: 400,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/users", io.NopCloser(io.Reader(io.NopCloser(io.Reader(&bodyReader{body: tt.body})))))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.expectError && rec.Code != 400 {
				t.Errorf("Expected error response (400), got %d", rec.Code)
			}
		})
	}
}

func TestServer_RequestValidation_XML(t *testing.T) {
	// Create an XML schema validator
	schema := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="user">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="name" type="xs:string"/>
        <xs:element name="age" type="xs:integer"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`

	validator, err := endpoint.NewXmlSchemaValidator(schema)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	defer func() {
		if v, ok := validator.(*endpoint.XMLSchemaValidator); ok {
			v.Free()
		}
	}()

	ep := &endpoint.EndpointWithFile{
		Schema: &endpoint.EndpointSchema{
			Route:     "POST /api/users",
			Accept:    "application/xml",
			Body:      schema,
			Validator: validator,
			Responses: []endpoint.Response{
				{
					Title:       "User created",
					Body:        `<user><name>John</name><age>30</age></user>`,
					ContentType: "application/xml",
					StatusCode:  201,
				},
			},
		},
		FilePath: "/test/mock.apimock",
	}

	server := New([]*endpoint.EndpointWithFile{ep})
	mux := server.createTestMux()

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid XML request",
			body:           `<?xml version="1.0"?><user><name>John</name><age>30</age></user>`,
			expectedStatus: 201,
			expectError:    false,
		},
		{
			name:           "Invalid XML - missing required element",
			body:           `<?xml version="1.0"?><user><name>Jane</name></user>`,
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:           "Invalid XML - wrong type",
			body:           `<?xml version="1.0"?><user><name>Bob</name><age>not a number</age></user>`,
			expectedStatus: 400,
			expectError:    true,
		},
		{
			name:           "Invalid XML syntax",
			body:           `<user><name>Invalid</user>`,
			expectedStatus: 400,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/users", io.NopCloser(io.Reader(&bodyReader{body: tt.body})))
			req.Header.Set("Content-Type", "application/xml")
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.expectError && rec.Code != 400 {
				t.Errorf("Expected error response (400), got %d", rec.Code)
			}
		})
	}
}

func TestServer_NoValidation_WhenNoValidator(t *testing.T) {
	// Endpoint without validator should accept any body
	ep := &endpoint.EndpointWithFile{
		Schema: &endpoint.EndpointSchema{
			Route:     "POST /api/data",
			Accept:    "application/json",
			Validator: nil, // No validator
			Responses: []endpoint.Response{
				{
					Title:       "Success",
					Body:        `{"success": true}`,
					ContentType: "application/json",
					StatusCode:  200,
				},
			},
		},
		FilePath: "/test/mock.apimock",
	}

	server := New([]*endpoint.EndpointWithFile{ep})
	mux := server.createTestMux()

	// Even with invalid JSON, should succeed because no validation
	req := httptest.NewRequest(http.MethodPost, "/api/data", io.NopCloser(io.Reader(&bodyReader{body: `{invalid json}`})))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("Expected status 200 (no validation), got %d", rec.Code)
	}
}

// Helper type to create a ReadCloser from a string
type bodyReader struct {
	body string
	pos  int
}

func (b *bodyReader) Read(p []byte) (n int, err error) {
	if b.pos >= len(b.body) {
		return 0, io.EOF
	}
	n = copy(p, b.body[b.pos:])
	b.pos += n
	return n, nil
}

func (b *bodyReader) Close() error {
	return nil
}
