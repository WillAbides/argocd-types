package main

import (
	"fmt"
	"testing"
)

func Test_generate(t *testing.T) {
	inputFile := "/Users/wroden/repos/argoproj/argo-cd/pkg/apis/application/v1alpha1/types.go"
	got, err := generate("mypkg", inputFile)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(got)
}
