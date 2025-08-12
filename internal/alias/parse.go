package alias

import (
	"strings"

	"go.chrisrx.dev/x/ptr"
	"go.chrisrx.dev/x/set"
	"go.chrisrx.dev/x/slices"
	"golang.org/x/tools/go/packages"

	"go.chrisrx.dev/tools/internal/ast"
)

func Parse(pkg *packages.Package, opts ...Option) *File {
	o := newOptions(opts)
	if len(pkg.Syntax) == 0 {
		return nil
	}

	imports := set.New(pkg.PkgPath)
	for _, pkg := range pkg.Imports {
		if strings.Contains(pkg.PkgPath, "internal/") {
			continue
		}
		imports.Add(pkg.PkgPath)
	}
	file := &File{
		Name: pkg.Syntax[0].Name.String(),
		Docs: o.IncludeDocs(slices.FlatMap(pkg.Syntax, func(f *ast.File) []string {
			return slices.Map(ptr.From(f.Doc).List, func(c *ast.Comment) string {
				return c.Text
			})
		}), "package", "all"),
		PkgPath: pkg.PkgPath,
		Imports: imports.List(),
	}

	for _, f := range pkg.Syntax {
		for _, d := range f.Decls {
			switch d := d.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch t := spec.(type) {
					case *ast.TypeSpec:
						if !t.Name.IsExported() {
							continue
						}
						if !o.ShouldInclude(t.Name.String()) {
							continue
						}
						file.Aliases = append(file.Aliases, &Alias{
							Name: t.Name.String(),
							Docs: o.IncludeDocs(slices.Map(ptr.From(d.Doc).List, func(c *ast.Comment) string {
								return c.Text
							}), "decls", "all"),
						})
					case *ast.ValueSpec:
						for _, name := range t.Names {
							if name.IsExported() {
								if !o.ShouldInclude(name.String()) {
									continue
								}
								switch name.Obj.Kind {
								case ast.Con:
									file.Consts = append(file.Consts, &Const{
										Name: name.String(),
										Docs: o.IncludeDocs(slices.Map(ptr.From(t.Doc).List, func(c *ast.Comment) string {
											return c.Text
										}), "decls", "all"),
									})
								case ast.Var:
									file.Vars = append(file.Vars, &Var{
										Name: name.String(),
										Docs: o.IncludeDocs(slices.Map(ptr.From(t.Doc).List, func(c *ast.Comment) string {
											return c.Text
										}), "decls", "all"),
									})
								}
							}
						}
					}
				}
			case *ast.FuncDecl:
				if !d.Name.IsExported() {
					continue
				}
				if d.Recv != nil {
					continue
				}
				if !o.ShouldInclude(d.Name.String()) {
					continue
				}

				// Add a trailor to the field name if the parameter overlaps with an
				// import. Currently just looks at the package being aliased.
				for _, field := range ptr.From(d.Type.Params).List {
					for i, name := range field.Names {
						if name.String() == file.Name {
							field.Names[i].Name = name.String() + "_"
						}
					}
				}

				file.Funcs = append(file.Funcs, &Func{
					Name: d.Name.String(),
					Docs: o.IncludeDocs(slices.Map(ptr.From(d.Doc).List, func(c *ast.Comment) string {
						return c.Text
					}), "decls", "all"),
					TypeParams: d.Type.TypeParams,
					Params:     d.Type.Params,
					Results:    d.Type.Results,
				})
			}
		}
	}

	return file
}
