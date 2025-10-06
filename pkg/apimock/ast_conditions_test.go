package apimock

import (
	"testing"
)

// TestConditionLine_Structure tests the ConditionLine structure
func TestConditionLine_Structure(t *testing.T) {
	expr := BooleanValue{Value: true}
	condition := ConditionLine{
		Expression:    expr,
		IsOrCondition: false,
		Line:          5,
	}

	if condition.Expression == nil {
		t.Error("Expected Expression to be set")
	}

	if condition.IsOrCondition {
		t.Error("Expected IsOrCondition to be false")
	}

	if condition.Line != 5 {
		t.Errorf("Expected Line to be 5, got %d", condition.Line)
	}
}

// TestBinaryExpression_String tests binary expression string representation
func TestBinaryExpression_String(t *testing.T) {
	tests := []struct {
		name     string
		expr     BinaryExpression
		expected string
	}{
		{
			name: "Addition",
			expr: BinaryExpression{
				Left:     NumberValue{Value: 5},
				Operator: "+",
				Right:    NumberValue{Value: 3},
			},
			expected: "(5 + 3)",
		},
		{
			name: "Comparison",
			expr: BinaryExpression{
				Left:     VariableReference{Name: "call_count"},
				Operator: ">",
				Right:    NumberValue{Value: 10},
			},
			expected: "(call_count > 10)",
		},
		{
			name: "String concatenation",
			expr: BinaryExpression{
				Left:     StringValue{Value: "Hello"},
				Operator: "..",
				Right:    StringValue{Value: "World"},
			},
			expected: `("Hello" .. "World")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.expr.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestUnaryExpression_String tests unary expression string representation
func TestUnaryExpression_String(t *testing.T) {
	expr := UnaryExpression{
		Operator: "not",
		Operand:  BooleanValue{Value: true},
	}

	expected := "(not True)"
	result := expr.String()

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestAttribution_String tests attribution string representation
func TestAttribution_String(t *testing.T) {
	tests := []struct {
		name     string
		attr     Attribution
		expected string
	}{
		{
			name: "Single variable",
			attr: Attribution{
				Value:     NumberValue{Value: 42},
				Variables: []string{"answer"},
			},
			expected: "42 >> [answer]",
		},
		{
			name: "Destructuring",
			attr: Attribution{
				Value:     VariableReference{Name: "result"},
				Variables: []string{"x", "y", "z"},
			},
			expected: "result >> [x y z]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.attr.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestNumberValue tests NumberValue type
func TestNumberValue(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected string
	}{
		{"Integer", 42, "42"},
		{"Float", 3.14, "3.14"},
		{"Negative", -17, "-17"},
		{"Zero", 0, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num := NumberValue{Value: tt.value}
			if num.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, num.String())
			}
			if num.GetValue() != tt.value {
				t.Errorf("Expected value %v, got %v", tt.value, num.GetValue())
			}
		})
	}
}

// TestBooleanValue tests BooleanValue type
func TestBooleanValue(t *testing.T) {
	trueVal := BooleanValue{Value: true}
	if trueVal.String() != "True" {
		t.Errorf("Expected 'True', got %q", trueVal.String())
	}
	if !trueVal.GetValue().(bool) {
		t.Error("Expected true value")
	}

	falseVal := BooleanValue{Value: false}
	if falseVal.String() != "False" {
		t.Errorf("Expected 'False', got %q", falseVal.String())
	}
	if falseVal.GetValue().(bool) {
		t.Error("Expected false value")
	}
}

// TestStringValue tests StringValue type
func TestStringValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"Simple", "hello", `"hello"`},
		{"Empty", "", `""`},
		{"With spaces", "hello world", `"hello world"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := StringValue{Value: tt.value}
			if str.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, str.String())
			}
			if str.GetValue() != tt.value {
				t.Errorf("Expected value %q, got %q", tt.value, str.GetValue())
			}
		})
	}
}

// TestTableValue tests TableValue type
func TestTableValue(t *testing.T) {
	t.Run("Array table", func(t *testing.T) {
		table := TableValue{
			IsArray: true,
			Array:   []Value{NumberValue{Value: 1}, NumberValue{Value: 2}, NumberValue{Value: 3}},
		}

		if table.String() != "{array[3]}" {
			t.Errorf("Expected '{array[3]}', got %q", table.String())
		}

		if len(table.Array) != 3 {
			t.Errorf("Expected 3 elements, got %d", len(table.Array))
		}
	})

	t.Run("Dictionary table", func(t *testing.T) {
		table := TableValue{
			IsArray: false,
			Dict: map[string]Value{
				"name": StringValue{Value: "Silas"},
				"age":  NumberValue{Value: 30},
			},
		}

		if table.String() != "{dict[2]}" {
			t.Errorf("Expected '{dict[2]}', got %q", table.String())
		}

		if len(table.Dict) != 2 {
			t.Errorf("Expected 2 elements, got %d", len(table.Dict))
		}
	})
}

// TestRangeValue tests RangeValue type
func TestRangeValue(t *testing.T) {
	tests := []struct {
		name     string
		start    float64
		end      float64
		expected string
	}{
		{"1 to 10", 1, 10, "1..10"},
		{"0 to 100", 0, 100, "0..100"},
		{"10 to 20", 10, 20, "10..20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RangeValue{Start: tt.start, End: tt.end}
			if r.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, r.String())
			}
		})
	}
}

// TestVariableReference tests VariableReference type
func TestVariableReference(t *testing.T) {
	tests := []struct {
		name     string
		varRef   VariableReference
		expected string
	}{
		{
			name:     "Simple variable",
			varRef:   VariableReference{Name: "call_count"},
			expected: "call_count",
		},
		{
			name: "Property access",
			varRef: VariableReference{
				Name:       "body",
				AccessPath: []Access{{Type: PropertyAccess, Key: "email"}},
			},
			expected: "body.email",
		},
		{
			name: "Index access",
			varRef: VariableReference{
				Name:       "headers",
				AccessPath: []Access{{Type: IndexAccess, Key: "Authorization"}},
			},
			expected: `headers["Authorization"]`,
		},
		{
			name: "Nested access",
			varRef: VariableReference{
				Name: "body",
				AccessPath: []Access{
					{Type: PropertyAccess, Key: "user"},
					{Type: PropertyAccess, Key: "profile"},
					{Type: PropertyAccess, Key: "name"},
				},
			},
			expected: "body.user.profile.name",
		},
		{
			name: "Mixed access",
			varRef: VariableReference{
				Name: "data",
				AccessPath: []Access{
					{Type: PropertyAccess, Key: "items"},
					{Type: IndexAccess, Key: "0"},
					{Type: PropertyAccess, Key: "value"},
				},
			},
			expected: `data.items["0"].value`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.varRef.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestFunctionCall tests FunctionCall type
func TestFunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		funcCall FunctionCall
		expected string
	}{
		{
			name: "Method call with target",
			funcCall: FunctionCall{
				Target: VariableReference{Name: "email"},
				Name:   "split",
				Args:   []Expression{StringValue{Value: "@"}},
			},
			expected: "email.split(1 args)",
		},
		{
			name: "Global function",
			funcCall: FunctionCall{
				Target: nil,
				Name:   "random_int",
				Args:   []Expression{NumberValue{Value: 1}, NumberValue{Value: 100}},
			},
			expected: ".random_int(2 args)",
		},
		{
			name: "Function without args",
			funcCall: FunctionCall{
				Target: VariableReference{Name: "text"},
				Name:   "trim",
				Args:   []Expression{},
			},
			expected: "text.trim(0 args)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.funcCall.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestResponseSection_WithConditions tests ResponseSection with conditions
func TestResponseSection_WithConditions(t *testing.T) {
	resp := NewResponseSection()
	resp.StatusCode = 429
	resp.Description = "Too Many Requests"
	resp.Conditions = []ConditionLine{
		{
			Expression: BinaryExpression{
				Left:     VariableReference{Name: "call_count"},
				Operator: ">",
				Right:    NumberValue{Value: 5},
			},
			IsOrCondition: false,
			Line:          10,
		},
	}

	if len(resp.Conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(resp.Conditions))
	}

	if resp.Conditions[0].Line != 10 {
		t.Errorf("Expected condition line 10, got %d", resp.Conditions[0].Line)
	}

	if resp.Conditions[0].IsOrCondition {
		t.Error("Expected IsOrCondition to be false")
	}
}
