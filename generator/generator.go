package main

import (
	"bytes"
	"flag"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

var replacements = map[string]string{
	"health.HealthStatusCode":   "string",
	"synccommon.ResultCode":     "string",
	"synccommon.HookType":       "string",
	"synccommon.OperationPhase": "string",
	"synccommon.SyncPhase":      "string",
}

var allowedFuncs = []string{
	"MarshalJSON",
	"UnmarshalJSON",
	"DeepCopy",
	"DeepCopyInto",
}

func stringSliceContains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func generate(inputFile string) (string, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, inputFile, nil, parser.ParseComments)
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer

	// Iterate over the declarations in the input file
	for _, decl := range file.Decls {
		// Skip function declarations except for allowedFuncs
		if fnDecl, ok := decl.(*ast.FuncDecl); ok {
			if !stringSliceContains(allowedFuncs, fnDecl.Name.Name) {
				continue
			}
		}

		// Skip var declarations
		if _, ok := decl.(*ast.GenDecl); ok {
			if decl.(*ast.GenDecl).Tok == token.VAR {
				continue
			}
		}

		// Print any comments associated with the declaration
		for _, c := range file.Comments {
			if c.Pos() == decl.Pos() {
				// Print each comment in the CommentGroup individually
				for _, cmt := range c.List {
					buffer.WriteString(cmt.Text + "\n")
				}
			}
		}
		buffer.WriteString("\n")

		// Print the declaration to the output file
		err = printer.Fprint(&buffer, fileSet, decl)
		if err != nil {
			return "", err
		}
		buffer.WriteString("\n")
	}
	val := buffer.String()
	for k, v := range replacements {
		val = strings.ReplaceAll(val, k, v)
	}
	return val, nil
}

func main() {
	// get path to input file from -input flag
	var inputFile string
	flag.StringVar(&inputFile, "input", "", "path to input file")
	flag.Parse()

	// generate output file
	got, err := generate(inputFile)
	if err != nil {
		panic(err)
	}
	_, err = os.Stdout.WriteString(got)
	if err != nil {
		panic(err)
	}
}
