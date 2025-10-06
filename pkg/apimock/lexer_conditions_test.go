package apimock

import (
	"testing"
)

// TestLexer_ConditionLine tests basic condition line tokenization
func TestLexer_ConditionLine(t *testing.T) {
	tests := []struct {
		name              string
		input             []string
		expectedType      TokenType
		expectedExpr      string
		expectedIsOr      bool
		expectedLineCount int
	}{
		{
			name:              "Simple condition",
			input:             []string{"> call_count > 5"},
			expectedType:      TokenConditionLine,
			expectedExpr:      "call_count > 5",
			expectedIsOr:      false,
			expectedLineCount: 1,
		},
		{
			name:              "Empty condition",
			input:             []string{">"},
			expectedType:      TokenConditionLine,
			expectedExpr:      "",
			expectedIsOr:      false,
			expectedLineCount: 1,
		},
		{
			name:              "OR condition",
			input:             []string{"> or method == \"POST\""},
			expectedType:      TokenConditionLine,
			expectedExpr:      "method == \"POST\"",
			expectedIsOr:      true,
			expectedLineCount: 1,
		},
		{
			name:              "Condition with comment",
			input:             []string{"> True # This is a comment"},
			expectedType:      TokenConditionLine,
			expectedExpr:      "True",
			expectedIsOr:      false,
			expectedLineCount: 1,
		},
		{
			name:              "Condition with inline comment",
			input:             []string{"> call_count <= 10 # Rate limit check"},
			expectedType:      TokenConditionLine,
			expectedExpr:      "call_count <= 10",
			expectedIsOr:      false,
			expectedLineCount: 1,
		},
		{
			name: "Multiple conditions (AND)",
			input: []string{
				"> call_count > 5",
				"> method == \"GET\"",
			},
			expectedType:      TokenConditionLine,
			expectedExpr:      "call_count > 5",
			expectedIsOr:      false,
			expectedLineCount: 2,
		},
		{
			name: "Multiple conditions (mixed AND/OR)",
			input: []string{
				"> call_count > 5",
				"> method == \"GET\"",
				"> or headers[\"X-Test\"] == \"true\"",
			},
			expectedType:      TokenConditionLine,
			expectedLineCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Lex()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			conditionTokens := filterTokensByType(tokens, TokenConditionLine)
			if len(conditionTokens) != tt.expectedLineCount {
				t.Errorf("Expected %d condition tokens, got %d", tt.expectedLineCount, len(conditionTokens))
			}

			if len(conditionTokens) > 0 {
				firstToken := conditionTokens[0]
				if firstToken.Type != tt.expectedType {
					t.Errorf("Expected token type %v, got %v", tt.expectedType, firstToken.Type)
				}

				if tt.expectedExpr != "" && firstToken.ConditionExpression != tt.expectedExpr {
					t.Errorf("Expected expression %q, got %q", tt.expectedExpr, firstToken.ConditionExpression)
				}

				if firstToken.IsOrCondition != tt.expectedIsOr {
					t.Errorf("Expected IsOrCondition=%v, got %v", tt.expectedIsOr, firstToken.IsOrCondition)
				}
			}
		})
	}
}

// TestLexer_ConditionInResponse tests conditions within a response section
func TestLexer_ConditionInResponse(t *testing.T) {
	input := []string{
		"-- 429: Too Many Requests",
		"ContentType: application/json",
		"> call_count > 5",
		"",
		`{"error": "Rate limit"}`,
	}

	lexer := NewLexer(input)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Expected token sequence
	expectedTypes := []TokenType{
		TokenResponseStart,
		TokenHeader,
		TokenConditionLine,
		TokenBlankLine,
		TokenBodyLine,
	}

	if len(tokens) != len(expectedTypes) {
		t.Fatalf("Expected %d tokens, got %d", len(expectedTypes), len(tokens))
	}

	for i, expectedType := range expectedTypes {
		if tokens[i].Type != expectedType {
			t.Errorf("Token %d: expected type %v, got %v", i, expectedType, tokens[i].Type)
		}
	}

	// Verify condition token
	conditionToken := tokens[2]
	if conditionToken.ConditionExpression != "call_count > 5" {
		t.Errorf("Expected condition expression 'call_count > 5', got %q", conditionToken.ConditionExpression)
	}
}

// TestLexer_ComplexConditions tests more complex condition scenarios
func TestLexer_ComplexConditions(t *testing.T) {
	input := []string{
		"-- 401: Unauthorized",
		"ContentType: application/json",
		"> headers[\"Authorization\"] >> token",
		"> token >> .split \" \" >> bearer, token_value",
		"> token_value != \"valid-secret-123\"",
		"",
		`{"error": "Unauthorized"}`,
	}

	lexer := NewLexer(input)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	conditionTokens := filterTokensByType(tokens, TokenConditionLine)
	if len(conditionTokens) != 3 {
		t.Fatalf("Expected 3 condition tokens, got %d", len(conditionTokens))
	}

	expectedExpressions := []string{
		`headers["Authorization"] >> token`,
		`token >> .split " " >> bearer, token_value`,
		`token_value != "valid-secret-123"`,
	}

	for i, expected := range expectedExpressions {
		if conditionTokens[i].ConditionExpression != expected {
			t.Errorf("Condition %d: expected %q, got %q", i, expected, conditionTokens[i].ConditionExpression)
		}
	}
}

// TestLexer_ConditionNotConfusedWithHeader tests that conditions don't get confused with headers
func TestLexer_ConditionNotConfusedWithHeader(t *testing.T) {
	input := []string{
		"-- 200: OK",
		"Content-Type: application/json",
		"X-Custom-Header: value",
		"> method == \"GET\"",
		"",
		`{"status": "ok"}`,
	}

	lexer := NewLexer(input)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	headerCount := 0
	conditionCount := 0

	for _, token := range tokens {
		if token.Type == TokenHeader {
			headerCount++
		}
		if token.Type == TokenConditionLine {
			conditionCount++
		}
	}

	if headerCount != 2 {
		t.Errorf("Expected 2 header tokens, got %d", headerCount)
	}

	if conditionCount != 1 {
		t.Errorf("Expected 1 condition token, got %d", conditionCount)
	}
}

// Helper function to filter tokens by type
func filterTokensByType(tokens []Token, tokenType TokenType) []Token {
	filtered := make([]Token, 0)
	for _, token := range tokens {
		if token.Type == tokenType {
			filtered = append(filtered, token)
		}
	}
	return filtered
}
