package walker

import (
	"errors"
	"reflect"

	"github.com/Nykenik24/oxy/internal/frontend/parser"
)

type VisitFunc func(node parser.Node) error

type Walker struct {
	root     parser.Node
	visitors map[string]VisitFunc
}

func New(root parser.Node) *Walker {
	return &Walker{
		root:     root,
		visitors: make(map[string]VisitFunc),
	}
}

func (w *Walker) Before(nodeType string, fn VisitFunc) *Walker {
	w.visitors["before:"+nodeType] = fn
	return w
}

func (w *Walker) After(nodeType string, fn VisitFunc) *Walker {
	w.visitors["after:"+nodeType] = fn
	return w
}

func (w *Walker) On(nodeType string, fn VisitFunc) *Walker {
	w.Before(nodeType, fn)
	return w
}

var SkipChildren = errors.New("skip children")

func (w *Walker) dispatch(node parser.Node) error {
	if node == nil {
		return nil
	}
	key := nodeTypeName(node)

	if fn, ok := w.visitors["before:"+key]; ok {
		if err := fn(node); err != nil {
			if err == SkipChildren {
				return nil
			}
			return err
		}
	}

	if err := w.walkChildren(node); err != nil {
		return err
	}

	if fn, ok := w.visitors["after:"+key]; ok {
		if err := fn(node); err != nil {
			return err
		}
	}

	return nil
}

func (w *Walker) Walk() error {
	return w.dispatch(w.root)
}

func nodeTypeName(n parser.Node) string {
	t := reflect.TypeOf(n)
	if t.Kind() == reflect.Pointer {
		return t.Elem().Name()
	}
	return t.Name()
}

func (w *Walker) walkChildren(node parser.Node) error {
	switch n := node.(type) {

	case *parser.Program:
		return w.walkNodes(n.Items)

	case *parser.Module:
		return nil

	case *parser.FunctionDecl:
		if err := w.walkNodes(n.TypeArgs); err != nil {
			return err
		}
		if err := w.walkDeclArgs(n.Args); err != nil {
			return err
		}
		if n.Recv != nil {
			if err := w.dispatch(n.Recv.Type); err != nil {
				return err
			}
		}
		if err := w.walkNodes(n.Returns); err != nil {
			return err
		}
		return w.walkNodes(n.Body)

	case *parser.ClosureExpr:
		if err := w.walkDeclArgs(n.Params); err != nil {
			return err
		}
		if err := w.walkNodes(n.Returns); err != nil {
			return err
		}
		return w.walkNodes(n.Body)

	case *parser.CallExpr:
		if err := w.dispatch(n.Callee); err != nil {
			return err
		}
		return w.walkArgs(n.Args)

	case *parser.MethodCall:
		if err := w.dispatch(n.Object); err != nil {
			return err
		}
		return w.walkArgs(n.Args)

	case *parser.NewExpr:
		if err := w.walkNodes(n.TypeArgs); err != nil {
			return err
		}
		return w.walkArgs(n.Args)

	case *parser.BinaryExpr:
		if err := w.dispatch(n.Left); err != nil {
			return err
		}
		return w.dispatch(n.Right)

	case *parser.UnaryExpr:
		return w.dispatch(n.Operand)

	case *parser.TernaryExpr:
		if err := w.dispatch(n.Cond); err != nil {
			return err
		}
		if err := w.dispatch(n.Then); err != nil {
			return err
		}
		return w.dispatch(n.Else)

	case *parser.CoalesceExpr:
		if err := w.dispatch(n.Left); err != nil {
			return err
		}
		if err := w.dispatch(n.Right); err != nil {
			return err
		}
		return w.walkNodes(n.Block)

	case *parser.GroupExpr:
		return w.dispatch(n.Inner)

	case *parser.FieldAccess:
		return w.dispatch(n.Object)

	case *parser.IndexExpr:
		if err := w.dispatch(n.Object); err != nil {
			return err
		}
		return w.dispatch(n.Index)

	case *parser.ForceUnwrap:
		return w.dispatch(n.Operand)

	case *parser.OptionalChain:
		return w.dispatch(n.Operand)

	case *parser.CatchExpr:
		if err := w.dispatch(n.Operand); err != nil {
			return err
		}
		return w.walkNodes(n.Body)

	case *parser.VarDecl:
		return w.walkNodes(n.Exprs)

	case *parser.Return:
		return w.walkNodes(n.Exprs)

	case *parser.Raise:
		return w.dispatch(n.Expr)

	case *parser.ExprStmt:
		return w.dispatch(n.Expr)

	case *parser.IfStmt:
		if err := w.dispatch(n.Cond); err != nil {
			return err
		}
		if err := w.walkNodes(n.Body); err != nil {
			return err
		}
		for _, elif := range n.Elifs {
			if err := w.dispatch(elif.Cond); err != nil {
				return err
			}
			if err := w.walkNodes(elif.Body); err != nil {
				return err
			}
		}
		if n.Else != nil {
			return w.walkNodes(*n.Else)
		}
		return nil

	case *parser.For:
		if err := w.dispatch(n.Iter); err != nil {
			return err
		}
		return w.walkNodes(n.Body)

	case *parser.Do:
		if err := w.walkNodes(n.Body); err != nil {
			return err
		}
		return w.dispatch(n.Cond)

	case *parser.Switch:
		if err := w.dispatch(n.Operand); err != nil {
			return err
		}
		for _, c := range n.Cases {
			if err := w.dispatch(c.Pattern.Expr); err != nil {
				return err
			}
			if c.Result.Expr != nil {
				if err := w.dispatch(*c.Result.Expr); err != nil {
					return err
				}
			} else {
				if err := w.walkNodes(c.Result.Block); err != nil {
					return err
				}
			}
		}
		if n.Default.Expr != nil {
			return w.dispatch(*n.Default.Expr)
		}
		return w.walkNodes(n.Default.Block)

	case *parser.Struct:
		if err := w.walkNodes(n.Generics); err != nil {
			return err
		}
		if err := w.walkNodes(n.Interfaces); err != nil {
			return err
		}
		if err := w.walkNodes(parser.MapToNodes(n.Fields)); err != nil {
			return err
		}
		for _, init := range n.Inits {
			if err := w.walkDeclArgs(init.Params); err != nil {
				return err
			}
			if err := w.walkNodes(init.Body); err != nil {
				return err
			}
		}
		return nil

	case *parser.Record:
		if err := w.walkNodes(n.Generics); err != nil {
			return err
		}
		return w.walkNodes(n.Fields)

	case *parser.Interface:
		if err := w.walkNodes(n.Generics); err != nil {
			return err
		}
		return w.walkNodes(n.Members)

	case *parser.Enum:
		for _, v := range n.Variants {
			if err := w.walkDeclArgs(v.Params); err != nil {
				return err
			}
		}
		return nil

	case *parser.Variant:
		for _, f := range n.Fields {
			if err := w.dispatch(f.Type); err != nil {
				return err
			}
		}
		return nil

	case *parser.Alias:
		return w.dispatch(n.Type)

	case *parser.Const:
		if err := w.dispatch(n.Type); err != nil {
			return err
		}
		return w.dispatch(n.Expr)

	case *parser.FnSig:
		if err := w.walkDeclArgs(n.Args); err != nil {
			return err
		}
		return w.walkNodes(n.Returns)

	case *parser.BaseType:
		return w.walkNodes(n.TypeArgs)

	case *parser.FunctionType:
		if err := w.walkNodes(n.Args); err != nil {
			return err
		}
		return w.walkNodes(n.Returns)

	case *parser.ArrayType:
		return w.dispatch(n.Values)

	case *parser.ArrayLit:
		return w.walkNodes(n.Elements)

	case *parser.MapLit:
		return w.walkNodes(n.Elements)

	case *parser.MapPair:
		if err := w.dispatch(n.Key); err != nil {
			return err
		}
		return w.dispatch(n.Value)

	case *parser.IntLit, *parser.FloatLit, *parser.StringLit,
		*parser.CharLit, *parser.BoolLit, *parser.NoneLit,
		*parser.Ident, *parser.SelfField, *parser.Annotation:
		return nil
	}

	return nil
}

func (w *Walker) walkNodes(nodes []parser.Node) error {
	for _, n := range nodes {
		if err := w.dispatch(n); err != nil {
			return err
		}
	}
	return nil
}

func (w *Walker) walkArgs(args []parser.Arg) error {
	for _, a := range args {
		if err := w.dispatch(a.Value); err != nil {
			return err
		}
	}
	return nil
}

func (w *Walker) walkDeclArgs(args []parser.DeclArg) error {
	for _, a := range args {
		if err := w.dispatch(a.Type); err != nil {
			return err
		}
	}
	return nil
}
