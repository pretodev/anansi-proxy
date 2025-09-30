package main

import (
	"fmt"

	"github.com/pretodev/anansi-proxy/internal/parser"
)

func main() {
	resPath := "./resources/api-rest.httpr"

	res := parser.Parse(resPath)
	fmt.Println(res)
}
