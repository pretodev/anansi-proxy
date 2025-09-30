package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pretodev/anansi-proxy/internal/parser"
	"github.com/pretodev/anansi-proxy/internal/server"
	"github.com/pretodev/anansi-proxy/internal/state"
	"github.com/pretodev/anansi-proxy/internal/ui"
)

func main() {
	var resPath string
	var port int

	flag.StringVar(&resPath, "file", "", "Path to the HTTP response file to parse (required)")
	flag.StringVar(&resPath, "f", "", "Path to the HTTP response file to parse (required, shorthand)")
	flag.IntVar(&port, "port", 8977, "Port number for the HTTP server")
	flag.IntVar(&port, "p", 8977, "Port number for the HTTP server (shorthand)")
	flag.Parse()

	if resPath == "" {
		fmt.Println("Error: file path is required. Use -file or -f to specify the HTTP response file.")
		flag.Usage()
		os.Exit(1)
	}
	res, err := parser.Parse(resPath)
	if err != nil {
		fmt.Printf("Error parsing file: %v\n", err)
		os.Exit(1)
	}

	if len(res) == 0 {
		fmt.Println("No responses found in the file.")
		os.Exit(0)
	}

	sm := state.New()

	httpSrv := server.New(sm, res)
	go func() {
		if err := httpSrv.Serve(port); err != nil {
			fmt.Printf("HTTP server error: %v\n", err)
			os.Exit(1)
		}
	}()

	if err := ui.Render(sm, res); err != nil {
		fmt.Printf("UI error: %v\n", err)
		os.Exit(1)
	}
}
