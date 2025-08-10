package ast

import (
	"fmt"
	"go/ast"
	"strings"

	"go.chrisrx.dev/x/set"
	"go.chrisrx.dev/x/slices"
)

//go:generate go tool aliaspkg -package go/ast

type Printer struct {
	Imports set.Set[string]
}

func (p *Printer) Visit(node ast.Node) {
	switch t := node.(type) {
	case ast.Decl:
		p.VisitDecl(t)
	case ast.Expr:
		p.VisitExpr(t)
	}
}

func (p *Printer) VisitDecl(decl ast.Decl) {
	switch t := decl.(type) {
	case *ast.GenDecl:
		for _, spec := range t.Specs {
			switch t := spec.(type) {
			case *ast.TypeSpec:
				p.PrintFieldList(t.TypeParams)
			}
		}
	case *ast.FuncDecl:
		p.PrintFieldList(t.Type.TypeParams)
		p.PrintFieldList(t.Type.Params)
		p.PrintFieldList(t.Type.Results)
	}
}

func (p *Printer) VisitExpr(expr ast.Expr) {
	p.PrintExpr(expr)
}

func (p *Printer) Print(node ast.Node) string {
	switch t := node.(type) {
	case ast.Decl:
		return ""
	case ast.Expr:
		return p.PrintExpr(t)
	default:
		return ""
	}
}

func (p *Printer) PrintExpr(x ast.Expr) string {
	switch t := x.(type) {
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", p.PrintExpr(t.X))
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		p.Imports.Add(p.PrintExpr(t.X))
		return fmt.Sprintf("%s.%s", p.PrintExpr(t.X), t.Sel.Name)
	case *ast.ArrayType:
		return fmt.Sprintf("[]%s", p.PrintExpr(t.Elt))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", p.PrintExpr(t.Key), p.PrintExpr(t.Value))
	case *ast.Ellipsis:
		return fmt.Sprintf("...%s", p.PrintExpr(t.Elt))
	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", p.PrintExpr(t.X), t.Index)
	case *ast.IndexListExpr:
		indices := strings.Join(slices.Map(t.Indices, p.PrintExpr), ", ")
		return fmt.Sprintf("%s[%s]", p.PrintExpr(t.X), indices)
	case *ast.FuncType:
		var sb strings.Builder
		sb.WriteString("func(")
		sb.WriteString(strings.Join(p.PrintFieldList(t.Params), ", "))
		sb.WriteString(") ")
		sb.WriteString("(")
		sb.WriteString(strings.Join(p.PrintFieldList(t.Results), ", "))
		sb.WriteString(")")
		return sb.String()
	case *ast.UnaryExpr:
		return p.PrintExpr(t.X)
	case *ast.StructType:
		p.PrintFieldList(t.Fields)
		return "idk"
	default:
		panic(fmt.Errorf("tt:  %T", t))
	}
}

func (p *Printer) PrintFieldList(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	return slices.Map(fields.List, func(field *ast.Field) string {
		if len(field.Names) == 0 {
			return p.PrintExpr(field.Type)
		}
		return fmt.Sprintf("%s %s", strings.Join(
			slices.Map(field.Names, func(ident *ast.Ident) string {
				return ident.String()
			}), ", ",
		), p.PrintExpr(field.Type))
	})
}
