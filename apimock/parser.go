package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Parser parses .apimock files and builds an AST
type Parser struct {
	filename string
	lines    []string
	lineNum  int
	errors   []string
}

// Additional regular expressions for parsing (others are in validator.go)
var (
	pathSegmentRegex         = regexp.MustCompile(`/([a-zA-Z0-9_.\-]+|\{[a-zA-Z0-9_.\-]+\})`)
	queryParamRegex          = regexp.MustCompile(`([a-zA-Z0-9_.\-]+)=(\S+)`)
	responseLineCaptureRegex = regexp.MustCompile(`^--\s*(\d{3}):\s*(.*)`)
	propertyCaptureRegex     = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9_.\-]*):\s*(.+)`)
)

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

// Parse parses the file and returns the AST
func (p *Parser) Parse() (*APIMockFile, error) {
	ast := NewAPIMockFile()

	// Skip leading blank lines
	p.skipBlankLines()

	// Optional request section
	if p.lineNum < len(p.lines) && !p.isResponseLine(p.currentLine()) {
		req, err := p.parseRequestSection()
		if err != nil {
			return nil, err
		}
		ast.Request = req
		p.skipBlankLines()
	}

	// At least one response section required
	if p.lineNum >= len(p.lines) {
		return nil, fmt.Errorf("expected at least one response section")
	}

	// Parse all response sections
	for p.lineNum < len(p.lines) {
		p.skipBlankLines()
		if p.lineNum >= len(p.lines) {
			break
		}

		resp, err := p.parseResponseSection()
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

// parseRequestSection parses the request section
func (p *Parser) parseRequestSection() (*RequestSection, error) {
	req := NewRequestSection()

	// Parse method line (method + path + continuations)
	if err := p.parseMethodLine(req); err != nil {
		return nil, err
	}

	// Parse headers (properties)
	for p.lineNum < len(p.lines) {
		line := p.currentLine()
		if p.isBlankLine(line) || p.isResponseLine(line) {
			break
		}

		matches := propertyCaptureRegex.FindStringSubmatch(line)
		if matches != nil {
			key := matches[1]
			value := strings.TrimSpace(matches[2])
			req.Headers[key] = value
			p.lineNum++
		} else {
			// Not a property - must be start of body
			break
		}
	}

	// Parse optional body
	if p.lineNum < len(p.lines) && p.isBlankLine(p.currentLine()) {
		p.lineNum++ // skip blank line
	}

	// Read body lines
	bodyLines := make([]string, 0)
	for p.lineNum < len(p.lines) {
		line := p.currentLine()

		// Stop if we hit a response line
		if p.isResponseLine(line) {
			break
		}

		// Stop if blank lines followed by response
		if p.isBlankLine(line) {
			nextNonBlank := p.lineNum + 1
			for nextNonBlank < len(p.lines) && p.isBlankLine(p.lines[nextNonBlank]) {
				nextNonBlank++
			}
			if nextNonBlank < len(p.lines) && p.isResponseLine(p.lines[nextNonBlank]) {
				break
			}
		}

		bodyLines = append(bodyLines, line)
		p.lineNum++
	}

	if len(bodyLines) > 0 {
		req.BodySchema = strings.Join(bodyLines, "\n")
	}

	return req, nil
}

// parseMethodLine parses the method and path
func (p *Parser) parseMethodLine(req *RequestSection) error {
	if p.lineNum >= len(p.lines) {
		return fmt.Errorf("line %d: expected method line or path", p.lineNum+1)
	}

	line := p.currentLine()

	// Extract HTTP method (optional)
	methodMatches := httpMethodRegex.FindStringSubmatch(line)
	if methodMatches != nil {
		req.Method = strings.TrimSpace(methodMatches[1])
		// Remove method from line to get path
		line = strings.TrimSpace(line[len(methodMatches[0]):])
	}

	// Extract path
	pathMatches := pathStartRegex.FindString(line)
	if pathMatches == "" {
		// If no method was found, try to parse the whole line as path
		if req.Method == "" {
			pathMatches = pathStartRegex.FindString(p.currentLine())
		}
		if pathMatches == "" {
			return fmt.Errorf("line %d: invalid path format", p.lineNum+1)
		}
	}

	req.Path = pathMatches
	req.PathSegments = p.parsePathSegments(pathMatches)

	p.lineNum++

	// Parse path continuations (indented lines)
	for p.lineNum < len(p.lines) {
		line := p.currentLine()
		matches := pathContRegex.FindStringSubmatch(line)
		if matches == nil {
			break
		}

		continuation := strings.TrimSpace(matches[1])

		// Check if it's a path segment
		if strings.HasPrefix(continuation, "/") {
			req.Path += continuation
			segments := p.parsePathSegments(continuation)
			req.PathSegments = append(req.PathSegments, segments...)
		} else if strings.HasPrefix(continuation, "?") || strings.HasPrefix(continuation, "&") {
			// Parse query parameters
			p.parseQueryParams(continuation[1:], req.QueryParams)
		}

		p.lineNum++
	}

	return nil
}

// parsePathSegments parses path into segments
func (p *Parser) parsePathSegments(path string) []PathSegment {
	segments := make([]PathSegment, 0)
	matches := pathSegmentRegex.FindAllStringSubmatch(path, -1)

	for _, match := range matches {
		segment := match[1]
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			// It's a parameter
			paramName := segment[1 : len(segment)-1]
			segments = append(segments, PathSegment{
				Value:       segment,
				IsParameter: true,
				Name:        paramName,
			})
		} else {
			// It's a literal segment
			segments = append(segments, PathSegment{
				Value:       segment,
				IsParameter: false,
			})
		}
	}

	return segments
}

// parseQueryParams parses query parameters
func (p *Parser) parseQueryParams(queryString string, params map[string]string) {
	matches := queryParamRegex.FindAllStringSubmatch(queryString, -1)
	for _, match := range matches {
		key := match[1]
		value := match[2]
		params[key] = value
	}
}

// parseResponseSection parses a response section
func (p *Parser) parseResponseSection() (ResponseSection, error) {
	resp := NewResponseSection()

	if p.lineNum >= len(p.lines) {
		return resp, fmt.Errorf("unexpected end of file")
	}

	// Parse response line: -- 200: Description
	line := p.currentLine()
	matches := responseLineCaptureRegex.FindStringSubmatch(line)
	if matches == nil {
		return resp, fmt.Errorf("line %d: invalid response line format: '%s'", p.lineNum+1, line)
	}

	// Extract status code
	statusCode, err := strconv.Atoi(matches[1])
	if err != nil {
		return resp, fmt.Errorf("line %d: invalid status code: %s", p.lineNum+1, matches[1])
	}
	resp.StatusCode = statusCode
	resp.Description = strings.TrimSpace(matches[2])

	p.lineNum++

	// Parse headers (properties)
	for p.lineNum < len(p.lines) {
		line := p.currentLine()
		if p.isBlankLine(line) {
			break
		}
		if p.isResponseLine(line) {
			// Next response section
			return resp, nil
		}

		propMatches := propertyCaptureRegex.FindStringSubmatch(line)
		if propMatches != nil {
			key := propMatches[1]
			value := strings.TrimSpace(propMatches[2])
			resp.Headers[key] = value
			p.lineNum++
		} else {
			return resp, fmt.Errorf("line %d: invalid response property: '%s'", p.lineNum+1, line)
		}
	}

	// Parse optional response body
	if p.lineNum < len(p.lines) && p.isBlankLine(p.currentLine()) {
		p.lineNum++ // skip blank line
	}

	// Read body lines
	bodyLines := make([]string, 0)
	for p.lineNum < len(p.lines) {
		line := p.currentLine()
		if p.isResponseLine(line) {
			break
		}
		bodyLines = append(bodyLines, line)
		p.lineNum++
	}

	if len(bodyLines) > 0 {
		// Remove trailing blank lines
		for len(bodyLines) > 0 && strings.TrimSpace(bodyLines[len(bodyLines)-1]) == "" {
			bodyLines = bodyLines[:len(bodyLines)-1]
		}
		resp.Body = strings.Join(bodyLines, "\n")
	}

	return resp, nil
}

// Helper methods

func (p *Parser) currentLine() string {
	if p.lineNum >= len(p.lines) {
		return ""
	}
	return p.lines[p.lineNum]
}

func (p *Parser) isBlankLine(line string) bool {
	return strings.TrimSpace(line) == ""
}

func (p *Parser) isResponseLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "--")
}

func (p *Parser) skipBlankLines() {
	for p.lineNum < len(p.lines) && p.isBlankLine(p.currentLine()) {
		p.lineNum++
	}
}
