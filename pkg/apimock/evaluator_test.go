package apimock

import (
	"testing"
)

func TestEvaluator_NumberValues(t *testing.T) {
	ctx := NewExecutionContext(0, nil)
	eval := NewEvaluator(ctx)

	tests := []struct {
		name     string
		expr     Expression
		expected interface{}
	}{
		{"Integer", NumberValue{Value: 42}, 42.0},
		{"Float", NumberValue{Value: 3.14}, 3.14},
		{"Zero", NumberValue{Value: 0}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Evaluate(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_BooleanValues(t *testing.T) {
	ctx := NewExecutionContext(0, nil)
	eval := NewEvaluator(ctx)

	tests := []struct {
		name     string
		expr     Expression
		expected bool
	}{
		{"True", BooleanValue{Value: true}, true},
		{"False", BooleanValue{Value: false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Evaluate(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_StringValues(t *testing.T) {
	ctx := NewExecutionContext(0, nil)
	eval := NewEvaluator(ctx)

	tests := []struct {
		name     string
		expr     Expression
		expected string
	}{
		{"Simple string", StringValue{Value: "hello"}, "hello"},
		{"Empty string", StringValue{Value: ""}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Evaluate(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_ArithmeticOperators(t *testing.T) {
	ctx := NewExecutionContext(0, nil)
	eval := NewEvaluator(ctx)

	tests := []struct {
		name     string
		expr     Expression
		expected float64
	}{
		{
			"Addition",
			BinaryExpression{
				Left:     NumberValue{Value: 10},
				Operator: "+",
				Right:    NumberValue{Value: 5},
			},
			15.0,
		},
		{
			"Subtraction",
			BinaryExpression{
				Left:     NumberValue{Value: 10},
				Operator: "-",
				Right:    NumberValue{Value: 3},
			},
			7.0,
		},
		{
			"Multiplication",
			BinaryExpression{
				Left:     NumberValue{Value: 6},
				Operator: "*",
				Right:    NumberValue{Value: 7},
			},
			42.0,
		},
		{
			"Division",
			BinaryExpression{
				Left:     NumberValue{Value: 20},
				Operator: "/",
				Right:    NumberValue{Value: 4},
			},
			5.0,
		},
		{
			"Modulo",
			BinaryExpression{
				Left:     NumberValue{Value: 10},
				Operator: "%",
				Right:    NumberValue{Value: 3},
			},
			1.0,
		},
		{
			"Integer Division",
			BinaryExpression{
				Left:     NumberValue{Value: 10},
				Operator: "//",
				Right:    NumberValue{Value: 3},
			},
			3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Evaluate(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_ComparisonOperators(t *testing.T) {
	ctx := NewExecutionContext(0, nil)
	eval := NewEvaluator(ctx)

	tests := []struct {
		name     string
		expr     Expression
		expected bool
	}{
		{
			"Equal - true",
			BinaryExpression{
				Left:     NumberValue{Value: 5},
				Operator: "==",
				Right:    NumberValue{Value: 5},
			},
			true,
		},
		{
			"Equal - false",
			BinaryExpression{
				Left:     NumberValue{Value: 5},
				Operator: "==",
				Right:    NumberValue{Value: 10},
			},
			false,
		},
		{
			"Not Equal - true",
			BinaryExpression{
				Left:     NumberValue{Value: 5},
				Operator: "!=",
				Right:    NumberValue{Value: 10},
			},
			true,
		},
		{
			"Greater Than - true",
			BinaryExpression{
				Left:     NumberValue{Value: 10},
				Operator: ">",
				Right:    NumberValue{Value: 5},
			},
			true,
		},
		{
			"Less Than - true",
			BinaryExpression{
				Left:     NumberValue{Value: 3},
				Operator: "<",
				Right:    NumberValue{Value: 8},
			},
			true,
		},
		{
			"Greater or Equal - true",
			BinaryExpression{
				Left:     NumberValue{Value: 10},
				Operator: ">=",
				Right:    NumberValue{Value: 10},
			},
			true,
		},
		{
			"Less or Equal - true",
			BinaryExpression{
				Left:     NumberValue{Value: 5},
				Operator: "<=",
				Right:    NumberValue{Value: 10},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Evaluate(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_LogicalOperators(t *testing.T) {
	ctx := NewExecutionContext(0, nil)
	eval := NewEvaluator(ctx)

	tests := []struct {
		name     string
		expr     Expression
		expected bool
	}{
		{
			"AND - true",
			BinaryExpression{
				Left:     BooleanValue{Value: true},
				Operator: "and",
				Right:    BooleanValue{Value: true},
			},
			true,
		},
		{
			"AND - false",
			BinaryExpression{
				Left:     BooleanValue{Value: true},
				Operator: "and",
				Right:    BooleanValue{Value: false},
			},
			false,
		},
		{
			"OR - true",
			BinaryExpression{
				Left:     BooleanValue{Value: true},
				Operator: "or",
				Right:    BooleanValue{Value: false},
			},
			true,
		},
		{
			"OR - false",
			BinaryExpression{
				Left:     BooleanValue{Value: false},
				Operator: "or",
				Right:    BooleanValue{Value: false},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Evaluate(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_UnaryOperators(t *testing.T) {
	ctx := NewExecutionContext(0, nil)
	eval := NewEvaluator(ctx)

	tests := []struct {
		name     string
		expr     Expression
		expected bool
	}{
		{
			"NOT true",
			UnaryExpression{
				Operator: "not",
				Operand:  BooleanValue{Value: true},
			},
			false,
		},
		{
			"NOT false",
			UnaryExpression{
				Operator: "not",
				Operand:  BooleanValue{Value: false},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Evaluate(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_StringConcatenation(t *testing.T) {
	ctx := NewExecutionContext(0, nil)
	eval := NewEvaluator(ctx)

	expr := BinaryExpression{
		Left:     StringValue{Value: "hello"},
		Operator: "..",
		Right:    StringValue{Value: " world"},
	}

	result, err := eval.Evaluate(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "hello world"
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestEvaluator_Attribution(t *testing.T) {
	ctx := NewExecutionContext(0, nil)
	eval := NewEvaluator(ctx)

	// Test simple attribution
	expr := Attribution{
		Value:     NumberValue{Value: 42},
		Variables: []string{"answer"},
	}

	result, err := eval.Evaluate(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 42.0 {
		t.Errorf("expected 42.0, got %v", result)
	}

	// Check that variable was set
	if ctx.Variables["answer"] != 42.0 {
		t.Errorf("expected variable 'answer' to be 42.0, got %v", ctx.Variables["answer"])
	}
}

func TestEvaluator_VariableReference(t *testing.T) {
	ctx := NewExecutionContext(5, nil)
	eval := NewEvaluator(ctx)

	// Test call_count
	expr := VariableReference{Name: "call_count"}
	result, err := eval.Evaluate(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 5 {
		t.Errorf("expected 5, got %v", result)
	}

	// Test user-defined variable
	ctx.Variables["myvar"] = "test value"
	expr = VariableReference{Name: "myvar"}
	result, err = eval.Evaluate(expr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "test value" {
		t.Errorf("expected 'test value', got %v", result)
	}
}

func TestEvaluator_RequestContext(t *testing.T) {
	reqCtx := &RequestContext{
		Method: "POST",
		Path:   "/api/users",
		Query: map[string]string{
			"page": "1",
		},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"name":  "John",
			"email": "john@example.com",
		},
	}

	ctx := NewExecutionContext(1, reqCtx)
	eval := NewEvaluator(ctx)

	tests := []struct {
		name     string
		expr     VariableReference
		expected interface{}
	}{
		{
			"request.method",
			VariableReference{
				Name:       "request",
				AccessPath: []Access{{Type: PropertyAccess, Key: "method"}},
			},
			"POST",
		},
		{
			"request.path",
			VariableReference{
				Name:       "request",
				AccessPath: []Access{{Type: PropertyAccess, Key: "path"}},
			},
			"/api/users",
		},
		{
			"request.query[\"page\"]",
			VariableReference{
				Name: "request",
				AccessPath: []Access{
					{Type: PropertyAccess, Key: "query"},
					{Type: IndexAccess, Key: "page"},
				},
			},
			"1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Evaluate(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_StringFunctions(t *testing.T) {
	ctx := NewExecutionContext(0, nil)
	eval := NewEvaluator(ctx)

	tests := []struct {
		name     string
		funcCall FunctionCall
		expected interface{}
	}{
		{
			"upper",
			FunctionCall{
				Target: StringValue{Value: "hello"},
				Name:   "upper",
				Args:   []Expression{},
			},
			"HELLO",
		},
		{
			"lower",
			FunctionCall{
				Target: StringValue{Value: "WORLD"},
				Name:   "lower",
				Args:   []Expression{},
			},
			"world",
		},
		{
			"trim",
			FunctionCall{
				Target: StringValue{Value: "  test  "},
				Name:   "trim",
				Args:   []Expression{},
			},
			"test",
		},
		{
			"contains - true",
			FunctionCall{
				Target: StringValue{Value: "hello world"},
				Name:   "contains",
				Args:   []Expression{StringValue{Value: "world"}},
			},
			true,
		},
		{
			"contains - false",
			FunctionCall{
				Target: StringValue{Value: "hello world"},
				Name:   "contains",
				Args:   []Expression{StringValue{Value: "foo"}},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Evaluate(tt.funcCall)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_CollectionFunctions(t *testing.T) {
	ctx := NewExecutionContext(0, nil)
	eval := NewEvaluator(ctx)

	tests := []struct {
		name     string
		funcCall FunctionCall
		expected interface{}
	}{
		{
			"len - string",
			FunctionCall{
				Target: StringValue{Value: "hello"},
				Name:   "len",
				Args:   []Expression{},
			},
			5.0,
		},
		{
			"len - array",
			FunctionCall{
				Target: TableValue{
					IsArray: true,
					Array: []Value{
						NumberValue{Value: 1},
						NumberValue{Value: 2},
						NumberValue{Value: 3},
					},
				},
				Name: "len",
				Args: []Expression{},
			},
			3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := eval.Evaluate(tt.funcCall)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestEvaluator_EvaluateConditions(t *testing.T) {
	tests := []struct {
		name       string
		ctx        *ExecutionContext
		conditions []ConditionLine
		expected   bool
	}{
		{
			"Empty conditions - true",
			NewExecutionContext(0, nil),
			[]ConditionLine{},
			true,
		},
		{
			"Single true condition",
			NewExecutionContext(5, nil),
			[]ConditionLine{
				{
					Expression: BinaryExpression{
						Left:     VariableReference{Name: "call_count"},
						Operator: ">",
						Right:    NumberValue{Value: 3},
					},
					IsOrCondition: false,
				},
			},
			true,
		},
		{
			"Single false condition",
			NewExecutionContext(2, nil),
			[]ConditionLine{
				{
					Expression: BinaryExpression{
						Left:     VariableReference{Name: "call_count"},
						Operator: ">",
						Right:    NumberValue{Value: 5},
					},
					IsOrCondition: false,
				},
			},
			false,
		},
		{
			"Multiple AND conditions - all true",
			NewExecutionContext(5, nil),
			[]ConditionLine{
				{
					Expression: BinaryExpression{
						Left:     VariableReference{Name: "call_count"},
						Operator: ">",
						Right:    NumberValue{Value: 3},
					},
					IsOrCondition: false,
				},
				{
					Expression: BinaryExpression{
						Left:     VariableReference{Name: "call_count"},
						Operator: "<",
						Right:    NumberValue{Value: 10},
					},
					IsOrCondition: false,
				},
			},
			true,
		},
		{
			"Multiple AND conditions - one false",
			NewExecutionContext(15, nil),
			[]ConditionLine{
				{
					Expression: BinaryExpression{
						Left:     VariableReference{Name: "call_count"},
						Operator: ">",
						Right:    NumberValue{Value: 3},
					},
					IsOrCondition: false,
				},
				{
					Expression: BinaryExpression{
						Left:     VariableReference{Name: "call_count"},
						Operator: "<",
						Right:    NumberValue{Value: 10},
					},
					IsOrCondition: false,
				},
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eval := NewEvaluator(tt.ctx)
			result, err := eval.EvaluateConditions(tt.conditions)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
