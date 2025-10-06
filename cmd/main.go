package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pretodev/anansi-proxy/internal/discovery"
	"github.com/pretodev/anansi-proxy/internal/endpoint"
	"github.com/pretodev/anansi-proxy/internal/server"
	"github.com/pretodev/anansi-proxy/internal/state"
	"github.com/pretodev/anansi-proxy/internal/ui"
)

func main() {
	var port int
	var interactive bool

	flag.IntVar(&port, "port", 8977, "Port number for the HTTP server")
	flag.IntVar(&port, "p", 8977, "Port number for the HTTP server (shorthand)")
	flag.BoolVar(&interactive, "it", false, "Interactive mode - display response selection UI")
	flag.Parse()

	// Get paths from positional arguments
	paths := flag.Args()
	if len(paths) == 0 {
		fmt.Println("Error: at least one file or directory path is required.")
		fmt.Println("\nUsage:")
		fmt.Println("  anansi-proxy [options] <file_or_directory>...")
		fmt.Println("\nExamples:")
		fmt.Println("  anansi-proxy ./docs/example/simple.apimock")
		fmt.Println("  anansi-proxy ./docs/example/simple.apimock ./docs/example/xml.apimock")
		fmt.Println("  anansi-proxy ./docs/example")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	filePaths, err := discovery.FindAPIMockFiles(paths...)
	if err != nil {
		fmt.Printf("Error finding .apimock files: %v\n", err)
		os.Exit(1)
	}

	if len(filePaths) > 0 && interactive {
		fmt.Println("Warning: Interactive mode is not supported when multiple files are provided. Defaulting to non-interactive mode.")
		interactive = false
	}

	fmt.Printf("Found %d .apimock file(s)\n", len(filePaths))

	endpoints, err := endpoint.ParseAPIMockFiles(filePaths...)
	if err != nil {
		fmt.Printf("Error parsing files: %v\n", err)
		os.Exit(1)
	}

	if len(endpoints) == 0 {
		fmt.Println("Error: no valid endpoints found")
		os.Exit(1)
	}

	if len(endpoints) == 1 && interactive {
		runInteractiveMode(endpoints[0].Schema, port)
		return
	}

	httpSrv := server.New(endpoints)
	go func() {
		if err := httpSrv.Serve(port); err != nil {
			fmt.Printf("HTTP server error: %v\n", err)
			os.Exit(1)
		}
	}()

	fmt.Printf("\nServer ready! Serving %d endpoint(s):\n", len(endpoints))

	for i, ep := range endpoints {
		responses := ep.Schema.SliceResponses()
		if len(responses) > 0 {
			firstResponse := responses[0]
			fmt.Printf("  [%d] %s -> [%d] %s\n", i, ep.Schema.Route, firstResponse.StatusCode, firstResponse.Title)
		} else {
			fmt.Printf("  [%d] %s -> (no responses)\n", i, ep.Schema.Route)
		}
	}
	select {}
}

func runInteractiveMode(endpoint *endpoint.EndpointSchema, port int) {
	sm := state.New(endpoint.CountResponses())

	httpSrv := server.NewInteractive(sm, endpoint)
	go func() {
		if err := httpSrv.Serve(port); err != nil {
			fmt.Printf("HTTP server error: %v\n", err)
			os.Exit(1)
		}
	}()

	if err := ui.Render(sm, endpoint); err != nil {
		fmt.Printf("UI error: %v\n", err)
		os.Exit(1)
	}
}
