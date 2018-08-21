package looker

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/pkg/errors"
)

var (
	buildFlags = flag.String("build_flags", "", "Additional flags for go build.")
)

func Reflect(importPath string, symbols []string) (*Package, error) {
	program, err := writeProgram(importPath, symbols)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("PROGRAMM \n%s\n----------\n", string(program))

	//wd, _ := os.Getwd()

	//// Try to run the program in the same directory as the input package.
	//if p, err := build.Import(importPath, wd, build.FindOnly); err == nil {
	//	dir := p.Dir
	//	if p, err := buildAndRun(program, dir); err == nil {
	//		return p, nil
	//	}
	//}
	//
	//// Since that didn't work, try to run it in the current working directory.
	//if p, err := buildAndRun(program, wd); err == nil {
	//	return p, nil
	//}

	// Since that didn't work, try to run it in a standard temp directory.
	return buildAndRun(program, "")
}

// run the given program and parse the output as a model.Package.
func run(program string) (*Package, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temp file")
	}
	filename := f.Name()
	defer os.Remove(filename)
	if err := f.Close(); err != nil {
		return nil, err
	}

	// Run the program.
	cmd := exec.Command(program, "-output", filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	f, err = os.Open(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file %s", filename)
	}

	// Process output.
	var pkg Package
	gob.Register(&StructElement{})
	gob.Register(&SliceElement{})
	gob.Register(&InterfaceElement{})

	if err := gob.NewDecoder(f).Decode(&pkg); err != nil {
		return nil, errors.Wrap(err, "failed to decode pkg")
	}

	if err := f.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close file")
	}

	return &pkg, nil
}

func buildAndRun(program []byte, dir string) (*Package, error) {
	// We use TempDir instead of TempFile so we can control the filename.
	tmpDir, err := ioutil.TempDir(dir, "sal_reflect_")
	if err != nil {
		return nil, fmt.Errorf("failed to create tmp dir: %s", err)
	}
	defer func() { os.RemoveAll(tmpDir) }()
	const progSource = "prog.go"
	var progBinary = "prog.bin"
	if runtime.GOOS == "windows" {
		// Windows won't execute a program unless it has a ".exe" suffix.
		progBinary += ".exe"
	}

	if err := ioutil.WriteFile(filepath.Join(tmpDir, progSource), program, 0600); err != nil {
		return nil, err
	}

	cmdArgs := []string{}
	cmdArgs = append(cmdArgs, "build")
	if *buildFlags != "" {
		cmdArgs = append(cmdArgs, *buildFlags)
	}
	cmdArgs = append(cmdArgs, "-o", progBinary, progSource)

	// Build the program.
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run cmd %v: %s", cmdArgs, err)
	}
	return run(filepath.Join(tmpDir, progBinary))
}

func writeProgram(importPath string, symbols []string) ([]byte, error) {
	var program bytes.Buffer
	data := reflectData{
		ImportPath: importPath,
		Symbols:    symbols,
	}
	if err := reflectProgram.Execute(&program, &data); err != nil {
		return nil, err
	}
	return program.Bytes(), nil
}

type reflectData struct {
	ImportPath string
	Symbols    []string
}

func EncodeGob(output string, pkg *Package) error {
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

	gob.Register(&StructElement{})
	gob.Register(&SliceElement{})
	gob.Register(&InterfaceElement{})
	//gob.Register(Parameters{})
	//gob.Register(Field{})
	//gob.Register(Fields{})

	if err := gob.NewEncoder(outfile).Encode(pkg); err != nil {
		fmt.Errorf("gob encode: %s", err)
	}

	return nil
}

// This program reflects on an interface value, and prints the
// gob encoding of a model.Package to standard output.
// JSON doesn't work because of the model.Type interface.
var reflectProgram = template.Must(template.New("program").Parse(`
package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	"github.com/go-gad/sal/looker"

	pkg_ {{printf "%q" .ImportPath}}
)

var output = flag.String("output", "", "The output file name, or empty to use stdout.")

func main() {
	flag.Parse()
	
	pkgPath := {{printf "%q" .ImportPath}}
	var list = []reflect.Type{
		{{range .Symbols}}
		reflect.TypeOf((*pkg_.{{.}})(nil)).Elem(),
		{{end}}
	}

	pkg := looker.LookAtInterfaces(pkgPath, list)

	if err := looker.EncodeGob(*output, pkg); err != nil {
		fmt.Fprintf(os.Stderr, "failed EncodeGob: %s\n", err)
		os.Exit(1)
	}
}
`))
