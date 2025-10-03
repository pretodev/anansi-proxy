package apimock

import (
	"fmt"
	"os"
	"strings"
)

// Parser parses .apimock files and builds an AST
type Parser struct {
	filename string
	lines    []string
	lineNum  int
	errors   []string
}

// NewParser creates a new parser for a .apimock file
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

// Parse parses the file and returns the AST using a tokenized input
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
		return nil, fmt.Errorf("expected at least one response section")
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
		return nil, fmt.Errorf("expected at least one response section")
	}

	return ast, nil
}

// parseRequestSection parses the request section using tokens
func (p *Parser) parseRequestSection(tokens []Token, i *int) (*RequestSection, error) {
	req := NewRequestSection()

	if *i >= len(tokens) || tokens[*i].Type != TokenRequestLine {
		return nil, fmt.Errorf("line %d: expected method/path line", tokens[*i].Line)
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
			req.Headers[tok.Key] = tok.Value
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

// parseResponseSection parses a response section using tokens
func (p *Parser) parseResponseSection(tokens []Token, i *int) (ResponseSection, error) {
	resp := NewResponseSection()

	if *i >= len(tokens) || tokens[*i].Type != TokenResponseStart {
		return resp, fmt.Errorf("line %d: invalid response line", tokens[*i].Line)
	}
	resp.StatusCode = tokens[*i].StatusCode
	resp.Description = tokens[*i].Description
	*i++

	// Parse headers
	for *i < len(tokens) {
		tok := tokens[*i]
		if tok.Type == TokenHeader {
			resp.Headers[tok.Key] = tok.Value
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
