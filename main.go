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
	var interactive bool

	flag.StringVar(&resPath, "file", "", "Path to the HTTP response file to parse (required)")
	flag.StringVar(&resPath, "f", "", "Path to the HTTP response file to parse (required, shorthand)")
	flag.IntVar(&port, "port", 8977, "Port number for the HTTP server")
	flag.IntVar(&port, "p", 8977, "Port number for the HTTP server (shorthand)")
	flag.BoolVar(&interactive, "it", false, "Interactive mode - display response selection UI")
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

	if !interactive {
		sm.SetIndex(0)
	}

	httpSrv := server.New(sm, res)
	go func() {
		if err := httpSrv.Serve(port); err != nil {
			fmt.Printf("HTTP server error: %v\n", err)
			os.Exit(1)
		}
	}()

	if interactive {
		if err := ui.Render(sm, res); err != nil {
			fmt.Printf("UI error: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Server running on port %d using response: [%d] %s\n", port, res[0].StatusCode, res[0].Title)
		select {}
	}
}
