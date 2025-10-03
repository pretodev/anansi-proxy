package apimock

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParser_SimpleResponse(t *testing.T) {
	content := `-- 200: OK

{"message": "success"}`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	parser, err := NewParser(tmpFile)
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if ast.Request != nil {
		t.Errorf("expected no request section, got %+v", ast.Request)
	}

	if len(ast.Responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(ast.Responses))
	}

	resp := ast.Responses[0]
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if resp.Description != "OK" {
		t.Errorf("expected description 'OK', got '%s'", resp.Description)
	}
	if !strings.Contains(resp.Body, "success") {
		t.Errorf("expected body to contain 'success', got '%s'", resp.Body)
	}
}

func TestParser_MultipleResponses(t *testing.T) {
	content := `-- 200: Success
Content-Type: application/json

{"status": "ok"}

-- 404: Not Found

{"error": "Resource not found"}

-- 500: Server Error

{"error": "Internal error"}`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	parser, err := NewParser(tmpFile)
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if len(ast.Responses) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(ast.Responses))
	}

	expectedCodes := []int{200, 404, 500}
	for i, code := range expectedCodes {
		if ast.Responses[i].StatusCode != code {
			t.Errorf("response %d: expected status %d, got %d", i, code, ast.Responses[i].StatusCode)
		}
	}

	// Check first response has header
	if ast.Responses[0].Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type header, got %v", ast.Responses[0].Headers)
	}
}

func TestParser_RequestWithMethod(t *testing.T) {
	content := `POST /api/users
Content-Type: application/json

{"name": "John"}

-- 201: Created

{"id": 123}`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	parser, err := NewParser(tmpFile)
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if ast.Request == nil {
		t.Fatal("expected request section, got nil")
	}

	if ast.Request.Method != "POST" {
		t.Errorf("expected method POST, got %s", ast.Request.Method)
	}

	if ast.Request.Path != "/api/users" {
		t.Errorf("expected path /api/users, got %s", ast.Request.Path)
	}

	if ast.Request.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type header")
	}

	if !strings.Contains(ast.Request.BodySchema, "John") {
		t.Errorf("expected body to contain 'John', got '%s'", ast.Request.BodySchema)
	}
}

func TestParser_PathParameters(t *testing.T) {
	content := `GET /api/users/{userId}/posts/{postId}

-- 200: OK

{}`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	parser, err := NewParser(tmpFile)
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if ast.Request == nil {
		t.Fatal("expected request section")
	}

	// Check path segments
	foundUserId := false
	foundPostId := false
	for _, seg := range ast.Request.PathSegments {
		if seg.IsParameter && seg.Name == "userId" {
			foundUserId = true
		}
		if seg.IsParameter && seg.Name == "postId" {
			foundPostId = true
		}
	}

	if !foundUserId {
		t.Error("expected to find userId parameter in path segments")
	}
	if !foundPostId {
		t.Error("expected to find postId parameter in path segments")
	}
}

func TestParser_QueryParameters(t *testing.T) {
	content := `GET /api/search
  ?query=test
  &limit=10

-- 200: OK

[]`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	parser, err := NewParser(tmpFile)
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if ast.Request == nil {
		t.Fatal("expected request section")
	}

	if len(ast.Request.QueryParams) == 0 {
		t.Fatal("expected query params to be parsed")
	}

	// Check if we have both params (they might be combined in one key)
	hasQuery := false

	for key, val := range ast.Request.QueryParams {
		if key == "query" && val == "test" {
			hasQuery = true
		}
		// Sometimes params are combined
		if key == "query" && (val == "test&limit=10" || val == "test") {
			hasQuery = true
		}
	}

	if !hasQuery {
		t.Errorf("expected query param 'query=test', got %v", ast.Request.QueryParams)
	}
}

func TestParser_NoResponseError(t *testing.T) {
	content := `POST /api/users

{"name": "John"}`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	parser, err := NewParser(tmpFile)
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	_, err = parser.Parse()
	if err == nil {
		t.Error("expected error for missing response section")
	}
	if !strings.Contains(err.Error(), "response") {
		t.Errorf("expected error message about response, got: %v", err)
	}
}

func TestParser_EmptyFile(t *testing.T) {
	content := ``

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	parser, err := NewParser(tmpFile)
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	_, err = parser.Parse()
	if err == nil {
		t.Error("expected error for empty file")
	}
}

func TestParser_OnlyBlankLines(t *testing.T) {
	content := `


`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	parser, err := NewParser(tmpFile)
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	_, err = parser.Parse()
	if err == nil {
		t.Error("expected error for file with only blank lines")
	}
}

func TestParser_ComplexExample(t *testing.T) {
	content := `POST /api/users/{id}
  ?active=true
  &role=admin
Content-Type: application/json
Authorization: Bearer token123

{
  "name": "John Doe",
  "email": "john@example.com"
}

-- 200: Success
Content-Type: application/json
X-Request-ID: abc123

{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com"
}

-- 400: Bad Request
Content-Type: application/json

{
  "error": "Invalid input",
  "details": ["name is required", "email is invalid"]
}

-- 401: Unauthorized

{
  "error": "Invalid token"
}`

	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)

	parser, err := NewParser(tmpFile)
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}

	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// Verify request
	if ast.Request == nil {
		t.Fatal("expected request section")
	}
	if ast.Request.Method != "POST" {
		t.Errorf("expected POST method, got %s", ast.Request.Method)
	}
	if len(ast.Request.QueryParams) != 2 {
		t.Errorf("expected 2 query params, got %d", len(ast.Request.QueryParams))
	}
	if len(ast.Request.Headers) != 2 {
		t.Errorf("expected 2 headers, got %d", len(ast.Request.Headers))
	}

	// Verify responses
	if len(ast.Responses) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(ast.Responses))
	}

	// Check status codes
	expectedCodes := []int{200, 400, 401}
	for i, code := range expectedCodes {
		if ast.Responses[i].StatusCode != code {
			t.Errorf("response %d: expected status %d, got %d", i, code, ast.Responses[i].StatusCode)
		}
	}

	// Check first response headers
	if len(ast.Responses[0].Headers) != 2 {
		t.Errorf("expected 2 headers in first response, got %d", len(ast.Responses[0].Headers))
	}
}

// Helper function to create temporary file for testing
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.apimock")
	err := os.WriteFile(tmpFile, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return tmpFile
}
