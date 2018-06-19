package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/go-gad/sal/looker"
)

var (
	destination = flag.String("destination", "", "Output file; defaults to stdout.")
	packageName = flag.String("package", "", "Package of the generated code.")
)

func main() {
	flag.Parse()

	if flag.NArg() != 2 {
		log.Fatal("Expected exactly two arguments")
	}
	pkg, err := looker.Reflect(flag.Arg(0), strings.Split(flag.Arg(1), ","))

	if err != nil {
		log.Fatalf("Loading input failed: %v", err)
	}

	dst := os.Stdout
	if len(*destination) > 0 {
		f, err := os.Create(*destination)
		if err != nil {
			log.Fatalf("Failed opening destination file: %v", err)
		}
		defer f.Close()
		dst = f
	}

	g := new(generator)

	if err := g.Generate(pkg, *packageName); err != nil {
		log.Fatalf("Failed generating mock: %v", err)
	}
	if _, err := dst.Write(g.Output()); err != nil {
		log.Fatalf("Failed writing to destination: %v", err)
	}

}
