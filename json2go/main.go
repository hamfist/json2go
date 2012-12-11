package main

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
)

import (
	"github.com/modcloth/json2go"
)

func main() {
	var file *ast.File
	var err error

	if file, err = json2go.Json2Ast(os.Stdin); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ast.Print(nil, file)

	printer.Fprint(os.Stdout, token.NewFileSet(), file)
}
