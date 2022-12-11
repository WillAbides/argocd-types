package main

import (
	"bytes"
	"flag"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"strings"
)

var replacements = map[string]string{
	"health.HealthStatusCode":   "string",
	"synccommon.ResultCode":     "string",
	"synccommon.HookType":       "string",
	"synccommon.OperationPhase": "string",
	"synccommon.SyncPhase":      "string",
	"v1.ResourceName":           "string",
	"watch.EventType":           "string",
}

func generate(packageName, inputFile string) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, inputFile, nil, parser.ParseComments)
	if err != nil {
		return "", err
	}
	var writer bytes.Buffer

	// Iterate over the declarations in the input file
	for _, decl := range prependPackageDecl(packageName, f.Decls) {
		// Skip function declarations except for MarshalJSON and UnmarshalJSON
		if fnDecl, ok := decl.(*ast.FuncDecl); ok {
			_ = fnDecl
			if fnDecl.Name.Name != "MarshalJSON" && fnDecl.Name.Name != "UnmarshalJSON" {
				continue
			}
			continue
		}

		// Skip var declarations
		if _, ok := decl.(*ast.GenDecl); ok {
			if decl.(*ast.GenDecl).Tok == token.VAR {
				continue
			}
		}

		// Print any comments associated with the declaration
		for _, c := range f.Comments {
			if c.Pos() == decl.Pos() {
				// Print each comment in the CommentGroup individually
				for _, cmt := range c.List {
					_, err = writer.WriteString(cmt.Text)
					if err != nil {
						return "", err
					}
					mustWriteNewline(&writer)
				}
			}
		}

		mustWriteNewline(&writer)

		// Print the declaration to the output file
		err = printer.Fprint(&writer, fset, decl)
		if err != nil {
			return "", err
		}
		mustWriteNewline(&writer)
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

func mustWriteNewline(writer io.Writer) {
	_, err := writer.Write([]byte{'\n'})
	if err != nil {
		panic(err)
	}
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
