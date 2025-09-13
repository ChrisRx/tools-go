package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"golang.org/x/tools/go/packages"

	"go.chrisrx.dev/tools/internal/alias"
)

var opts struct {
	Docs              string
	Package           string
	Ignore            string
	Include           string
	File              string
	Stdout            bool
	UseGoBuildVersion bool
}

func init() {
	flag.StringVar(&opts.Docs, "docs", "none", "Documentation to generate")
	flag.StringVar(&opts.Ignore, "ignore", "", "A comma-delimited list of identifiers to ignore")
	flag.StringVar(&opts.Include, "include", "", "A comma-delimited list of identifiers to include (overrides ignore)")
	flag.StringVar(&opts.File, "file", "", "Output location for generated file")
	flag.BoolVar(&opts.Stdout, "stdout", false, "Output to stdout")
	flag.BoolVar(&opts.UseGoBuildVersion, "gover", false, "Generate file specific to Go build version")
	flag.Parse()
	if flag.NArg() > 0 {
		opts.Package = flag.Args()[0]
	}

	if opts.Package == "" {
		pkg, ok := os.LookupEnv("GOPACKAGE")
		if !ok {
			fmt.Fprint(os.Stderr, Usage)
			os.Exit(1)
		}
		opts.Package = pkg
	}
	if opts.File == "" {
		_, ok := os.LookupEnv("GOFILE")
		if !ok && !opts.Stdout {
			log.Fatal("must provide output directory or stdout")
		}
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		filename := "alias.go"
		if opts.UseGoBuildVersion {
			filename = fmt.Sprintf("alias_%s.go", strings.ReplaceAll(goBuildVersion(), ".", "_"))
		}
		opts.File = filepath.Join(dir, filename)
	}
}

func goBuildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		panic("must build with module support to use Go build version")
	}
	switch strings.Count(info.GoVersion, ".") {
	case 0, 1:
		return info.GoVersion
	default:
		return strings.Join(strings.Split(info.GoVersion, ".")[:2], ".")
	}

}

const Usage = `aliaspkg: generate aliases for Go packages

Usage:

  aliaspkg  [-docs=package|decls|all|none] [-file=...] <package>

Flags:

-docs               Specifies the documentation that should be included, one of:

                      package      Package-level docs
                      decls        Docs for functions/types/variables
                      all          All available docs
                      none         Don't include docs

                    If not provided, defaults to none.

-file               Specifies the output file for the generated package alias.

                    If GOFILE is set in the environment then this defaults to the
                    $(pwd)/alias.go.

                    Ignored if -stdout is provided.

-include/-ignore    Specifies types, functions and variables to include/ignore
                    from the package being aliased.

-stdout             Include the package's tests in the analysis.
-gover              Generate a file only for this Go build version
`

func main() {
	pkgs, err := packages.Load(&packages.Config{Mode: packages.LoadAllSyntax}, opts.Package)
	if err != nil {
		log.Fatal(err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	if len(pkgs) == 0 {
		log.Fatal("no packages")
	}

	parseOpts := []alias.Option{
		alias.Ignore(strings.Split(opts.Ignore, ",")...),
		alias.Include(strings.Split(opts.Include, ",")...),
		alias.Docs(opts.Docs),
	}
	if opts.UseGoBuildVersion {
		parseOpts = append(parseOpts, alias.GoBuildVersion(goBuildVersion()))
	}
	data, err := alias.GenerateFile(alias.Parse(pkgs[0], parseOpts...))
	if err != nil {
		log.Fatal(err)
	}

	if opts.Stdout {
		fmt.Printf("%s\n", data)
		return
	}
	if err := os.WriteFile(opts.File, data, 0644); err != nil {
		log.Fatal(err)
	}
}
