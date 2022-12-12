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
	"health.HealthStatusCode":                 "string",
	"synccommon.ResultCode":                   "string",
	"synccommon.HookType":                     "string",
	"synccommon.OperationPhase":               "string",
	"synccommon.SyncPhase":                    "string",
	"github.com/argoproj/argo-cd/v2/pkg/apis": "github.com/willabides/argocd-types/argocd-apis",
}

var allowedFuncs = []string{
	"MarshalJSON",
	"UnmarshalJSON",
	"DeepCopy",
	"DeepCopyInto",
	"DeepCopyObject",
	"addKnownTypes",
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

		// Filter var declarations
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
			var specs []ast.Spec
			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				// must be exported
				if !valueSpec.Names[0].IsExported() {
					continue
				}
				// must not have a value from the env package like:
				// K8sClientConfigQPS float32 = env.ParseFloatFromEnv(EnvK8sClientQPS, 50, 0, math.MaxFloat32)
				if len(valueSpec.Values) > 0 {
					if callExpr, ok := valueSpec.Values[0].(*ast.CallExpr); ok {
						if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
							if selExpr.X.(*ast.Ident).Name == "env" {
								continue
							}
						}
					}
				}
				specs = append(specs, valueSpec)
			}
			if len(specs) == 0 {
				continue
			}
			genDecl.Specs = specs
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
