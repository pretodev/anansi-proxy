package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pretodev/anansi-proxy/internal/endpoint"
	"github.com/pretodev/anansi-proxy/pkg/apimock"
)

func TestServer_ConditionEvaluation_CallCount(t *testing.T) {
	// Create a simple .apimock file with call_count conditions
	ast := &apimock.APIMockFile{
		Request: &apimock.RequestSection{
			Method: "GET",
			Path:   "/api/test-callcount",
		},
		Responses: []apimock.ResponseSection{
			{
				StatusCode:  429,
				Description: "Too Many Requests",
				Properties: map[string]string{
					"ContentType": "application/json",
				},
				Conditions: []apimock.ConditionLine{
					{
						Expression: apimock.BinaryExpression{
							Left:     apimock.VariableReference{Name: "call_count"},
							Operator: ">",
							Right:    apimock.NumberValue{Value: 3},
						},
						IsOrCondition: false,
					},
				},
				Body: `{"error": "Too Many Requests", "call_count": {{call_count}}}`,
			},
			{
				StatusCode:  200,
				Description: "OK",
				Properties: map[string]string{
					"ContentType": "application/json",
				},
				Conditions: []apimock.ConditionLine{},
				Body:       `{"message": "Success", "call_count": {{call_count}}}`,
			},
		},
	}

	// Convert AST to EndpointSchema
	schema, err := endpoint.FromAPIMockFile(ast)
	if err != nil {
		t.Fatalf("Failed to create endpoint schema: %v", err)
	}

	endpoints := []*endpoint.EndpointWithFile{
		{
			Schema:   schema,
			FilePath: "/tmp/test.apimock",
		},
	}

	server := New(endpoints)
	// Manually set AST in cache since we're not loading from file
	server.astCache["/tmp/test.apimock"] = ast

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(server.createHandlerFromEndpoint(endpoints[0])))
	defer ts.Close()

	// Test: First 3 calls should return 200
	for i := 1; i <= 3; i++ {
		resp, err := http.Get(ts.URL)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Request %d: expected status 200, got %d", i, resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		t.Logf("Request %d (200 OK): %s", i, string(body))
	}

	// Test: 4th call onwards should return 429
	for i := 4; i <= 6; i++ {
		resp, err := http.Get(ts.URL)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 429 {
			t.Errorf("Request %d: expected status 429, got %d", i, resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		t.Logf("Request %d (429 Too Many Requests): %s", i, string(body))
	}
}

func TestServer_ConditionEvaluation_RequestHeaders(t *testing.T) {
	// Create .apimock file with header conditions
	ast := &apimock.APIMockFile{
		Request: &apimock.RequestSection{
			Method: "GET",
			Path:   "/api/test-headers",
		},
		Responses: []apimock.ResponseSection{
			{
				StatusCode:  401,
				Description: "Unauthorized",
				Properties: map[string]string{
					"ContentType": "application/json",
				},
				Conditions: []apimock.ConditionLine{
					{
						Expression: apimock.UnaryExpression{
							Operator: "not",
							Operand: apimock.VariableReference{
								Name: "headers",
								AccessPath: []apimock.Access{
									{Type: apimock.IndexAccess, Key: "authorization"},
								},
							},
						},
						IsOrCondition: false,
					},
				},
				Body: `{"error": "Unauthorized"}`,
			},
			{
				StatusCode:  200,
				Description: "OK",
				Properties: map[string]string{
					"ContentType": "application/json",
				},
				Conditions: []apimock.ConditionLine{},
				Body:       `{"message": "Success"}`,
			},
		},
	}

	// Convert AST to EndpointSchema
	schema, err := endpoint.FromAPIMockFile(ast)
	if err != nil {
		t.Fatalf("Failed to create endpoint schema: %v", err)
	}

	endpoints := []*endpoint.EndpointWithFile{
		{
			Schema:   schema,
			FilePath: "/tmp/test_headers.apimock",
		},
	}

	server := New(endpoints)
	server.astCache["/tmp/test_headers.apimock"] = ast

	ts := httptest.NewServer(http.HandlerFunc(server.createHandlerFromEndpoint(endpoints[0])))
	defer ts.Close()

	// Test: Request without authorization header should return 401
	t.Run("Without Authorization Header", func(t *testing.T) {
		resp, err := http.Get(ts.URL)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 401 {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 401, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})

	// Test: Request with authorization header should return 200
	t.Run("With Authorization Header", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.URL, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer token123")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})
}

func TestServer_ConditionEvaluation_QueryParameters(t *testing.T) {
	// Create .apimock file with query parameter conditions
	ast := &apimock.APIMockFile{
		Request: &apimock.RequestSection{
			Method: "GET",
			Path:   "/api/test-query",
		},
		Responses: []apimock.ResponseSection{
			{
				StatusCode:  400,
				Description: "Bad Request",
				Properties: map[string]string{
					"ContentType": "application/json",
				},
				Conditions: []apimock.ConditionLine{
					{
						Expression: apimock.UnaryExpression{
							Operator: "not",
							Operand: apimock.VariableReference{
								Name: "query",
								AccessPath: []apimock.Access{
									{Type: apimock.IndexAccess, Key: "q"},
								},
							},
						},
						IsOrCondition: false,
					},
				},
				Body: `{"error": "Missing query parameter"}`,
			},
			{
				StatusCode:  200,
				Description: "OK",
				Properties: map[string]string{
					"ContentType": "application/json",
				},
				Conditions: []apimock.ConditionLine{},
				Body:       `{"results": []}`,
			},
		},
	}

	schema, err := endpoint.FromAPIMockFile(ast)
	if err != nil {
		t.Fatalf("Failed to create endpoint schema: %v", err)
	}

	endpoints := []*endpoint.EndpointWithFile{
		{
			Schema:   schema,
			FilePath: "/tmp/test_query.apimock",
		},
	}

	server := New(endpoints)
	server.astCache["/tmp/test_query.apimock"] = ast

	ts := httptest.NewServer(http.HandlerFunc(server.createHandlerFromEndpoint(endpoints[0])))
	defer ts.Close()

	// Test: Request without query parameter should return 400
	t.Run("Without Query Parameter", func(t *testing.T) {
		resp, err := http.Get(ts.URL)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 400 {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 400, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})

	// Test: Request with query parameter should return 200
	t.Run("With Query Parameter", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "?q=test")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
		}
	})
}

func TestServer_VariableInterpolation(t *testing.T) {
	// Create .apimock file with variable interpolation
	ast := &apimock.APIMockFile{
		Request: &apimock.RequestSection{
			Method: "GET",
			Path:   "/api/test-interpolation",
		},
		Responses: []apimock.ResponseSection{
			{
				StatusCode:  200,
				Description: "OK",
				Properties: map[string]string{
					"ContentType": "application/json",
				},
				Conditions: []apimock.ConditionLine{},
				Body:       `{"count": {{call_count}}, "message": "Request number {{call_count}}"}`,
			},
		},
	}

	schema, err := endpoint.FromAPIMockFile(ast)
	if err != nil {
		t.Fatalf("Failed to create endpoint schema: %v", err)
	}

	endpoints := []*endpoint.EndpointWithFile{
		{
			Schema:   schema,
			FilePath: "/tmp/test_interpolation.apimock",
		},
	}

	server := New(endpoints)
	server.astCache["/tmp/test_interpolation.apimock"] = ast

	ts := httptest.NewServer(http.HandlerFunc(server.createHandlerFromEndpoint(endpoints[0])))
	defer ts.Close()

	// Test: Verify call_count increments and is interpolated correctly
	expectedBodies := []string{
		`{"count": 1, "message": "Request number 1"}`,
		`{"count": 2, "message": "Request number 2"}`,
		`{"count": 3, "message": "Request number 3"}`,
	}

	for i, expected := range expectedBodies {
		resp, err := http.Get(ts.URL)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		actual := string(body)

		if actual != expected {
			t.Errorf("Request %d: expected body %q, got %q", i+1, expected, actual)
		} else {
			t.Logf("Request %d: ✓ %s", i+1, actual)
		}
	}
}

func TestServer_ComplexConditions_BodyEvaluation(t *testing.T) {
	// Create .apimock file with body content evaluation
	ast := &apimock.APIMockFile{
		Request: &apimock.RequestSection{
			Method: "POST",
			Path:   "/api/test-body",
		},
		Responses: []apimock.ResponseSection{
			{
				StatusCode:  409,
				Description: "Conflict",
				Properties: map[string]string{
					"ContentType": "application/json",
				},
				Conditions: []apimock.ConditionLine{
					{
						Expression: apimock.FunctionCall{
							Target: apimock.VariableReference{Name: "body"},
							Name:   "contains",
							Args: []apimock.Expression{
								apimock.StringValue{Value: "admin@example.com"},
							},
						},
						IsOrCondition: false,
					},
				},
				Body: `{"error": "Email already exists"}`,
			},
			{
				StatusCode:  201,
				Description: "Created",
				Properties: map[string]string{
					"ContentType": "application/json",
				},
				Conditions: []apimock.ConditionLine{},
				Body:       `{"message": "User created"}`,
			},
		},
	}

	schema, err := endpoint.FromAPIMockFile(ast)
	if err != nil {
		t.Fatalf("Failed to create endpoint schema: %v", err)
	}

	endpoints := []*endpoint.EndpointWithFile{
		{
			Schema:   schema,
			FilePath: "/tmp/test_body.apimock",
		},
	}

	server := New(endpoints)
	server.astCache["/tmp/test_body.apimock"] = ast

	ts := httptest.NewServer(http.HandlerFunc(server.createHandlerFromEndpoint(endpoints[0])))
	defer ts.Close()

	// Test: Request with admin@example.com should return 409
	t.Run("With Conflicting Email", func(t *testing.T) {
		body := bytes.NewBufferString(`{"email": "admin@example.com", "name": "Admin"}`)
		resp, err := http.Post(ts.URL, "application/json", body)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 409 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 409, got %d. Body: %s", resp.StatusCode, string(respBody))
		}
	})

	// Test: Request with different email should return 201
	t.Run("With New Email", func(t *testing.T) {
		body := bytes.NewBufferString(`{"email": "user@example.com", "name": "User"}`)
		resp, err := http.Post(ts.URL, "application/json", body)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 201, got %d. Body: %s", resp.StatusCode, string(respBody))
		}
	})
}
