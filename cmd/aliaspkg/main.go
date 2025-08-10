package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"

	"go.chrisrx.dev/tools/internal/alias"
)

func main() {
	var (
		pkg     string
		ignore  string
		include string
		out     string
		stdout  bool
	)
	flag.StringVar(&pkg, "package", "", "The package path to alias")
	flag.StringVar(&ignore, "ignore", "", "A comma-delimited list of identifiers to ignore")
	flag.StringVar(&include, "include", "", "A comma-delimited list of identifiers to include (overrides ignore)")
	flag.StringVar(&out, "out", "", "Directory to output alias.go file")
	flag.BoolVar(&stdout, "stdout", false, "Output to stdout")
	flag.Parse()

	if pkg == "" {
		v, ok := os.LookupEnv("GOPACKAGE")
		if !ok {
			log.Fatal("must provide package")
		}
		pkg = v
	}

	if out == "" {
		_, ok := os.LookupEnv("GOFILE")
		if !ok && !stdout {
			log.Fatal("must provide output directory or stdout")
		}
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		out = filepath.Join(dir, "alias.go")
	}

	pkgs, err := packages.Load(&packages.Config{Mode: packages.LoadAllSyntax}, pkg)
	if err != nil {
		log.Fatal(err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	if len(pkgs) == 0 {
		log.Fatal("no packages")
	}

	data, err := alias.GenerateFile(
		alias.Parse(pkgs[0],
			alias.Ignore(strings.Split(ignore, ",")...),
			alias.Include(strings.Split(include, ",")...),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	if stdout {
		fmt.Printf("%s\n", data)
		return
	}
	if err := os.WriteFile(out, data, 0644); err != nil {
		log.Fatal(err)
	}
}
