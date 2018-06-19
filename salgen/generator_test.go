package main

import (
	"log"
	"testing"

	"github.com/go-gad/sal/looker"
)

func TestGenerator_Generate(t *testing.T) {
	pkg, err := looker.Reflect("github.com/go-gad/sal/internal/bookstore", []string{"StoreClient"})
	if err != nil {
		t.Fatal(err)
	}
	g := new(generator)
	if err := g.Generate(pkg, "actsal"); err != nil {
		log.Fatalf("Failed generating mock: %v", err)
	}
	var fbody = g.Output()
	t.Logf("\n%s", string(fbody))

}
