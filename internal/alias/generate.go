package alias

import (
	"bytes"
	"embed"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"go.chrisrx.dev/x/slices"
	goimports "golang.org/x/tools/imports"

	"go.chrisrx.dev/tools/internal/ast"
)

//go:embed *.go.tmpl
var fs embed.FS
var templates = template.Must(template.New("").Funcs(Funcs).ParseFS(fs, "*.go.tmpl"))

var Funcs = map[string]any{
	"join": func(sep string, v any) string {
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Array, reflect.Slice:
			var s []string
			for i := range rv.Len() {
				fv := rv.Index(i)
				s = append(s, fmt.Sprintf("%s", fv.Interface()))
			}
			return strings.Join(s, sep)
		default:
			return fmt.Sprint(v)
		}
	},
	"fields": (&ast.Printer{}).PrintFieldList,
	"fieldnames": func(fields []*ast.Field) []string {
		return slices.FlatMap(fields, func(field *ast.Field) []string {
			names := slices.Map(field.Names, func(ident *ast.Ident) string {
				return ident.String()
			})
			if _, ok := field.Type.(*ast.Ellipsis); ok {
				return slices.Map(names, func(name string) string {
					return name + "..."
				})
			}
			for i, name := range names {
				if name == "_" {
					// need zero value
					if ident, ok := field.Type.(*ast.Ident); ok {
						if ident.Name == "string" {
							names[i] = `""`
						}
					}
				}
			}
			return names
		})
	},
}

func GenerateFile(file *File) ([]byte, error) {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "alias.go.tmpl", file); err != nil {
		return nil, err
	}
	return goimports.Process("", buf.Bytes(), &goimports.Options{
		TabIndent: true,
		AllErrors: true,
		Fragment:  true,
		Comments:  true,
	})
}
