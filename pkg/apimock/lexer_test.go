package apimock

import (
	"testing"
)

func TestLexer_BlankLines(t *testing.T) {
	lines := []string{
		"",
		"  ",
		"\t",
	}
	lexer := NewLexer(lines)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	for i, tok := range tokens {
		if tok.Type != TokenBlankLine {
			t.Errorf("token %d: expected TokenBlankLine, got %v", i, tok.Type)
		}
	}
}

func TestLexer_RequestLine_WithMethod(t *testing.T) {
	lines := []string{"POST /api/users"}
	lexer := NewLexer(lines)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	tok := tokens[0]
	if tok.Type != TokenRequestLine {
		t.Fatalf("expected TokenRequestLine, got %v", tok.Type)
	}
	if tok.Method != "POST" {
		t.Errorf("expected method POST, got %s", tok.Method)
	}
	if tok.Path != "/api/users" {
		t.Errorf("expected path /api/users, got %s", tok.Path)
	}
}

func TestLexer_RequestLine_PathOnly(t *testing.T) {
	lines := []string{"/api/products"}
	lexer := NewLexer(lines)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}
	tok := tokens[0]
	if tok.Type != TokenRequestLine {
		t.Fatalf("expected TokenRequestLine, got %v", tok.Type)
	}
	if tok.Method != "" {
		t.Errorf("expected empty method, got %s", tok.Method)
	}
	if tok.Path != "/api/products" {
		t.Errorf("expected path /api/products, got %s", tok.Path)
	}
}

func TestLexer_PathWithParameters(t *testing.T) {
	lines := []string{"GET /api/users/{id}/posts/{postId}"}
	lexer := NewLexer(lines)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tok := tokens[0]

	// Verify we got path segments
	if len(tok.PathSegments) == 0 {
		t.Fatal("expected path segments, got none")
	}

	// Check that we have parameter segments with correct names
	foundIdParam := false
	foundPostIdParam := false

	for _, seg := range tok.PathSegments {
		if seg.IsParameter && seg.Name == "id" {
			foundIdParam = true
		}
		if seg.IsParameter && seg.Name == "postId" {
			foundPostIdParam = true
		}
	}

	if !foundIdParam {
		t.Error("expected to find {id} parameter in path segments")
	}
	if !foundPostIdParam {
		t.Error("expected to find {postId} parameter in path segments")
	}
}

func TestLexer_QueryParams(t *testing.T) {
	lines := []string{
		"GET /api/search",
		"  ?query=test",
		"  &limit=10",
	}
	lexer := NewLexer(lines)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have: 1 request line + at least 2 query params
	if len(tokens) < 3 {
		t.Fatalf("expected at least 3 tokens, got %d", len(tokens))
	}

	// Find query param tokens
	queryFound := false
	limitFound := false

	for _, tok := range tokens {
		if tok.Type == TokenQueryParam {
			if tok.Key == "query" && tok.Value == "test" {
				queryFound = true
			}
			if tok.Key == "limit" && tok.Value == "10" {
				limitFound = true
			}
		}
	}

	if !queryFound {
		t.Error("expected to find query=test parameter")
	}
	if !limitFound {
		t.Error("expected to find limit=10 parameter")
	}
}

func TestLexer_Headers(t *testing.T) {
	lines := []string{
		"Content-Type: application/json",
		"Authorization: Bearer token123",
	}
	lexer := NewLexer(lines)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(tokens))
	}

	for i, tok := range tokens {
		if tok.Type != TokenHeader {
			t.Errorf("token %d: expected TokenHeader, got %v", i, tok.Type)
		}
	}

	if tokens[0].Key != "Content-Type" || tokens[0].Value != "application/json" {
		t.Errorf("token 0: expected Content-Type: application/json, got %s: %s", tokens[0].Key, tokens[0].Value)
	}

	if tokens[1].Key != "Authorization" || tokens[1].Value != "Bearer token123" {
		t.Errorf("token 1: expected Authorization: Bearer token123, got %s: %s", tokens[1].Key, tokens[1].Value)
	}
}

func TestLexer_ResponseStart(t *testing.T) {
	lines := []string{
		"-- 200: Success",
		"-- 404: Not Found",
		"-- 500:",
	}
	lexer := NewLexer(lines)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}

	expectedResponses := []struct {
		statusCode  int
		description string
	}{
		{200, "Success"},
		{404, "Not Found"},
		{500, ""},
	}

	for i, tok := range tokens {
		if tok.Type != TokenResponseStart {
			t.Errorf("token %d: expected TokenResponseStart, got %v", i, tok.Type)
		}
		expected := expectedResponses[i]
		if tok.StatusCode != expected.statusCode {
			t.Errorf("token %d: expected status code %d, got %d", i, expected.statusCode, tok.StatusCode)
		}
		if tok.Description != expected.description {
			t.Errorf("token %d: expected description '%s', got '%s'", i, expected.description, tok.Description)
		}
	}
}

func TestLexer_BodyLines(t *testing.T) {
	lines := []string{
		"-- 200: OK",
		"",
		`{"message": "success"}`,
		`{"status": "ok"}`,
	}
	lexer := NewLexer(lines)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have: 1 response start, 1 blank, 2 body lines
	if len(tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d", len(tokens))
	}

	if tokens[2].Type != TokenBodyLine {
		t.Errorf("token 2: expected TokenBodyLine, got %v", tokens[2].Type)
	}
	if tokens[3].Type != TokenBodyLine {
		t.Errorf("token 3: expected TokenBodyLine, got %v", tokens[3].Type)
	}
}

func TestLexer_ComplexFile(t *testing.T) {
	lines := []string{
		"POST /api/users/{id}",
		"  ?active=true",
		"Content-Type: application/json",
		"Authorization: Bearer token",
		"",
		`{"name": "John"}`,
		"",
		"-- 201: Created",
		"Location: /api/users/123",
		"",
		`{"id": 123}`,
		"",
		"-- 400: Bad Request",
		"",
		`{"error": "Invalid input"}`,
	}

	lexer := NewLexer(lines)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify we have a reasonable number of tokens
	if len(tokens) < 10 {
		t.Fatalf("expected at least 10 tokens, got %d", len(tokens))
	}

	// Verify first token is request line
	if tokens[0].Type != TokenRequestLine {
		t.Errorf("first token should be TokenRequestLine, got %v", tokens[0].Type)
	}
	if tokens[0].Method != "POST" {
		t.Errorf("expected POST method, got %s", tokens[0].Method)
	}
}
