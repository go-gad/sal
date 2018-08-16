package main

import (
	"io/ioutil"
	"testing"
)

func TestGenerateCode(t *testing.T) {
	code, err := GenerateCode("actsal", "github.com/go-gad/sal/examples/bookstore1", []string{"StoreClient"})
	if err != nil {
		t.Fatalf("Failed to generate a code: %+v", err)
	}

	t.Logf("\n%s", string(code))
	if err = ioutil.WriteFile("../examples/bookstore1/actsal/sal_client.go", code, 0666); err != nil {
		t.Fatalf("failed to write file: %+v", err)
	}
}
