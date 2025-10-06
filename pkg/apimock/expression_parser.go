package apimock

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ExpressionParser parses condition expressions into AST nodes.
// It implements a recursive descent parser with operator precedence.
type ExpressionParser struct {
	input  string
	pos    int
	tokens []string
}

// NewExpressionParser creates a new expression parser for the given input.
func NewExpressionParser(input string) *ExpressionParser {
	return &ExpressionParser{
		input:  input,
		pos:    0,
		tokens: tokenizeExpression(input),
	}
}

// Parse parses the expression and returns an AST node.
func (p *ExpressionParser) Parse() (Expression, error) {
	if len(p.tokens) == 0 {
		// Empty condition evaluates to False
		return BooleanValue{Value: false}, nil
	}
	return p.parseLogicalOr()
}

// parseLogicalOr parses 'or' expressions (lowest precedence).
func (p *ExpressionParser) parseLogicalOr() (Expression, error) {
	left, err := p.parseLogicalAnd()
	if err != nil {
		return nil, err
	}

	for p.peek() == "or" {
		p.consume() // consume 'or'
		right, err := p.parseLogicalAnd()
		if err != nil {
			return nil, err
		}
		left = BinaryExpression{
			Left:     left,
			Operator: "or",
			Right:    right,
		}
	}

	return left, nil
}

// parseLogicalAnd parses 'and' expressions.
func (p *ExpressionParser) parseLogicalAnd() (Expression, error) {
	left, err := p.parseAttribution()
	if err != nil {
		return nil, err
	}

	for p.peek() == "and" {
		p.consume() // consume 'and'
		right, err := p.parseAttribution()
		if err != nil {
			return nil, err
		}
		left = BinaryExpression{
			Left:     left,
			Operator: "and",
			Right:    right,
		}
	}

	return left, nil
}

// parseAttribution parses attribution expressions (>>).
func (p *ExpressionParser) parseAttribution() (Expression, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	if p.peek() == ">>" {
		p.consume() // consume '>>'

		// Parse variable list (supports destructuring)
		variables := make([]string, 0)
		for {
			varName := p.peek()
			if !isIdentifier(varName) {
				return nil, fmt.Errorf("expected variable name after >>, got %q", varName)
			}
			variables = append(variables, varName)
			p.consume()

			if p.peek() == "," {
				p.consume() // consume ','
				continue
			}
			break
		}

		return Attribution{
			Value:     left,
			Variables: variables,
		}, nil
	}

	return left, nil
}

// parseComparison parses comparison expressions (==, !=, >, <, >=, <=).
func (p *ExpressionParser) parseComparison() (Expression, error) {
	left, err := p.parseStringConcat()
	if err != nil {
		return nil, err
	}

	op := p.peek()
	if isComparisonOp(op) {
		p.consume()
		right, err := p.parseStringConcat()
		if err != nil {
			return nil, err
		}
		return BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
		}, nil
	}

	return left, nil
}

// parseStringConcat parses string concatenation (..).
func (p *ExpressionParser) parseStringConcat() (Expression, error) {
	left, err := p.parseAddSub()
	if err != nil {
		return nil, err
	}

	for p.peek() == ".." {
		p.consume() // consume '..'
		right, err := p.parseAddSub()
		if err != nil {
			return nil, err
		}
		left = BinaryExpression{
			Left:     left,
			Operator: "..",
			Right:    right,
		}
	}

	return left, nil
}

// parseAddSub parses addition and subtraction.
func (p *ExpressionParser) parseAddSub() (Expression, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return nil, err
	}

	for {
		op := p.peek()
		if op != "+" && op != "-" {
			break
		}
		p.consume()
		right, err := p.parseMulDiv()
		if err != nil {
			return nil, err
		}
		left = BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left, nil
}

// parseMulDiv parses multiplication, division, and modulo.
func (p *ExpressionParser) parseMulDiv() (Expression, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for {
		op := p.peek()
		if op != "*" && op != "/" && op != "%" && op != "//" {
			break
		}
		p.consume()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left, nil
}

// parseUnary parses unary expressions (not).
func (p *ExpressionParser) parseUnary() (Expression, error) {
	if p.peek() == "not" {
		p.consume() // consume 'not'
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return UnaryExpression{
			Operator: "not",
			Operand:  operand,
		}, nil
	}

	return p.parsePrimary()
}

// parsePrimary parses primary expressions (values, variables, function calls, parentheses).
func (p *ExpressionParser) parsePrimary() (Expression, error) {
	token := p.peek()

	// Parentheses
	if token == "(" {
		p.consume() // consume '('
		expr, err := p.parseLogicalOr()
		if err != nil {
			return nil, err
		}
		if p.peek() != ")" {
			return nil, fmt.Errorf("expected ')', got %q", p.peek())
		}
		p.consume() // consume ')'
		return expr, nil
	}

	// Numbers
	if isNumber(token) {
		p.consume()
		val, _ := strconv.ParseFloat(token, 64)
		expr := NumberValue{Value: val}
		return p.parseFunctionCall(expr)
	}

	// Booleans
	if token == "True" {
		p.consume()
		expr := BooleanValue{Value: true}
		return p.parseFunctionCall(expr)
	}
	if token == "False" {
		p.consume()
		expr := BooleanValue{Value: false}
		return p.parseFunctionCall(expr)
	}

	// Strings
	if isString(token) {
		p.consume()
		// Remove quotes
		val := token[1 : len(token)-1]
		expr := StringValue{Value: val}
		return p.parseFunctionCall(expr)
	}

	// Tables
	if token == "{" {
		return p.parseTable()
	}

	// Ranges (e.g., 1..10)
	if p.isRange() {
		return p.parseRange()
	}

	// Global function calls (e.g., .random_int 1 100)
	if strings.HasPrefix(token, ".") {
		return p.parseGlobalFunction()
	}

	// Variables and identifiers
	if isIdentifier(token) {
		return p.parseVariableOrFunction()
	}

	return nil, fmt.Errorf("unexpected token: %q", token)
}

// parseFunctionCall checks if there's a function call after a value.
func (p *ExpressionParser) parseFunctionCall(target Expression) (Expression, error) {
	// Check for method call (e.g., value.function_name)
	if p.peek() == "." {
		// Check next token to see if it's a method name
		if p.pos+1 >= len(p.tokens) {
			return target, nil
		}

		nextToken := p.tokens[p.pos+1]
		// If next token is a dot (for .. operator), don't treat as method call
		if nextToken == "." {
			return target, nil
		}

		// If next token starts with dot, it's a global function, not a method
		if strings.HasPrefix(nextToken, ".") {
			return target, nil
		}

		// If next token is an identifier, it's a method call
		if isIdentifier(nextToken) {
			p.consume() // consume '.'
			funcName := p.peek()
			p.consume() // consume function name

			// Parse function arguments
			args := make([]Expression, 0)
			for !p.isEnd() && !isOperator(p.peek()) && p.peek() != "," && p.peek() != ">>" && p.peek() != ")" && p.peek() != "]" {
				arg, err := p.parsePrimary()
				if err != nil {
					break
				}
				args = append(args, arg)
			}

			funcCall := FunctionCall{
				Target: target,
				Name:   funcName,
				Args:   args,
			}

			// Check for chained method calls
			return p.parseFunctionCall(funcCall)
		}
	}

	return target, nil
}

// parseVariableOrFunction parses a variable reference or function call.
func (p *ExpressionParser) parseVariableOrFunction() (Expression, error) {
	name := p.peek()
	p.consume()

	// Build variable reference with access path
	varRef := VariableReference{
		Name:       name,
		AccessPath: make([]Access, 0),
	}

	// Parse property/index access
	for {
		next := p.peek()
		if next == "." {
			// Peek ahead to see if it's range operator (..)
			if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1] == "." {
				// It's a range operator, stop parsing access path
				break
			}

			// Peek ahead to see if next is an identifier (property access)
			if p.pos+1 < len(p.tokens) && isIdentifier(p.tokens[p.pos+1]) {
				p.consume() // consume '.'
				prop := p.peek()
				p.consume()
				varRef.AccessPath = append(varRef.AccessPath, Access{
					Type: PropertyAccess,
					Key:  prop,
				})
			} else {
				// Not a property access, break
				break
			}
		} else if next == "[" {
			p.consume() // consume '['
			key := p.peek()

			// Support numeric index
			var keyVal string
			if isString(key) {
				// Remove quotes from key
				keyVal = key[1 : len(key)-1]
			} else if isNumber(key) {
				// Keep number as string
				keyVal = key
			} else {
				return nil, fmt.Errorf("expected string or number in bracket notation, got %q", key)
			}

			p.consume()
			if p.peek() != "]" {
				return nil, fmt.Errorf("expected ']', got %q", p.peek())
			}
			p.consume() // consume ']'
			varRef.AccessPath = append(varRef.AccessPath, Access{
				Type: IndexAccess,
				Key:  keyVal,
			})
		} else {
			break
		}
	}

	// Check for function call on the variable
	return p.parseFunctionCall(varRef)
}

// parseGlobalFunction parses a global function call (e.g., .random_int 1 100).
func (p *ExpressionParser) parseGlobalFunction() (Expression, error) {
	token := p.peek()
	funcName := token[1:] // Remove the dot
	p.consume()

	// Parse function arguments
	args := make([]Expression, 0)
	for !p.isEnd() && !isOperator(p.peek()) && p.peek() != "," && p.peek() != ">>" && p.peek() != ")" && p.peek() != "]" {
		arg, err := p.parsePrimary()
		if err != nil {
			break
		}
		args = append(args, arg)
	}

	return FunctionCall{
		Target: nil,
		Name:   funcName,
		Args:   args,
	}, nil
}

// parseTable parses a table (array or dictionary).
func (p *ExpressionParser) parseTable() (Expression, error) {
	p.consume() // consume '{'

	elements := make([]Value, 0)
	dictPairs := make(map[string]Value)
	isDict := false

	for p.peek() != "}" && !p.isEnd() {
		// Check if it's a dictionary (key = value)
		if isIdentifier(p.peek()) {
			// Look ahead to check for '='
			if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1] == "=" {
				isDict = true
				key := p.peek()
				p.consume() // consume key
				p.consume() // consume '='

				val, err := p.parseLogicalOr()
				if err != nil {
					return nil, err
				}

				valValue, ok := val.(Value)
				if !ok {
					return nil, fmt.Errorf("table value must be a Value type")
				}
				dictPairs[key] = valValue

				if p.peek() == "," {
					p.consume() // consume ','
				}
				continue
			}
		}

		// Array element
		val, err := p.parseLogicalOr()
		if err != nil {
			return nil, err
		}
		valValue, ok := val.(Value)
		if !ok {
			return nil, fmt.Errorf("table element must be a Value type")
		}
		elements = append(elements, valValue)

		if p.peek() == "," {
			p.consume() // consume ','
		}
	}

	if p.peek() != "}" {
		return nil, fmt.Errorf("expected '}', got %q", p.peek())
	}
	p.consume() // consume '}'

	if isDict {
		return TableValue{
			IsArray: false,
			Dict:    dictPairs,
		}, nil
	}

	return TableValue{
		IsArray: true,
		Array:   elements,
	}, nil
}

// parseRange parses a range expression (e.g., 1..10).
func (p *ExpressionParser) parseRange() (Expression, error) {
	startToken := p.peek()
	start, _ := strconv.ParseFloat(startToken, 64)
	p.consume()
	p.consume() // consume '..'
	endToken := p.peek()
	end, _ := strconv.ParseFloat(endToken, 64)
	p.consume()

	return RangeValue{
		Start: start,
		End:   end,
	}, nil
}

// Helper methods

func (p *ExpressionParser) peek() string {
	if p.pos >= len(p.tokens) {
		return ""
	}
	return p.tokens[p.pos]
}

func (p *ExpressionParser) peekNext() string {
	if p.pos+1 >= len(p.tokens) {
		return ""
	}
	return p.tokens[p.pos+1]
}

func (p *ExpressionParser) consume() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func (p *ExpressionParser) isEnd() bool {
	return p.pos >= len(p.tokens)
}

func (p *ExpressionParser) isRange() bool {
	return p.pos+2 < len(p.tokens) &&
		isNumber(p.tokens[p.pos]) &&
		p.tokens[p.pos+1] == ".." &&
		isNumber(p.tokens[p.pos+2])
}

// Tokenization

var tokenPatterns = []struct {
	pattern string
	isRegex bool
}{
	{`"[^"]*"`, true},                // Strings
	{`\d+\.?\d*`, true},              // Numbers
	{`[a-zA-Z_][a-zA-Z0-9_]*`, true}, // Identifiers
	{`>>`, false},                    // Attribution
	{`==`, false},                    // Equal
	{`!=`, false},                    // Not equal
	{`>=`, false},                    // Greater equal
	{`<=`, false},                    // Less equal
	{`\.\.`, true},                   // Range/concat operator
	{`//`, false},                    // Integer division
	{`\.\w+`, true},                  // Global function (.func)
	{`[+\-*/%(),\[\]{}.<>=]`, true},  // Single char operators
}

func tokenizeExpression(input string) []string {
	tokens := make([]string, 0)
	i := 0

	for i < len(input) {
		// Skip whitespace
		if input[i] == ' ' || input[i] == '\t' {
			i++
			continue
		}

		// Try two-char operators first
		if i+1 < len(input) {
			twoChar := input[i : i+2]
			if twoChar == ">>" || twoChar == "==" || twoChar == "!=" ||
				twoChar == ">=" || twoChar == "<=" || twoChar == ".." || twoChar == "//" {
				tokens = append(tokens, twoChar)
				i += 2
				continue
			}
		}

		// Try string literal
		if input[i] == '"' {
			end := i + 1
			for end < len(input) && input[end] != '"' {
				end++
			}
			if end < len(input) {
				tokens = append(tokens, input[i:end+1])
				i = end + 1
				continue
			}
		}

		// Try number (including negative)
		if input[i] >= '0' && input[i] <= '9' {
			end := i
			for end < len(input) && ((input[end] >= '0' && input[end] <= '9') || input[end] == '.') {
				end++
			}
			tokens = append(tokens, input[i:end])
			i = end
			continue
		}

		// Try global function (.func_name)
		if input[i] == '.' && i+1 < len(input) && ((input[i+1] >= 'a' && input[i+1] <= 'z') || (input[i+1] >= 'A' && input[i+1] <= 'Z') || input[i+1] == '_') {
			end := i + 1
			for end < len(input) && ((input[end] >= 'a' && input[end] <= 'z') || (input[end] >= 'A' && input[end] <= 'Z') || (input[end] >= '0' && input[end] <= '9') || input[end] == '_') {
				end++
			}
			tokens = append(tokens, input[i:end])
			i = end
			continue
		}

		// Try identifier or keyword
		if (input[i] >= 'a' && input[i] <= 'z') || (input[i] >= 'A' && input[i] <= 'Z') || input[i] == '_' {
			end := i
			for end < len(input) && ((input[end] >= 'a' && input[end] <= 'z') || (input[end] >= 'A' && input[end] <= 'Z') || (input[end] >= '0' && input[end] <= '9') || input[end] == '_') {
				end++
			}
			tokens = append(tokens, input[i:end])
			i = end
			continue
		}

		// Single character operators
		if input[i] == '+' || input[i] == '-' || input[i] == '*' || input[i] == '/' || input[i] == '%' ||
			input[i] == '(' || input[i] == ')' || input[i] == ',' || input[i] == '[' || input[i] == ']' ||
			input[i] == '{' || input[i] == '}' || input[i] == '.' || input[i] == '<' || input[i] == '>' || input[i] == '=' {
			tokens = append(tokens, string(input[i]))
			i++
			continue
		}

		// Unknown character, skip
		i++
	}

	return tokens
}

func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func isString(s string) bool {
	return len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"'
}

func isIdentifier(s string) bool {
	if s == "" || s == "and" || s == "or" || s == "not" || s == "True" || s == "False" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, s)
	return matched
}

func isComparisonOp(s string) bool {
	return s == "==" || s == "!=" || s == ">" || s == "<" || s == ">=" || s == "<="
}

func isOperator(s string) bool {
	operators := map[string]bool{
		"+": true, "-": true, "*": true, "/": true, "%": true, "//": true,
		"==": true, "!=": true, ">": true, "<": true, ">=": true, "<=": true,
		"..": true, "and": true, "or": true, "not": true, ">>": true,
	}
	return operators[s]
}
