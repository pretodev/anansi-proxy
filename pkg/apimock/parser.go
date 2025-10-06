package apimock

import (
	"fmt"
	"os"
	"strings"
)

// Parser parses .apimock files and builds an Abstract Syntax Tree (AST).
// It uses a Lexer to tokenize the input and then constructs the APIMockFile structure.
type Parser struct {
	filename string
	lines    []string
	lineNum  int
	errors   []string
}

// NewParser creates a new parser for a .apimock file.
// It reads the file content and prepares it for parsing.
// Returns an error if the file cannot be read.
func NewParser(filename string) (*Parser, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	return &Parser{
		filename: filename,
		lines:    lines,
		lineNum:  0,
		errors:   make([]string, 0),
	}, nil
}

// Parse parses the file and returns the AST.
// It first tokenizes the input using a Lexer, then constructs the APIMockFile structure.
// The file must contain at least one response section.
// Returns an error if the file format is invalid or cannot be parsed.
func (p *Parser) Parse() (*APIMockFile, error) {
	ast := NewAPIMockFile()

	lexer := NewLexer(p.lines)
	tokens, err := lexer.Lex()
	if err != nil {
		return nil, err
	}

	i := 0
	// Skip leading blank lines
	for i < len(tokens) && tokens[i].Type == TokenBlankLine {
		i++
	}

	// Optional request section
	if i < len(tokens) && tokens[i].Type == TokenRequestLine {
		req, err := p.parseRequestSection(tokens, &i)
		if err != nil {
			return nil, err
		}
		ast.Request = req
		// Skip blanks before responses
		for i < len(tokens) && tokens[i].Type == TokenBlankLine {
			i++
		}
	}

	// At least one response section required
	if i >= len(tokens) || tokens[i].Type != TokenResponseStart {
		return nil, NewParseError(p.filename, 0, "expected at least one response section (format: -- CODE: Description)")
	}

	// Parse all response sections
	for i < len(tokens) {
		// Skip blanks
		for i < len(tokens) && tokens[i].Type == TokenBlankLine {
			i++
		}
		if i >= len(tokens) {
			break
		}
		if tokens[i].Type != TokenResponseStart {
			break
		}
		resp, err := p.parseResponseSection(tokens, &i)
		if err != nil {
			return nil, err
		}
		ast.Responses = append(ast.Responses, resp)
	}

	if len(ast.Responses) == 0 {
		return nil, NewParseError(p.filename, 0, "expected at least one response section")
	}

	return ast, nil
}

// parseRequestSection parses the request section using tokens.
// It extracts the HTTP method, path, query parameters, headers, and optional body schema.
// The method continues parsing until it encounters a response section or reaches the end of tokens.
func (p *Parser) parseRequestSection(tokens []Token, i *int) (*RequestSection, error) {
	req := NewRequestSection()

	if *i >= len(tokens) || tokens[*i].Type != TokenRequestLine {
		return nil, NewParseError(p.filename, tokens[*i].Line, "expected HTTP method and/or path (e.g., 'GET /api/users')")
	}

	// Request line
	req.Method = tokens[*i].Method
	req.Path = tokens[*i].Path
	req.PathSegments = append(req.PathSegments, tokens[*i].PathSegments...)
	*i++

	// Consume continuations, query params, headers until body or response start
	inBody := false
	bodyLines := make([]string, 0)
	for *i < len(tokens) {
		tok := tokens[*i]
		switch tok.Type {
		case TokenPathContinuation:
			if inBody {
				goto DONE
			}
			req.Path += tok.PathContinuation
			req.PathSegments = append(req.PathSegments, tok.PathSegments...)
			*i++
		case TokenQueryParam:
			if inBody {
				goto DONE
			}
			req.QueryParams[tok.Key] = tok.Value
			*i++
		case TokenHeader:
			if inBody {
				// Treat as body content if header lines appear after body start
				bodyLines = append(bodyLines, tok.Raw)
				*i++
				continue
			}
			req.Properties[tok.Key] = tok.Value
			*i++
		case TokenBlankLine:
			// Blank line indicates start of body (if any)
			inBody = true
			*i++
		case TokenBodyLine:
			inBody = true
			bodyLines = append(bodyLines, tok.Raw)
			*i++
		case TokenResponseStart:
			// End of request section
			goto DONE
		default:
			// Unknown or EOF
			goto DONE
		}
	}

DONE:
	if len(bodyLines) > 0 {
		req.BodySchema = strings.Join(bodyLines, "\n")
	}
	return req, nil
}

// parseResponseSection parses a response section using tokens.
// It extracts the status code, description, headers, conditions, and body content.
// The order is: response line -> properties -> conditions -> body.
// Trailing blank lines are removed from the response body.
func (p *Parser) parseResponseSection(tokens []Token, i *int) (ResponseSection, error) {
	resp := NewResponseSection()

	if *i >= len(tokens) || tokens[*i].Type != TokenResponseStart {
		return resp, NewParseError(p.filename, tokens[*i].Line, "invalid response line format (expected: -- CODE: Description)")
	}
	resp.StatusCode = tokens[*i].StatusCode
	resp.Description = tokens[*i].Description

	// Validate status code
	if !IsValidHTTPStatusCode(resp.StatusCode) {
		return resp, NewParseError(p.filename, tokens[*i].Line, fmt.Sprintf("invalid HTTP status code: %d (must be between %d-%d)", resp.StatusCode, MinHTTPStatusCode, MaxHTTPStatusCode))
	}

	*i++

	// Parse properties (headers)
	for *i < len(tokens) {
		tok := tokens[*i]
		if tok.Type == TokenHeader {
			resp.Properties[tok.Key] = tok.Value
			*i++
			continue
		}
		break
	}

	// Parse conditions
	for *i < len(tokens) {
		tok := tokens[*i]
		if tok.Type == TokenConditionLine {
			condition, err := p.parseConditionLine(tok)
			if err != nil {
				return resp, err
			}
			resp.Conditions = append(resp.Conditions, condition)
			*i++
			continue
		}
		break
	}

	// Optional blank line
	if *i < len(tokens) && tokens[*i].Type == TokenBlankLine {
		*i++
	}

	// Parse body lines until next response or EOF
	bodyLines := make([]string, 0)
	for *i < len(tokens) {
		tok := tokens[*i]
		if tok.Type == TokenResponseStart {
			break
		}
		if tok.Type == TokenBodyLine || tok.Type == TokenBlankLine || tok.Type == TokenHeader {
			// Treat any content here as body (including header-like lines)
			bodyLines = append(bodyLines, tok.Raw)
			*i++
			continue
		}
		break
	}

	// Remove trailing blank lines from body
	for len(bodyLines) > 0 && strings.TrimSpace(bodyLines[len(bodyLines)-1]) == "" {
		bodyLines = bodyLines[:len(bodyLines)-1]
	}
	if len(bodyLines) > 0 {
		resp.Body = strings.Join(bodyLines, "\n")
	}

	return resp, nil
}

// parseConditionLine parses a condition line token into a ConditionLine AST node.
// It uses the ExpressionParser to parse the condition expression and validates it.
func (p *Parser) parseConditionLine(token Token) (ConditionLine, error) {
	parser := NewExpressionParser(token.ConditionExpression)
	expr, err := parser.Parse()
	if err != nil {
		return ConditionLine{}, NewParseError(p.filename, token.Line, fmt.Sprintf("invalid condition expression: %s", err.Error()))
	}

	// Validate the condition expression
	validator := NewConditionValidator()
	conditionLine := ConditionLine{
		Expression:    expr,
		IsOrCondition: token.IsOrCondition,
		Line:          token.Line,
	}
	
	if err := validator.ValidateConditionLine(conditionLine); err != nil {
		return ConditionLine{}, NewParseError(p.filename, token.Line, fmt.Sprintf("condition validation error: %s", err.Error()))
	}

	return conditionLine, nil
}
