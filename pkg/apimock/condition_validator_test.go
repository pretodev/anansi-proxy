package apimock

import (
	"strings"
	"testing"
)

func TestConditionValidator_ValidFunctions(t *testing.T) {
	validator := NewConditionValidator()

	tests := []struct {
		name       string
		expression string
		wantError  bool
	}{
		// Global random functions (these work without targets)
		{
			name:       "random_int global",
			expression: `.random_int 1 10`,
			wantError:  false,
		},
		{
			name:       "random_float global",
			expression: `.random_float 0.0 1.0`,
			wantError:  false,
		},
		{
			name:       "random_bool global",
			expression: `.random_bool`,
			wantError:  false,
		},
		{
			name:       "random_int assigned to variable",
			expression: `.random_int 1 10 >> dice`,
			wantError:  false,
		},

		// Valid operators
		{
			name:       "comparison operators",
			expression: `call_count > 5`,
			wantError:  false,
		},
		{
			name:       "logical operators",
			expression: `True and False`,
			wantError:  false,
		},
		{
			name:       "arithmetic operators",
			expression: `5 + 3 * 2`,
			wantError:  false,
		},
		{
			name:       "variable access",
			expression: `body.email`,
			wantError:  false,
		},
		{
			name:       "index access",
			expression: `headers["Authorization"]`,
			wantError:  false,
		},
		{
			name:       "not operator",
			expression: `not body.value`,
			wantError:  false,
		},
		{
			name:       "simple attribution",
			expression: `call_count >> count`,
			wantError:  false,
		},
		{
			name:       "destructuring attribution",
			expression: `result >> x, y, z`,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExpressionParser(tt.expression)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("Failed to parse expression: %v", err)
			}

			err = validator.validateExpression(expr)
			if (err != nil) != tt.wantError {
				t.Errorf("validateExpression() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestConditionValidator_UnknownFunctions(t *testing.T) {
	validator := NewConditionValidator()

	tests := []struct {
		name          string
		expression    string
		expectedError string
	}{
		{
			name:          "unknown global function",
			expression:    `.unknown 1 2`,
			expectedError: "unknown function: .unknown",
		},
		{
			name:          "typo in global function",
			expression:    `.randm_int 1 10`,
			expectedError: "unknown function: .randm_int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExpressionParser(tt.expression)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("Failed to parse expression: %v", err)
			}

			err = validator.validateExpression(expr)
			if err == nil {
				t.Fatalf("Expected error but got nil")
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}

func TestConditionValidator_InvalidArgumentCount(t *testing.T) {
	validator := NewConditionValidator()

	tests := []struct {
		name          string
		expression    string
		expectedError string
	}{
		{
			name:          "random_int with no arguments",
			expression:    `.random_int`,
			expectedError: "requires at least 2 argument(s), got 0",
		},
		{
			name:          "random_int with one argument",
			expression:    `.random_int 10`,
			expectedError: "requires at least 2 argument(s), got 1",
		},
		{
			name:          "random_int with three arguments",
			expression:    `.random_int 1 10 20`,
			expectedError: "accepts at most 2 argument(s), got 3",
		},
		{
			name:          "random_bool with arguments",
			expression:    `.random_bool True`,
			expectedError: "accepts at most 0 argument(s), got 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExpressionParser(tt.expression)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("Failed to parse expression: %v", err)
			}

			err = validator.validateExpression(expr)
			if err == nil {
				t.Fatalf("Expected error but got nil")
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}

func TestConditionValidator_InvalidTargetUsage(t *testing.T) {
	validator := NewConditionValidator()

	tests := []struct {
		name          string
		expression    string
		expectedError string
	}{
		{
			name:          "split without target (global call)",
			expression:    `.split " "`,
			expectedError: "requires a target",
		},
		{
			name:          "contains without target (global call)",
			expression:    `.contains "test"`,
			expectedError: "requires a target",
		},
		{
			name:          "trim without target (global call)",
			expression:    `.trim`,
			expectedError: "requires a target",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExpressionParser(tt.expression)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("Failed to parse expression: %v", err)
			}

			err = validator.validateExpression(expr)
			if err == nil {
				t.Fatalf("Expected error but got nil")
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}

func TestConditionValidator_InvalidOperators(t *testing.T) {
	validator := NewConditionValidator()

	tests := []struct {
		name       string
		expression Expression
		wantError  bool
		errorMsg   string
	}{
		{
			name: "unknown binary operator",
			expression: BinaryExpression{
				Left:     NumberValue{Value: 5},
				Operator: "???",
				Right:    NumberValue{Value: 3},
			},
			wantError: true,
			errorMsg:  "unknown binary operator: ???",
		},
		{
			name: "unknown unary operator",
			expression: UnaryExpression{
				Operator: "negate",
				Operand:  BooleanValue{Value: true},
			},
			wantError: true,
			errorMsg:  "unknown unary operator: negate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateExpression(tt.expression)
			if (err != nil) != tt.wantError {
				t.Errorf("validateExpression() error = %v, wantError %v", err, tt.wantError)
			}

			if tt.wantError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
			}
		})
	}
}

func TestConditionValidator_InvalidVariableNames(t *testing.T) {
	validator := NewConditionValidator()

	tests := []struct {
		name          string
		expression    Expression
		expectedError string
	}{
		{
			name: "reserved keyword as variable",
			expression: Attribution{
				Value:     NumberValue{Value: 42},
				Variables: []string{"and"},
			},
			expectedError: "invalid variable name: and",
		},
		{
			name: "reserved True as variable",
			expression: Attribution{
				Value:     NumberValue{Value: 42},
				Variables: []string{"True"},
			},
			expectedError: "invalid variable name: True",
		},
		{
			name: "empty variable name",
			expression: Attribution{
				Value:     NumberValue{Value: 42},
				Variables: []string{""},
			},
			expectedError: "invalid variable name:",
		},
		{
			name: "variable starting with number",
			expression: Attribution{
				Value:     NumberValue{Value: 42},
				Variables: []string{"2fast"},
			},
			expectedError: "invalid variable name: 2fast",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateExpression(tt.expression)
			if err == nil {
				t.Fatalf("Expected error but got nil")
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}

func TestConditionValidator_ComplexExpressions(t *testing.T) {
	validator := NewConditionValidator()

	tests := []struct {
		name       string
		expression string
		wantError  bool
	}{
		{
			name:       "complex arithmetic",
			expression: `(5 + 3) * 2 - 4 / 2`,
			wantError:  false,
		},
		{
			name:       "logical combinations",
			expression: `call_count > 5 and (method == "POST" or method == "PUT")`,
			wantError:  false,
		},
		{
			name:       "random with attribution",
			expression: `.random_int 1 100 >> random_value`,
			wantError:  false,
		},
		{
			name:       "chained comparisons",
			expression: `body.age >= 18 and body.age <= 65`,
			wantError:  false,
		},
		{
			name:       "nested property access",
			expression: `body.user.address.city == "Salvador"`,
			wantError:  false,
		},
		{
			name:       "index access with comparison",
			expression: `headers["Content-Type"] == "application/json"`,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExpressionParser(tt.expression)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("Failed to parse expression: %v", err)
			}

			err = validator.validateExpression(expr)
			if (err != nil) != tt.wantError {
				t.Errorf("validateExpression() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestConditionValidator_ValidateConditions(t *testing.T) {
	validator := NewConditionValidator()

	tests := []struct {
		name       string
		conditions []ConditionLine
		wantError  bool
		errorMsg   string
	}{
		{
			name: "all valid conditions",
			conditions: []ConditionLine{
				{
					Expression:    BinaryExpression{Left: VariableReference{Name: "call_count"}, Operator: ">", Right: NumberValue{Value: 5}},
					IsOrCondition: false,
					Line:          1,
				},
				{
					Expression:    BinaryExpression{Left: VariableReference{Name: "method"}, Operator: "==", Right: StringValue{Value: "POST"}},
					IsOrCondition: false,
					Line:          2,
				},
			},
			wantError: false,
		},
		{
			name: "one invalid condition",
			conditions: []ConditionLine{
				{
					Expression:    BinaryExpression{Left: VariableReference{Name: "call_count"}, Operator: ">", Right: NumberValue{Value: 5}},
					IsOrCondition: false,
					Line:          1,
				},
				{
					Expression: FunctionCall{
						Target: nil,
						Name:   "invalid_func",
						Args:   []Expression{},
					},
					IsOrCondition: false,
					Line:          2,
				},
			},
			wantError: true,
			errorMsg:  "unknown function: .invalid_func",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateConditions(tt.conditions)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateConditions() error = %v, wantError %v", err, tt.wantError)
			}

			if tt.wantError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
			}
		})
	}
}

func TestConditionValidator_FunctionSuggestions(t *testing.T) {
	validator := NewConditionValidator()

	tests := []struct {
		input    string
		expected string
	}{
		{"splt", "split"},
		{"trimm", "trim"},
		{"conains", "contains"},
		{"randm_int", "random_int"},
		{"totally_wrong_xyz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			suggestion := validator.SuggestFunction(tt.input)
			if suggestion != tt.expected {
				t.Errorf("SuggestFunction(%q) = %q, want %q", tt.input, suggestion, tt.expected)
			}
		})
	}
}

func TestIsValidVariableName(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{"valid identifier", "my_var", true},
		{"valid with numbers", "var123", true},
		{"valid underscore prefix", "_private", true},
		{"invalid starts with number", "123var", false},
		{"invalid reserved and", "and", false},
		{"invalid reserved or", "or", false},
		{"invalid reserved not", "not", false},
		{"invalid reserved True", "True", false},
		{"invalid reserved False", "False", false},
		{"invalid empty", "", false},
		{"invalid special chars", "var-name", false},
		{"valid camelCase", "myVariable", true},
		{"valid PascalCase", "MyVariable", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidVariableName(tt.varName)
			if result != tt.expected {
				t.Errorf("isValidVariableName(%q) = %v, want %v", tt.varName, result, tt.expected)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"split", "splt", 1},
		{"contains", "conains", 1},
		{"random", "randm", 1},
		{"kitten", "sitting", 3},
	}

	for _, tt := range tests {
		t.Run(tt.s1+"_"+tt.s2, func(t *testing.T) {
			result := levenshteinDistance(tt.s1, tt.s2)
			if result != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, result, tt.expected)
			}
		})
	}
}
