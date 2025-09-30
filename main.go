package main

import (
	"fmt"
	"os"

	"github.com/pretodev/anansi-proxy/internal/parser"
	"github.com/pretodev/anansi-proxy/internal/server"
	"github.com/pretodev/anansi-proxy/internal/state"
	"github.com/pretodev/anansi-proxy/internal/ui"
)

func main() {
	resPath := "./resources/api-rest.httpr"
	port := 8977
	res, err := parser.Parse(resPath)
	if err != nil {
		fmt.Printf("Erro ao fazer o parse do arquivo: %v\n", err)
		os.Exit(1)
	}

	if len(res) == 0 {
		fmt.Println("Nenhuma resposta encontrada no arquivo.")
		os.Exit(0)
	}

	sm := state.New()

	httpSrv := server.New(sm, res)
	go func() {
		if err := httpSrv.Serve(port); err != nil {
			fmt.Printf("Erro no servidor HTTP: %v\n", err)
			os.Exit(1)
		}
	}()

	if err := ui.Render(sm, res); err != nil {
		fmt.Printf("Error na UI: %c\n", err)
		os.Exit(1)
	}
}
