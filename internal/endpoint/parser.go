package endpoint

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func Parse(filePath string) (*EndpointSchema, error) {
	content, err := readFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w", filePath, err)
	}

	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("file '%s' is empty", filePath)
	}

	parts := strings.SplitN(content, "###", 2)
	endpoint := EndpointSchema{}

	if len(parts) > 1 {
		requestSection := strings.TrimSpace(parts[0])
		endpoint = parseEndpoint(requestSection)
	}

	responseContent := content
	if len(parts) > 1 {
		responseContent = "###" + parts[1]
	}

	blocks := strings.SplitSeq(responseContent, "###")

	var responses []Response
	for block := range blocks {
		if strings.TrimSpace(block) == "" {
			continue
		}
		res, err := parseBlock(block)
		if err != nil {
			return nil, fmt.Errorf("failed to parse block: %w", err)
		}
		responses = append(responses, res)
	}

	if len(responses) == 0 {
		return nil, fmt.Errorf("no valid responses found in file '%s'", filePath)
	}

	endpoint.Responses = responses

	return &endpoint, nil
}

func parseBlock(block string) (Response, error) {
	lines := strings.Split(block, "\n")

	response := Response{
		Title:       strings.TrimSpace(lines[0]),
		StatusCode:  200,
		ContentType: "text/plain",
	}

	var bodyLines []string
	parsingBody := false

	for i := 1; i < len(lines); i++ {
		line := lines[i]

		if parsingBody {
			bodyLines = append(bodyLines, line)
			continue
		}

		if strings.TrimSpace(line) == "" {
			parsingBody = true
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			if err := parseHeader(&response, parts); err != nil {
				return response, err
			}
			continue
		}

		parsingBody = true
		bodyLines = append(bodyLines, line)
	}

	response.Body = strings.Join(bodyLines, "\n")

	return response, nil
}

func parseHeader(res *Response, parts []string) error {
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	switch strings.ToLower(key) {
	case "status-code":
		code, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid status code '%s': %w", value, err)
		}
		if code < 100 || code > 599 {
			return fmt.Errorf("status code %d is out of valid range (100-599)", code)
		}
		res.StatusCode = code
	case "content-type":
		if value == "" {
			return fmt.Errorf("content-type cannot be empty")
		}
		res.ContentType = value
	}

	return nil
}

func parseEndpoint(requestSection string) EndpointSchema {
	lines := strings.Split(requestSection, "\n")

	schema := EndpointSchema{
		Route:       "/",
		ContentType: DefaultContentType,
	}

	if len(lines) == 0 {
		return schema
	}

	schema.Route = strings.TrimSpace(lines[0])

	var schemaLines []string
	inSchema := false
	var propertiesMap map[string]string

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if strings.Contains(line, ":") && !inSchema {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if propertiesMap == nil {
					propertiesMap = make(map[string]string)
				}
				propertiesMap[strings.ToLower(key)] = value
			}
			continue
		}
		if line == "" && !inSchema {
			inSchema = true
			continue
		}
		if inSchema {
			schemaLines = append(schemaLines, lines[i])
		}
	}

	if len(schemaLines) > 0 {
		bodyContent := strings.Join(schemaLines, "\n")
		schema.Body = strings.TrimSpace(bodyContent)
	}

	if contentType, ok := propertiesMap[strings.ToLower(ContentTypeHeader)]; ok {
		schema.ContentType = contentType
	}

	return schema
}

func readFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file '%s': %w", filePath, err)
	}
	defer file.Close()

	content := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read file '%s': %w", filePath, err)
	}
	return content, nil
}
