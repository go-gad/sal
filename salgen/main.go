package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/go-gad/sal/looker"
	"github.com/pkg/errors"
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
	var (
		srcpkg  = flag.Arg(0)
		symbols = strings.Split(flag.Arg(1), ",")
	)

	dst := os.Stdout
	if len(*destination) > 0 {
		f, err := os.Create(*destination)
		if err != nil {
			log.Fatalf("Failed opening destination file: %v", err)
		}
		defer f.Close()
		dst = f
	}
	dstPkg := looker.ImportElement{Path: *packageName}
	code, err := GenerateCode(dstPkg, srcpkg, symbols)
	if err != nil {
		log.Fatalf("Failed to generate a code: %+v", err)
	}

	if _, err := dst.Write(code); err != nil {
		log.Fatalf("Failed writing to destination: %v", err)
	}

}

func GenerateCode(dstPkg looker.ImportElement, srcpkg string, symbols []string) ([]byte, error) {
	pkg, err := looker.Reflect(srcpkg, symbols)
	if err != nil {
		return nil, errors.Wrap(err, "failed to reflect package")
	}

	g := new(generator)

	if err := g.Generate(pkg, dstPkg); err != nil {
		return nil, errors.Wrap(err, "failed generating mock")
	}

	return g.Output(), nil
}
