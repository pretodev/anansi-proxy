package apimock

import (
	"testing"
)

func TestExpressionParser_Numbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"42", "42"},
		{"3.14", "3.14"},
		{"0", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewExpressionParser(tt.input)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, expr.String())
			}
		})
	}
}

func TestExpressionParser_Booleans(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"True", "True"},
		{"False", "False"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewExpressionParser(tt.input)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, expr.String())
			}
		})
	}
}

func TestExpressionParser_Strings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, `"hello"`},
		{`"world"`, `"world"`},
		{`""`, `""`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewExpressionParser(tt.input)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, expr.String())
			}
		})
	}
}

func TestExpressionParser_BinaryExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Addition", "1 + 2", "(1 + 2)"},
		{"Multiplication", "4 * 2", "(4 * 2)"},
		{"Division", "10 / 2", "(10 / 2)"},
		{"Modulo", "10 % 3", "(10 % 3)"},
		{"Integer Division", "10 // 3", "(10 // 3)"},
		{"Comparison Equal", "x == 5", "(x == 5)"},
		{"Comparison Not Equal", "x != 0", "(x != 0)"},
		{"Comparison Greater", "x > 10", "(x > 10)"},
		{"Comparison Less", "x < 5", "(x < 5)"},
		{"Comparison Greater or Equal", "x >= 5", "(x >= 5)"},
		{"Comparison Less or Equal", "x <= 10", "(x <= 10)"},
		{"String Concatenation", `"hello" .. " " .. "world"`, `(("hello" .. " ") .. "world")`},
		{"Logical And", "True and False", "(True and False)"},
		{"Logical Or", "True or False", "(True or False)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExpressionParser(tt.input)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, expr.String())
			}
		})
	}
}

func TestExpressionParser_UnaryExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Not True", "not True", "(not True)"},
		{"Not False", "not False", "(not False)"},
		{"Not Variable", "not active", "(not active)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExpressionParser(tt.input)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, expr.String())
			}
		})
	}
}

func TestExpressionParser_Attribution(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple Attribution", "42 >> x", "(42 >> x)"},
		{"Dictionary Attribution", `{name = "John", age = 30} >> name, age`, `({name = "John", age = 30} >> name, age)`},
		{"Expression Attribution", "(1 + 2) >> result", "((1 + 2) >> result)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExpressionParser(tt.input)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, expr.String())
			}
		})
	}
}

func TestExpressionParser_VariableReference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple Variable", "user", "user"},
		{"Property Access", "user.name", "user.name"},
		{"Index Access", `user["email"]`, `user["email"]`},
		{"Nested Access", "user.address.city", "user.address.city"},
		{"Mixed Access", `user.contacts[0].email`, `user.contacts["0"].email`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExpressionParser(tt.input)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, expr.String())
			}
		})
	}
}

func TestExpressionParser_FunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Global Function No Args", ".random", ".random()"},
		{"Global Function With Args", ".random_int 1 100", ".random_int(1, 100)"},
		{"Method Call No Args", "name.upper", "name.upper()"},
		{"Method Call With Args", "name.substring 0 5", "name.substring(0, 5)"},
		{"Chained Methods", "name.upper.substring 0 5", "name.upper().substring(0, 5)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExpressionParser(tt.input)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, expr.String())
			}
		})
	}
}

func TestExpressionParser_Tables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty Array", "{}", "[]"},
		{"Array With Numbers", "{1, 2, 3}", "[1, 2, 3]"},
		{"Array With Strings", `{"a", "b", "c"}`, `["a", "b", "c"]`},
		{"Dictionary", `{name = "John", age = 30}`, `{name = "John", age = 30}`},
		{"Nested Array", `{1, {2, 3}, 4}`, `[1, [2, 3], 4]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExpressionParser(tt.input)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, expr.String())
			}
		})
	}
}

func TestExpressionParser_Range(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple Range", "1..10", "1..10"},
		{"Float Range", "0.5..1.5", "0.5..1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExpressionParser(tt.input)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, expr.String())
			}
		})
	}
}

func TestExpressionParser_Precedence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Multiplication Before Addition", "1 + 2 * 3", "(1 + (2 * 3))"},
		{"Division Before Subtraction", "10 / 2", "(10 / 2)"},
		{"Parentheses Override", "(1 + 2) * 3", "((1 + 2) * 3)"},
		{"And Before Or", "True or False and True", "(True or (False and True))"},
		{"Comparison Before And", "x > 5 and y < 10", "((x > 5) and (y < 10))"},
		{"String Concat Before Comparison", `"a" .. "b" == "ab"`, `((("a" .. "b") == "ab"))`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExpressionParser(tt.input)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, expr.String())
			}
		})
	}
}

func TestExpressionParser_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"Condition With Attribution",
			`request.method == "POST" and request.body.email >> email`,
			`(((request.method == "POST") and request.body.email) >> email)`,
		},
		{
			"Function Call In Comparison",
			`.random_int 1 100 > 50`,
			`(.random_int(1, 100) > 50)`,
		},
		{
			"Nested Expressions",
			`(x > 0 and x < 10) or (x > 90 and x < 100)`,
			`(((x > 0) and (x < 10)) or ((x > 90) and (x < 100)))`,
		},
		{
			"Complex Attribution",
			`{status = "ok", count = 10} >> response`,
			`({status = "ok", count = 10} >> response)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExpressionParser(tt.input)
			expr, err := parser.Parse()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, expr.String())
			}
		})
	}
}

func TestExpressionParser_EmptyExpression(t *testing.T) {
	parser := NewExpressionParser("")
	expr, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty expression should evaluate to False
	boolVal, ok := expr.(BooleanValue)
	if !ok {
		t.Fatalf("expected BooleanValue, got %T", expr)
	}
	if boolVal.Value != false {
		t.Errorf("expected False, got %v", boolVal.Value)
	}
}

func TestExpressionParser_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Unclosed Parenthesis", "(1 + 2"},
		{"Missing Operand", "1 +"},
		{"Mismatched Brackets", `user["name"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExpressionParser(tt.input)
			_, err := parser.Parse()
			if err == nil {
				t.Errorf("expected error for input %q, got nil", tt.input)
			}
		})
	}
}
