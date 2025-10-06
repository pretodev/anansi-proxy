package apimock

import (
	"fmt"
	"strings"
)

// ConditionValidator validates condition expressions at compile time.
// It checks for syntax errors, unknown functions, invalid arguments, and other issues
// that would cause runtime failures.
type ConditionValidator struct {
	knownFunctions map[string]FunctionSignature
	contextVars    map[string]bool
}

// FunctionSignature defines the expected signature of a built-in function.
type FunctionSignature struct {
	Name           string
	RequiresTarget bool     // true if the function needs a target (e.g., str.split())
	MinArgs        int      // minimum number of arguments
	MaxArgs        int      // maximum number of arguments (-1 for unlimited)
	TargetTypes    []string // acceptable target types: "string", "number", "table", "any"
	Description    string
}

// NewConditionValidator creates a new condition validator with known built-in functions.
func NewConditionValidator() *ConditionValidator {
	v := &ConditionValidator{
		knownFunctions: make(map[string]FunctionSignature),
		contextVars:    make(map[string]bool),
	}

	v.initializeBuiltinFunctions()
	v.initializeContextVariables()

	return v
}

// initializeBuiltinFunctions registers all built-in functions with their signatures.
func (v *ConditionValidator) initializeBuiltinFunctions() {
	functions := []FunctionSignature{
		// String functions
		{Name: "split", RequiresTarget: true, MinArgs: 1, MaxArgs: 1, TargetTypes: []string{"string"}, Description: "splits string by delimiter"},
		{Name: "contains", RequiresTarget: true, MinArgs: 1, MaxArgs: 1, TargetTypes: []string{"string", "table"}, Description: "checks if string/table contains value"},
		{Name: "not_contains", RequiresTarget: true, MinArgs: 1, MaxArgs: 1, TargetTypes: []string{"string", "table"}, Description: "checks if string/table does not contain value"},
		{Name: "matches", RequiresTarget: true, MinArgs: 1, MaxArgs: 1, TargetTypes: []string{"string"}, Description: "matches string against regex"},
		{Name: "upper", RequiresTarget: true, MinArgs: 0, MaxArgs: 0, TargetTypes: []string{"string"}, Description: "converts to uppercase"},
		{Name: "lower", RequiresTarget: true, MinArgs: 0, MaxArgs: 0, TargetTypes: []string{"string"}, Description: "converts to lowercase"},
		{Name: "trim", RequiresTarget: true, MinArgs: 0, MaxArgs: 0, TargetTypes: []string{"string"}, Description: "trims whitespace"},
		{Name: "substring", RequiresTarget: true, MinArgs: 2, MaxArgs: 2, TargetTypes: []string{"string"}, Description: "extracts substring"},

		// Collection functions
		{Name: "len", RequiresTarget: true, MinArgs: 0, MaxArgs: 0, TargetTypes: []string{"string", "table"}, Description: "returns length"},

		// Type checking functions
		{Name: "is_string", RequiresTarget: true, MinArgs: 0, MaxArgs: 0, TargetTypes: []string{"any"}, Description: "checks if value is string"},
		{Name: "is_number", RequiresTarget: true, MinArgs: 0, MaxArgs: 0, TargetTypes: []string{"any"}, Description: "checks if value is number"},
		{Name: "is_boolean", RequiresTarget: true, MinArgs: 0, MaxArgs: 0, TargetTypes: []string{"any"}, Description: "checks if value is boolean"},
		{Name: "is_table", RequiresTarget: true, MinArgs: 0, MaxArgs: 0, TargetTypes: []string{"any"}, Description: "checks if value is table"},

		// Math functions
		{Name: "round", RequiresTarget: true, MinArgs: 0, MaxArgs: 0, TargetTypes: []string{"number"}, Description: "rounds to nearest integer"},
		{Name: "floor", RequiresTarget: true, MinArgs: 0, MaxArgs: 0, TargetTypes: []string{"number"}, Description: "rounds down"},
		{Name: "ceil", RequiresTarget: true, MinArgs: 0, MaxArgs: 0, TargetTypes: []string{"number"}, Description: "rounds up"},
		{Name: "abs", RequiresTarget: true, MinArgs: 0, MaxArgs: 0, TargetTypes: []string{"number"}, Description: "absolute value"},

		// Random functions (global)
		{Name: "random", RequiresTarget: false, MinArgs: 0, MaxArgs: 0, TargetTypes: []string{}, Description: "random float 0-1"},
		{Name: "random_bool", RequiresTarget: false, MinArgs: 0, MaxArgs: 0, TargetTypes: []string{}, Description: "random boolean"},
		{Name: "random_int", RequiresTarget: false, MinArgs: 2, MaxArgs: 2, TargetTypes: []string{}, Description: "random integer in range"},
		{Name: "random_float", RequiresTarget: false, MinArgs: 2, MaxArgs: 2, TargetTypes: []string{}, Description: "random float in range"},
		{Name: "random_choice", RequiresTarget: true, MinArgs: 0, MaxArgs: 0, TargetTypes: []string{"table"}, Description: "random element from array"},
	}

	for _, fn := range functions {
		v.knownFunctions[fn.Name] = fn
	}
}

// initializeContextVariables registers known context variables.
func (v *ConditionValidator) initializeContextVariables() {
	contextVars := []string{
		"call_count",
		"method",
		"path",
		"headers",
		"query",
		"body",
		"timestamp",
		"date",
	}

	for _, varName := range contextVars {
		v.contextVars[varName] = true
	}
}

// ValidateConditionLine validates a single condition line.
func (v *ConditionValidator) ValidateConditionLine(condition ConditionLine) error {
	return v.validateExpression(condition.Expression)
}

// ValidateConditions validates all condition lines in a response.
func (v *ConditionValidator) ValidateConditions(conditions []ConditionLine) error {
	for i, condition := range conditions {
		if err := v.ValidateConditionLine(condition); err != nil {
			return fmt.Errorf("condition line %d: %w", i+1, err)
		}
	}
	return nil
}

// validateExpression recursively validates an expression.
func (v *ConditionValidator) validateExpression(expr Expression) error {
	if expr == nil {
		return fmt.Errorf("expression cannot be nil")
	}

	switch e := expr.(type) {
	case BinaryExpression:
		return v.validateBinaryExpression(e)
	case UnaryExpression:
		return v.validateUnaryExpression(e)
	case Attribution:
		return v.validateAttribution(e)
	case FunctionCall:
		return v.validateFunctionCall(e)
	case VariableReference:
		return v.validateVariableReference(e)
	case NumberValue, BooleanValue, StringValue, TableValue, RangeValue:
		// Literals are always valid
		return nil
	default:
		return fmt.Errorf("unknown expression type: %T", expr)
	}
}

// validateBinaryExpression validates binary operators and their operands.
func (v *ConditionValidator) validateBinaryExpression(expr BinaryExpression) error {
	// Validate left operand
	if err := v.validateExpression(expr.Left); err != nil {
		return fmt.Errorf("left operand: %w", err)
	}

	// Validate right operand
	if err := v.validateExpression(expr.Right); err != nil {
		return fmt.Errorf("right operand: %w", err)
	}

	// Validate operator
	validOperators := map[string]bool{
		"+": true, "-": true, "*": true, "/": true, "%": true, "//": true,
		"==": true, "!=": true, ">": true, "<": true, ">=": true, "<=": true,
		"..": true, "and": true, "or": true,
	}

	if !validOperators[expr.Operator] {
		return fmt.Errorf("unknown binary operator: %s", expr.Operator)
	}

	return nil
}

// validateUnaryExpression validates unary operators and their operands.
func (v *ConditionValidator) validateUnaryExpression(expr UnaryExpression) error {
	// Validate operand
	if err := v.validateExpression(expr.Operand); err != nil {
		return fmt.Errorf("operand: %w", err)
	}

	// Validate operator
	if expr.Operator != "not" {
		return fmt.Errorf("unknown unary operator: %s", expr.Operator)
	}

	return nil
}

// validateAttribution validates attribution expressions.
func (v *ConditionValidator) validateAttribution(expr Attribution) error {
	// Validate value expression
	if err := v.validateExpression(expr.Value); err != nil {
		return fmt.Errorf("attribution value: %w", err)
	}

	// Validate variable names
	if len(expr.Variables) == 0 {
		return fmt.Errorf("attribution requires at least one variable name")
	}

	for _, varName := range expr.Variables {
		if !isValidVariableName(varName) {
			return fmt.Errorf("invalid variable name: %s", varName)
		}
	}

	return nil
}

// validateFunctionCall validates function calls.
func (v *ConditionValidator) validateFunctionCall(expr FunctionCall) error {
	// Check if function exists
	signature, exists := v.knownFunctions[expr.Name]
	if !exists {
		return fmt.Errorf("unknown function: .%s", expr.Name)
	}

	// Validate target
	if signature.RequiresTarget {
		if expr.Target == nil {
			return fmt.Errorf("function .%s requires a target (e.g., value.%s(...))", expr.Name, expr.Name)
		}
		if err := v.validateExpression(expr.Target); err != nil {
			return fmt.Errorf("function .%s target: %w", expr.Name, err)
		}
	} else {
		if expr.Target != nil {
			return fmt.Errorf("function .%s does not take a target (use .%s(...) instead of value.%s(...))", expr.Name, expr.Name, expr.Name)
		}
	}

	// Validate argument count
	argCount := len(expr.Args)
	if argCount < signature.MinArgs {
		return fmt.Errorf("function .%s requires at least %d argument(s), got %d",
			expr.Name, signature.MinArgs, argCount)
	}
	if signature.MaxArgs >= 0 && argCount > signature.MaxArgs {
		return fmt.Errorf("function .%s accepts at most %d argument(s), got %d",
			expr.Name, signature.MaxArgs, argCount)
	}

	// Validate arguments
	for i, arg := range expr.Args {
		if err := v.validateExpression(arg); err != nil {
			return fmt.Errorf("function .%s argument %d: %w", expr.Name, i+1, err)
		}
	}

	return nil
}

// validateVariableReference validates variable references.
func (v *ConditionValidator) validateVariableReference(expr VariableReference) error {
	// We do basic validation - we can't fully validate variable references
	// at compile time since they depend on runtime context
	if expr.Name == "" {
		return fmt.Errorf("variable reference cannot be empty")
	}

	// Check if it's a known context variable (this is a soft check - we don't error for unknowns
	// because variables can be defined via attribution)
	// This is mainly for documentation/warning purposes

	return nil
}

// isValidVariableName checks if a variable name is valid.
func isValidVariableName(name string) bool {
	if name == "" {
		return false
	}

	// Reserved keywords
	reserved := map[string]bool{
		"and": true, "or": true, "not": true,
		"True": true, "False": true,
	}

	if reserved[name] {
		return false
	}

	// Must start with letter or underscore
	first := rune(name[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Rest must be alphanumeric or underscore
	for _, ch := range name[1:] {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_') {
			return false
		}
	}

	return true
}

// GetKnownFunctions returns a list of all known function names.
func (v *ConditionValidator) GetKnownFunctions() []string {
	names := make([]string, 0, len(v.knownFunctions))
	for name := range v.knownFunctions {
		names = append(names, name)
	}
	return names
}

// GetFunctionSignature returns the signature of a function if it exists.
func (v *ConditionValidator) GetFunctionSignature(name string) (FunctionSignature, bool) {
	sig, ok := v.knownFunctions[name]
	return sig, ok
}

// SuggestFunction suggests a similar function name if an unknown function is used.
func (v *ConditionValidator) SuggestFunction(name string) string {
	// Simple Levenshtein-like suggestion
	bestMatch := ""
	minDistance := 999

	for knownFunc := range v.knownFunctions {
		dist := levenshteinDistance(strings.ToLower(name), strings.ToLower(knownFunc))
		if dist < minDistance && dist <= 3 { // Only suggest if distance is small
			minDistance = dist
			bestMatch = knownFunc
		}
	}

	return bestMatch
}

// levenshteinDistance calculates the Levenshtein distance between two strings.
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
