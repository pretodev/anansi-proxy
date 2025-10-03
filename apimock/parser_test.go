package main

import (
	"path/filepath"
	"testing"
)

func TestParseSimple(t *testing.T) {
	parser, err := NewParser("examples/simple.apimock")
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Should have no request section
	if ast.Request != nil {
		t.Errorf("Expected no request section, got one")
	}

	// Should have exactly 3 responses
	if len(ast.Responses) != 3 {
		t.Fatalf("Expected 3 responses, got %d", len(ast.Responses))
	}

	resp := ast.Responses[0]
	if resp.StatusCode != 201 {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	if resp.Description != "Text Response" {
		t.Errorf("Expected description 'Text Response', got '%s'", resp.Description)
	}

	if resp.Body == "" {
		t.Error("Expected body content, got empty")
	}
}

func TestParseGetJson(t *testing.T) {
	parser, err := NewParser("examples/get-json.apimock")
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Should have request section
	if ast.Request == nil {
		t.Fatal("Expected request section, got nil")
	}

	req := ast.Request

	// Check method
	if req.Method != "GET" {
		t.Errorf("Expected method GET, got '%s'", req.Method)
	}

	// Check path
	if req.Path != "/cars" {
		t.Errorf("Expected path '/cars', got '%s'", req.Path)
	}

	// Should have at least 1 response
	if len(ast.Responses) == 0 {
		t.Fatal("Expected at least 1 response")
	}
}

func TestParseJson(t *testing.T) {
	parser, err := NewParser("examples/json.apimock")
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Should have request section
	if ast.Request == nil {
		t.Fatal("Expected request section, got nil")
	}

	req := ast.Request

	// Check method
	if req.Method != "POST" {
		t.Errorf("Expected method POST, got '%s'", req.Method)
	}

	// Check path
	if req.Path != "/user" {
		t.Errorf("Expected path '/user', got '%s'", req.Path)
	}

	// Check headers
	if len(req.Headers) == 0 {
		t.Error("Expected headers, got none")
	}

	// Check body schema exists
	if req.BodySchema == "" {
		t.Error("Expected body schema, got empty")
	}

	// Should have responses
	if len(ast.Responses) < 2 {
		t.Errorf("Expected at least 2 responses, got %d", len(ast.Responses))
	}

	// Check first response
	resp1 := ast.Responses[0]
	if resp1.StatusCode != 201 {
		t.Errorf("Expected status 201, got %d", resp1.StatusCode)
	}

	if len(resp1.Headers) == 0 {
		t.Error("Expected response headers, got none")
	}

	if resp1.Body == "" {
		t.Error("Expected response body, got empty")
	}
}

func TestParseQueryPathParams(t *testing.T) {
	parser, err := NewParser("examples/query-path-params.apimock")
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Should have request section
	if ast.Request == nil {
		t.Fatal("Expected request section, got nil")
	}

	req := ast.Request

	// Check method
	if req.Method != "GET" {
		t.Errorf("Expected method GET, got '%s'", req.Method)
	}

	// Check path has multiple segments
	expectedPath := "/api/v1/users/{userId}/posts"
	if req.Path != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, req.Path)
	}

	// Check path segments
	if len(req.PathSegments) == 0 {
		t.Error("Expected path segments, got none")
	}

	// Find the parameter segment
	hasParameter := false
	for _, seg := range req.PathSegments {
		if seg.IsParameter && seg.Name == "userId" {
			hasParameter = true
			break
		}
	}
	if !hasParameter {
		t.Error("Expected to find {userId} parameter in path segments")
	}

	// Check query parameters
	if len(req.QueryParams) == 0 {
		t.Error("Expected query parameters, got none")
	}
}

func TestParseOctetStream(t *testing.T) {
	parser, err := NewParser("examples/octet-stream.apimock")
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Should have request section
	if ast.Request == nil {
		t.Fatal("Expected request section, got nil")
	}

	req := ast.Request

	// Check method
	if req.Method != "PATCH" {
		t.Errorf("Expected method PATCH, got '%s'", req.Method)
	}

	// Check path has parameter
	if !contains(req.Path, "{id}") {
		t.Errorf("Expected path to contain {id}, got '%s'", req.Path)
	}

	// Should have multiple responses
	if len(ast.Responses) < 2 {
		t.Errorf("Expected at least 2 responses, got %d", len(ast.Responses))
	}
}

func TestParseXml(t *testing.T) {
	parser, err := NewParser("examples/xml.apimock")
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Should have request section with XML schema
	if ast.Request == nil {
		t.Fatal("Expected request section, got nil")
	}

	if ast.Request.BodySchema == "" {
		t.Error("Expected XML body schema, got empty")
	}

	// Check if body contains XML
	if !contains(ast.Request.BodySchema, "<xs:schema>") {
		t.Error("Expected XML schema in body")
	}
}

func TestParseYaml(t *testing.T) {
	parser, err := NewParser("examples/yaml.apimock")
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Should have request section
	if ast.Request == nil {
		t.Fatal("Expected request section, got nil")
	}

	// Should have responses with YAML body
	if len(ast.Responses) == 0 {
		t.Fatal("Expected at least 1 response")
	}

	if ast.Responses[0].Body == "" {
		t.Error("Expected YAML response body, got empty")
	}
}

// TestParseAllExamples tests that all example files parse without errors
func TestParseAllExamples(t *testing.T) {
	examples := []string{
		"form.apimock",
		"get-json.apimock",
		"json.apimock",
		"multipart-form-data.apimock",
		"octet-stream.apimock",
		"query-path-params.apimock",
		"raw.apimock",
		"simple.apimock",
		"xml.apimock",
		"yaml.apimock",
	}

	for _, example := range examples {
		t.Run(example, func(t *testing.T) {
			path := filepath.Join("examples", example)
			parser, err := NewParser(path)
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			ast, err := parser.Parse()
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", example, err)
			}

			// Basic validation
			if len(ast.Responses) == 0 {
				t.Errorf("%s: Expected at least 1 response", example)
			}

			// Validate all responses have valid status codes
			for i, resp := range ast.Responses {
				if resp.StatusCode < 100 || resp.StatusCode > 599 {
					t.Errorf("%s: Response %d has invalid status code: %d", example, i, resp.StatusCode)
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr) && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
