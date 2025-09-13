package alias

import (
	"go.chrisrx.dev/tools/internal/ast"
)

type File struct {
	Name    string
	PkgPath string
	Docs    []string
	Imports []string
	Aliases []*Alias
	Funcs   []*Func
	Consts  []*Const
	Vars    []*Var

	GoBuildVersion string
}

type Const struct {
	Name string
	Docs []string
}

type Var struct {
	Name string
	Docs []string
}

type Func struct {
	Name       string
	Docs       []string
	TypeParams *ast.FieldList
	Params     *ast.FieldList
	Results    *ast.FieldList
}

type Alias struct {
	Name       string
	Docs       []string
	TypeParams *ast.FieldList
}
