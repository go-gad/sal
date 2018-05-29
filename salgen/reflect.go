package main

import (
	"encoding/gob"
	"fmt"
	"os"

	"github.com/go-gad/sal/looker"
)

func EncodeGob(output string, pkg *looker.Package) error {
	outfile := os.Stdout

	if len(output) != 0 {
		var err error
		if outfile, err = os.Create(output); err != nil {
			return fmt.Errorf("failed to open output file %q: %s", output, err)
		}
		defer func() {
			if err := outfile.Close(); err != nil {
				fmt.Errorf("failed to close output file %q: %s", output, err)
			}
		}()
	}

	if err := gob.NewEncoder(outfile).Encode(pkg); err != nil {
		fmt.Errorf("gob encode: %s", err)
	}

	return nil
}
