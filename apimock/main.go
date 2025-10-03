package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "validate":
		validateCommand(os.Args[2])
	case "parse":
		parseCommand(os.Args[2])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("╔════════════════════════════════════════════╗")
	fmt.Println("║   APIMock - API Mock File Tool (EBNF)     ║")
	fmt.Println("╚════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  apimock validate <directory>   Validate all .apimock files in directory")
	fmt.Println("  apimock parse <file>           Parse and display AST for a .apimock file")
	fmt.Println("  apimock help                   Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  apimock validate ./examples")
	fmt.Println("  apimock parse ./examples/json.apimock")
}

func validateCommand(dir string) {
	fmt.Println("╔════════════════════════════════════════════╗")
	fmt.Println("║   APIMock Grammar Validator (EBNF)        ║")
	fmt.Println("╚════════════════════════════════════════════╝")
	fmt.Println()

	valid, invalid, err := ValidateDirectory(dir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("╔════════════════════════════════════════════╗")
	fmt.Printf("║  Total: %d files                           \n", valid+invalid)
	fmt.Printf("║  ✅ Valid: %d                              \n", valid)
	fmt.Printf("║  ❌ Invalid: %d                            \n", invalid)
	fmt.Println("╚════════════════════════════════════════════╝")

	if invalid > 0 {
		os.Exit(1)
	}
}

func parseCommand(filename string) {
	fmt.Println("╔════════════════════════════════════════════╗")
	fmt.Println("║   APIMock Parser - AST Display            ║")
	fmt.Println("╚════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Printf("📄 Parsing: %s\n\n", filepath.Base(filename))

	parser, err := NewParser(filename)
	if err != nil {
		fmt.Printf("❌ Error reading file: %v\n", err)
		os.Exit(1)
	}

	ast, err := parser.Parse()
	if err != nil {
		fmt.Printf("❌ Parse error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Successfully parsed!")
	fmt.Println()

	// Display AST summary
	displayAST(ast)

	// Optionally output JSON
	if len(os.Args) > 3 && os.Args[3] == "--json" {
		fmt.Println("\n" + strings.Repeat("─", 50))
		fmt.Println("JSON Output:")
		fmt.Println(strings.Repeat("─", 50))
		jsonData, err := json.MarshalIndent(ast, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
	}
}

func displayAST(ast *APIMockFile) {
	fmt.Println("╔════════════════════════════════════════════╗")
	fmt.Println("║  Abstract Syntax Tree (AST)                ║")
	fmt.Println("╚════════════════════════════════════════════╝")
	fmt.Println()

	// Request Section
	if ast.Request != nil {
		fmt.Println("📤 REQUEST SECTION")
		fmt.Println(strings.Repeat("─", 50))

		if ast.Request.Method != "" {
			fmt.Printf("  Method: %s\n", ast.Request.Method)
		} else {
			fmt.Println("  Method: (not specified)")
		}

		fmt.Printf("  Path: %s\n", ast.Request.Path)

		// Path segments
		if len(ast.Request.PathSegments) > 0 {
			fmt.Println("  Path Segments:")
			for i, seg := range ast.Request.PathSegments {
				if seg.IsParameter {
					fmt.Printf("    [%d] {%s} (parameter)\n", i, seg.Name)
				} else {
					fmt.Printf("    [%d] %s (literal)\n", i, seg.Value)
				}
			}
		}

		// Query params
		if len(ast.Request.QueryParams) > 0 {
			fmt.Println("  Query Parameters:")
			for key, value := range ast.Request.QueryParams {
				fmt.Printf("    %s = %s\n", key, value)
			}
		}

		// Headers
		if len(ast.Request.Headers) > 0 {
			fmt.Println("  Headers:")
			for key, value := range ast.Request.Headers {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}

		// Body
		if ast.Request.BodySchema != "" {
			bodyPreview := ast.Request.BodySchema
			if len(bodyPreview) > 100 {
				bodyPreview = bodyPreview[:100] + "..."
			}
			fmt.Printf("  Body Schema: %d bytes\n", len(ast.Request.BodySchema))
			fmt.Printf("    Preview: %s\n", bodyPreview)
		}

		fmt.Println()
	}

	// Response Sections
	fmt.Printf("📥 RESPONSE SECTIONS (%d)\n", len(ast.Responses))
	fmt.Println(strings.Repeat("─", 50))

	for i, resp := range ast.Responses {
		fmt.Printf("\n  Response #%d:\n", i+1)
		fmt.Printf("    Status: %d %s\n", resp.StatusCode, resp.Description)

		if len(resp.Headers) > 0 {
			fmt.Println("    Headers:")
			for key, value := range resp.Headers {
				fmt.Printf("      %s: %s\n", key, value)
			}
		}

		if resp.Body != "" {
			bodyPreview := resp.Body
			if len(bodyPreview) > 100 {
				bodyPreview = bodyPreview[:100] + "..."
			}
			fmt.Printf("    Body: %d bytes\n", len(resp.Body))
			fmt.Printf("      Preview: %s\n", bodyPreview)
		}
	}

	fmt.Println()
}
