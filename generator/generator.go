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

func generate(packageName, inputFile string) (string, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, inputFile, nil, parser.ParseComments)
	if err != nil {
		return "", err
	}
	writer := bytes.NewBuffer(nil)
	writeString := func(s string) {
		_, err := writer.WriteString(s)
		if err != nil {
			panic(err)
		}
	}

	decls := prependPackageDecl(packageName, file.Decls)

	// Iterate over the declarations in the input file
	for _, decl := range decls {
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
					writeString(cmt.Text + "\n")
				}
			}
		}
		writeString("\n")

		// Print the declaration to the output file
		err = printer.Fprint(writer, fileSet, decl)
		if err != nil {
			return "", err
		}
		writeString("\n")
	}
	val := writer.String()
	for k, v := range replacements {
		val = strings.ReplaceAll(val, k, v)
	}
	return val, nil
}

func prependPackageDecl(packageName string, decls []ast.Decl) []ast.Decl {
	packageDecl := &ast.GenDecl{
		Tok: token.PACKAGE,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{
					ast.NewIdent(packageName),
				},
			},
		},
	}
	return append([]ast.Decl{packageDecl}, decls...)
}

func main() {
	// get path to input file from -input flag
	var inputFile string
	flag.StringVar(&inputFile, "input", "", "path to input file")
	var packageName string
	flag.StringVar(&packageName, "package", "", "package name")
	flag.Parse()

	// generate output file
	got, err := generate(packageName, inputFile)
	if err != nil {
		panic(err)
	}
	_, err = os.Stdout.WriteString(got)
	if err != nil {
		panic(err)
	}
}
