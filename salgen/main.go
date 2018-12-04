package main

import (
	"flag"
	"io"
	"log"
	"os"
	"strings"

	"github.com/go-gad/sal/looker"
	"github.com/pkg/errors"
)

var (
	destination = flag.String("destination", "", "Output file; defaults to stdout.")
	packageName = flag.String("package", "", "The full import path of the library for the generated implementation")
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		log.Fatal("Expected exactly two arguments")
	}
	var (
		srcpkg  = flag.Arg(0)
		symbols = strings.Split(flag.Arg(1), ",")
	)

	if len(*destination) > 0 {
		err := os.Remove(*destination)
		if err != nil && !os.IsNotExist(err) {
			log.Fatalf("Failed to remove destination file: %+v", err)
		}
	}

	dstPkg := looker.ImportElement{Path: *packageName}
	code, err := GenerateCode(dstPkg, srcpkg, symbols)
	if err != nil {
		log.Fatalf("Failed to generate a code: %+v", err)
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

func usage() {
	io.WriteString(os.Stderr, usageText)
	flag.PrintDefaults()
}

const usageText = `Usage:
    salgen [options...] <import_path> <interface_name>

Example:
    salgen -destination=./client.go -package=github.com/go-gad/sal/examples/profile/storage github.com/go-gad/sal/examples/profile/storage Store

  <import_path> 
        describes the complete package path where the interface is located.
  <interface_name> 
        indicates the interface name itself.

Options:
`
