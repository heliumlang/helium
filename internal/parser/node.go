package parser

import (
	"fmt"
	"strings"

	"github.com/Nykenik24/oxy/internal/lexer"
)

type Span struct{ Start, End lexer.Position }

type base struct{ Span Span }

func (n base) GetSpan() Span { return n.Span }

type Node interface {
	GetSpan() Span
	String() string
	tree() *treeNode
}

type treeNode struct {
	label    string
	children []*treeNode
}

func leaf(label string) *treeNode {
	return &treeNode{label: label}
}

func branch(label string, children ...*treeNode) *treeNode {
	var filtered []*treeNode
	for _, c := range children {
		if c != nil {
			filtered = append(filtered, c)
		}
	}
	return &treeNode{label: label, children: filtered}
}

func nodeTree(n Node) *treeNode {
	if n == nil {
		return nil
	}
	return n.tree()
}

func (t *treeNode) String() string {
	var b strings.Builder
	t.write(&b, "", "")
	return b.String()
}

func (t *treeNode) write(b *strings.Builder, prefix, connector string) {
	fmt.Fprintf(b, "%s%s%s\n", prefix, connector, t.label)
	childPrefix := prefix
	switch connector {
	case "└── ":
		childPrefix += "    "
	case "├── ":
		childPrefix += "│   "
	}
	for i, child := range t.children {
		if i == len(t.children)-1 {
			child.write(b, childPrefix, "└── ")
		} else {
			child.write(b, childPrefix, "├── ")
		}
	}
}

func join[T fmt.Stringer](s []T) string {
	var b strings.Builder
	for i, v := range s {
		b.WriteString(v.String())
		if i < len(s)-1 {
			b.WriteString(", ")
		}
	}
	return b.String()
}

func arrToNodes[T any](s []T, fn func(T) *treeNode) []*treeNode {
	var nodes []*treeNode

	for _, v := range s {
		nodes = append(nodes, fn(v))
	}

	return nodes
}

func nodesToChildren(nodes []Node) []*treeNode {
	out := make([]*treeNode, len(nodes))
	for i, n := range nodes {
		out[i] = nodeTree(n)
	}
	return out
}

func argListChildren(args []Arg) []*treeNode {
	out := make([]*treeNode, len(args))
	for i, a := range args {
		out[i] = a.tree()
	}
	return out
}

func funcArgChildren(args []DeclArg) []*treeNode {
	out := make([]*treeNode, len(args))
	for i, a := range args {
		out[i] = a.tree()
	}
	return out
}

func genericLabel(base string, generics []Node) string {
	if len(generics) == 0 {
		return base
	}
	var gs []string
	for _, g := range generics {
		gs = append(gs, g.String())
	}
	return base + "<" + strings.Join(gs, ", ") + ">"
}

func paramsChild(args []DeclArg) *treeNode {
	if len(args) == 0 {
		return nil
	}
	return branch("params", funcArgChildren(args)...)
}

func bodyChild(body []Node) *treeNode {
	if len(body) == 0 {
		return nil
	}
	return branch("body", nodesToChildren(body)...)
}

func returnsChild(returns []Node) *treeNode {
	if len(returns) == 0 || returns == nil {
		return leaf("returns: void")
	}
	return leaf("returns: " + join(returns))
}

type Program struct {
	base
	Items []Node
}

func (n Program) tree() *treeNode { return branch("Program", nodesToChildren(n.Items)...) }
func (n Program) String() string  { return n.tree().String() }

type Module struct {
	base
	Name string
}
type Use struct {
	base
	Members  []string
	From     *string
	Wildcard bool
}

func (n Use) tree() *treeNode {
	var children []*treeNode

	if !n.Wildcard {
		children = append(children, branch("members", arrToNodes(n.Members, func(s string) *treeNode {
			return leaf(s)
		})...))
	} else {
		children = append(children, leaf("all members"))
	}

	if n.From != nil {
		children = append(children, leaf("from:"+*n.From))
	}

	return branch("Use", children...)
}
func (n Module) tree() *treeNode { return leaf(fmt.Sprintf("Module(%q)", n.Name)) }

func (n Module) String() string { return n.tree().String() }
func (n Use) String() string    { return n.tree().String() }

type BaseType struct {
	base
	Typename  string
	TypeArgs  []Node
	Optional  bool
	Throwable bool
}

func (n BaseType) tree() *treeNode {
	label := n.Typename
	if n.Optional {
		label += "?"
	}
	if n.Throwable {
		label += "!"
	}
	if len(n.TypeArgs) == 0 {
		return leaf(label)
	}
	return branch(label, nodesToChildren(n.TypeArgs)...)
}
func (n BaseType) String() string {
	s := n.Typename
	if len(n.TypeArgs) > 0 {
		var args []string
		for _, a := range n.TypeArgs {
			args = append(args, a.String())
		}
		s += "<" + strings.Join(args, ", ") + ">"
	}
	if n.Optional {
		s += "?"
	}
	if n.Throwable {
		s += "!"
	}
	return s
}

type FunctionType struct {
	base
	Args    []Node
	Returns []Node
}

func (n FunctionType) tree() *treeNode {
	children := nodesToChildren(n.Args)
	children = append(children, returnsChild(n.Returns))
	return branch("fn type", children...)
}
func (n FunctionType) String() string {
	var args []string
	for _, a := range n.Args {
		args = append(args, a.String())
	}
	ret := "void"
	if n.Returns != nil {
		ret = join(n.Returns)
	}
	return fmt.Sprintf("fn(%s) %s", strings.Join(args, ", "), ret)
}

type ArrayType struct {
	base
	Values Node
}

func (n ArrayType) tree() *treeNode { return branch("[]", nodeTree(n.Values)) }
func (n ArrayType) String() string  { return n.Values.String() + "[]" }

type IntLit struct {
	base
	Value int64
}
type FloatLit struct {
	base
	Value float64
}
type StringLit struct {
	base
	Value string
}
type CharLit struct {
	base
	Value byte
}
type BoolLit struct {
	base
	Value bool
}
type NoneLit struct{ base }

func (n IntLit) tree() *treeNode    { return leaf(fmt.Sprintf("int: %d", n.Value)) }
func (n FloatLit) tree() *treeNode  { return leaf(fmt.Sprintf("float: %g", n.Value)) }
func (n StringLit) tree() *treeNode { return leaf(fmt.Sprintf("str: %q", n.Value)) }
func (n CharLit) tree() *treeNode   { return leaf(fmt.Sprintf("char: '%c'", n.Value)) }
func (n BoolLit) tree() *treeNode   { return leaf(fmt.Sprintf("bool: %t", n.Value)) }
func (n NoneLit) tree() *treeNode   { return leaf("none") }

func (n IntLit) String() string    { return n.tree().String() }
func (n FloatLit) String() string  { return n.tree().String() }
func (n StringLit) String() string { return n.tree().String() }
func (n CharLit) String() string   { return n.tree().String() }
func (n BoolLit) String() string   { return n.tree().String() }
func (n NoneLit) String() string   { return n.tree().String() }

type ArrayLit struct {
	base
	Elements []Node
}

func (n ArrayLit) tree() *treeNode { return branch("[]", nodesToChildren(n.Elements)...) }
func (n ArrayLit) String() string  { return n.tree().String() }

type MapPair struct {
	base
	Key   Node
	Value Node
}
type MapLit struct {
	base
	Elements []Node
}

func (n MapPair) tree() *treeNode {
	return branch("pair", nodesToChildren([]Node{n.Key, n.Value})...)
}
func (n MapLit) tree() *treeNode {
	return branch("{}", nodesToChildren(n.Elements)...)
}

func (n MapPair) String() string { return n.tree().String() }
func (n MapLit) String() string  { return n.tree().String() }

type Ident struct {
	base
	Name string
}
type SelfField struct {
	base
	Field string
}

func (n Ident) tree() *treeNode     { return leaf(n.Name) }
func (n SelfField) tree() *treeNode { return leaf("@" + n.Field) }
func (n Ident) String() string      { return n.tree().String() }
func (n SelfField) String() string  { return n.tree().String() }

type UnaryExpr struct {
	base
	Op      string
	Operand Node
}

type BinaryExpr struct {
	base
	Op          string
	Left, Right Node
}

type TernaryExpr struct {
	base
	Cond, Then, Else Node
}

func (n UnaryExpr) tree() *treeNode {
	return branch(n.Op, nodeTree(n.Operand))
}
func (n BinaryExpr) tree() *treeNode {
	return branch(n.Op, nodeTree(n.Left), nodeTree(n.Right))
}
func (n TernaryExpr) tree() *treeNode {
	cond := nodeTree(n.Cond)
	cond.label = "cond: " + cond.label
	then := nodeTree(n.Then)
	then.label = "then: " + then.label
	els := nodeTree(n.Else)
	els.label = "else: " + els.label
	return branch("?:", cond, then, els)
}

func (n UnaryExpr) String() string   { return n.tree().String() }
func (n BinaryExpr) String() string  { return n.tree().String() }
func (n TernaryExpr) String() string { return n.tree().String() }

type CoalesceExpr struct {
	base
	Block []Node
	Left  Node
	Right Node
}
type GroupExpr struct {
	base
	Inner Node
}
type FieldAccess struct {
	base
	Object Node
	Field  string
}
type IndexExpr struct {
	base
	Object, Index Node
}
type ForceUnwrap struct {
	base
	Operand Node
}
type OptionalChain struct {
	base
	Operand Node
}

func (n CoalesceExpr) tree() *treeNode {
	var children []*treeNode
	children = append(children, branch("left", nodeTree(n.Left)))
	if len(n.Block) > 0 {
		children = append(children, branch("block", nodesToChildren(n.Block)...))
	} else {
		children = append(children, branch("right", nodeTree(n.Right)))
	}
	return branch("??", children...)
}
func (n GroupExpr) tree() *treeNode     { return branch("()", nodeTree(n.Inner)) }
func (n FieldAccess) tree() *treeNode   { return branch("."+n.Field, nodeTree(n.Object)) }
func (n IndexExpr) tree() *treeNode     { return branch("[]", nodeTree(n.Object), nodeTree(n.Index)) }
func (n ForceUnwrap) tree() *treeNode   { return branch("!", nodeTree(n.Operand)) }
func (n OptionalChain) tree() *treeNode { return branch("?", nodeTree(n.Operand)) }

func (n CoalesceExpr) String() string  { return n.tree().String() }
func (n GroupExpr) String() string     { return n.tree().String() }
func (n FieldAccess) String() string   { return n.tree().String() }
func (n IndexExpr) String() string     { return n.tree().String() }
func (n ForceUnwrap) String() string   { return n.tree().String() }
func (n OptionalChain) String() string { return n.tree().String() }

type Arg struct {
	base
	Name   *string
	Value  Node
	Spread bool
}

type DeclArg struct {
	base
	Type     Node
	Name     string
	Variadic bool
}

func (n Arg) tree() *treeNode {
	label := "<pos>"
	if n.Name != nil {
		label = *n.Name + ": "
	}
	if n.Spread {
		label += "..."
	}
	return branch(label, nodeTree(n.Value))
}
func (n Arg) String() string { return n.tree().String() }

func (n DeclArg) tree() *treeNode {
	label := n.Name
	if n.Variadic {
		label += "..."
	}
	return branch(label, nodeTree(n.Type))
}

func (n DeclArg) String() string { return n.tree().String() }

type CallExpr struct {
	base
	Callee Node
	Args   []Arg
}

type MethodCall struct {
	base
	Object Node
	Method string
	Args   []Arg
}

type NewExpr struct {
	base
	TypeName string
	TypeArgs []Node
	Args     []Arg
}

func (n CallExpr) tree() *treeNode {
	return branch("call", append([]*treeNode{nodeTree(n.Callee)}, argListChildren(n.Args)...)...)
}
func (n MethodCall) tree() *treeNode {
	return branch("."+n.Method+"()", append([]*treeNode{nodeTree(n.Object)}, argListChildren(n.Args)...)...)
}
func (n NewExpr) tree() *treeNode {
	label := "new " + genericLabel(n.TypeName, n.TypeArgs)
	return branch(label, argListChildren(n.Args)...)
}

func (n CallExpr) String() string   { return n.tree().String() }
func (n MethodCall) String() string { return n.tree().String() }
func (n NewExpr) String() string    { return n.tree().String() }

type CatchExpr struct {
	base
	Operand  Node
	ErrIdent string
	Body     []Node
}

func (n CatchExpr) tree() *treeNode {
	return branch("catch "+n.ErrIdent,
		append([]*treeNode{nodeTree(n.Operand)}, nodesToChildren(n.Body)...)...,
	)
}
func (n CatchExpr) String() string { return n.tree().String() }

type ClosureExpr struct {
	base
	Params  []DeclArg
	Returns []Node
	Body    []Node
}

func (n ClosureExpr) tree() *treeNode {
	ret := "void"
	if len(n.Returns) > 0 {
		ret = join(n.Returns)
	}
	return branch("fn",
		paramsChild(n.Params),
		leaf("returns: "+ret),
		bodyChild(n.Body),
	)
}
func (n ClosureExpr) String() string { return n.tree().String() }

type VarDecl struct {
	base
	Idents []string
	Exprs  []Node
	Const  bool
}

type Return struct {
	base
	Exprs []Node
}

type Raise struct {
	base
	Expr Node
}

type ExprStmt struct {
	base
	Expr Node
}

func (n VarDecl) tree() *treeNode {
	prefix := ""
	if n.Const {
		prefix = "const"
	}
	return branch(prefix+" := "+strings.Join(n.Idents, ", "), nodesToChildren(n.Exprs)...)
}
func (n Return) tree() *treeNode   { return branch("return", nodesToChildren(n.Exprs)...) }
func (n Raise) tree() *treeNode    { return branch("raise", nodeTree(n.Expr)) }
func (n ExprStmt) tree() *treeNode { return nodeTree(n.Expr) }

func (n VarDecl) String() string  { return n.tree().String() }
func (n Return) String() string   { return n.tree().String() }
func (n Raise) String() string    { return n.tree().String() }
func (n ExprStmt) String() string { return n.tree().String() }

type Elif struct {
	base
	Cond Node
	Body []Node
}

type IfStmt struct {
	base
	Cond  Node
	Body  []Node
	Elifs []Elif
	Else  *[]Node
}

type For struct {
	base
	Idents []string
	Iter   Node
	Body   []Node
}

type Do struct {
	base
	Cond Node
	Body []Node
}

func (n Elif) tree() *treeNode {
	cond := nodeTree(n.Cond)
	cond.label = "cond: " + cond.label
	return branch("elif", cond, bodyChild(n.Body))
}
func (n Elif) String() string { return n.tree().String() }

func (n IfStmt) tree() *treeNode {
	cond := nodeTree(n.Cond)
	cond.label = "cond: " + cond.label
	children := []*treeNode{cond, branch("then", nodesToChildren(n.Body)...)}
	for _, elif := range n.Elifs {
		children = append(children, elif.tree())
	}
	if n.Else != nil {
		children = append(children, branch("else", nodesToChildren(*n.Else)...))
	}
	return branch("if", children...)
}
func (n IfStmt) String() string { return n.tree().String() }

func (n For) tree() *treeNode {
	idents := leaf(strings.Join(n.Idents, ", "))
	return branch("for", idents, nodeTree(n.Iter), bodyChild(n.Body))
}
func (n For) String() string { return n.tree().String() }

func (n Do) tree() *treeNode {
	cond := nodeTree(n.Cond)
	cond.label = "while: " + cond.label
	return branch("do", bodyChild(n.Body), cond)
}
func (n Do) String() string { return n.tree().String() }

type SwitchResult struct {
	base
	Expr  *Node
	Block []Node
}

type SwitchPattern struct {
	base
	Expr   Node
	Params []string
}

type SwitchCase struct {
	base
	Pattern SwitchPattern
	Result  SwitchResult
}

type Switch struct {
	base
	Operand Node
	Cases   []SwitchCase
	Default SwitchResult
}

func (n SwitchPattern) tree() *treeNode {
	var children []*treeNode
	children = append(children, nodeTree(n.Expr))
	if len(n.Params) > 0 {
		paramsBranch := branch("params")
		for _, param := range n.Params {
			paramsBranch.children = append(paramsBranch.children, leaf(param))
		}
		children = append(children, paramsBranch)
	}
	return branch("pattern", children...)
}

func (n SwitchResult) tree() *treeNode {
	if n.Expr != nil {
		return nodeTree(*n.Expr)
	}
	return branch("block", nodesToChildren(n.Block)...)
}

func (n SwitchCase) tree() *treeNode {
	return branch("case", nodeTree(n.Pattern), n.Result.tree())
}

func (n Switch) tree() *treeNode {
	children := []*treeNode{nodeTree(n.Operand)}
	for _, c := range n.Cases {
		children = append(children, c.tree())
	}
	def := n.Default.tree()
	def.label = "default: " + def.label
	return branch("switch", append(children, def)...)
}

func (n Switch) String() string        { return n.tree().String() }
func (n SwitchPattern) String() string { return n.tree().String() }
func (n SwitchCase) String() string    { return n.tree().String() }
func (n SwitchResult) String() string  { return n.tree().String() }

type Annotation struct {
	base
	Name  string
	Value string
}

type Receiver struct {
	Name string
	Type Node
}

type FunctionDecl struct {
	base
	Name        string
	Args        []DeclArg
	TypeArgs    []Node
	Body        []Node
	Recv        *Receiver
	Returns     []Node
	Annotations []Annotation
}

type FnSig struct {
	base
	Name    string
	Args    []DeclArg
	Returns []Node
}

type StructField struct {
	base
	Type       Node
	Name       string
	Qualifiers []string
}

type Init struct {
	base
	Params []DeclArg
	Body   []Node
}

type Struct struct {
	base
	Name       string
	Generics   []Node
	Interfaces []Node
	Exported   []Node
	Unexported []Node
	Inits      []Init
}

type Record struct {
	base
	Name     string
	Generics []Node
	Fields   []Node
}

type Interface struct {
	base
	Name     string
	Generics []Node
	Members  []Node
}

type EnumVariant struct {
	base
	Name   string
	Params []DeclArg
}

type Enum struct {
	base
	Name     string
	Variants []EnumVariant
}

type VariantField struct {
	base
	Type Node
	Name string
}

type Variant struct {
	base
	Name   string
	Fields []VariantField
}

type Alias struct {
	base
	Name string
	Type Node
}

type Const struct {
	base
	Type Node
	Name string
	Expr Node
}

func (n Annotation) tree() *treeNode {
	if n.Value != "" {
		return leaf(fmt.Sprintf("@%s(%q)", n.Name, n.Value))
	}
	return leaf("@" + n.Name)
}
func (n Annotation) String() string { return n.tree().String() }

func (n FunctionDecl) tree() *treeNode {
	label := "fn " + n.Name
	if n.Recv != nil {
		label = fmt.Sprintf("fn [%s %s] %s", n.Recv.Name, n.Recv.Type.String(), n.Name)
	}
	var children []*treeNode
	if len(n.TypeArgs) > 0 {
		typeArgsBranch := branch("generics")
		for _, arg := range n.TypeArgs {
			typeArgsBranch.children = append(typeArgsBranch.children, nodeTree(arg))
		}
		children = append(children, typeArgsBranch)
	}
	for _, a := range n.Annotations {
		children = append(children, a.tree())
	}
	children = append(children, paramsChild(n.Args), returnsChild(n.Returns), bodyChild(n.Body))
	return branch(label, children...)
}
func (n FunctionDecl) String() string { return n.tree().String() }

func (n FnSig) tree() *treeNode {
	return branch("fn "+n.Name, paramsChild(n.Args), returnsChild(n.Returns))
}
func (n FnSig) String() string { return n.tree().String() }

func (n StructField) tree() *treeNode {
	label := n.Type.String() + " " + n.Name
	if len(n.Qualifiers) > 0 {
		label = strings.Join(n.Qualifiers, " ") + " " + label
	}
	return leaf(label)
}
func (n StructField) String() string { return n.tree().String() }

func (n Init) tree() *treeNode {
	return branch("init", paramsChild(n.Params), bodyChild(n.Body))
}
func (n Init) String() string { return n.tree().String() }

func (n Struct) tree() *treeNode {
	label := genericLabel("struct "+n.Name, n.Generics)
	var children []*treeNode
	if len(n.Interfaces) > 0 {
		children = append(children, leaf("is: "+join(n.Interfaces)))
	}
	if len(n.Unexported) > 0 {
		children = append(children, branch("unexported", nodesToChildren(n.Unexported)...))
	}
	if len(n.Exported) > 0 {
		children = append(children, branch("exported", nodesToChildren(n.Exported)...))
	}
	for _, init := range n.Inits {
		children = append(children, init.tree())
	}
	return branch(label, children...)
}
func (n Struct) String() string { return n.tree().String() }

func (n Record) tree() *treeNode {
	return branch(genericLabel("record "+n.Name, n.Generics), nodesToChildren(n.Fields)...)
}
func (n Record) String() string { return n.tree().String() }

func (n Interface) tree() *treeNode {
	return branch(genericLabel("interface "+n.Name, n.Generics), nodesToChildren(n.Members)...)
}
func (n Interface) String() string { return n.tree().String() }

func (n EnumVariant) tree() *treeNode {
	if len(n.Params) == 0 {
		return leaf(n.Name)
	}
	return branch(n.Name, funcArgChildren(n.Params)...)
}
func (n EnumVariant) String() string { return n.tree().String() }

func (n Enum) tree() *treeNode {
	children := make([]*treeNode, len(n.Variants))
	for i, v := range n.Variants {
		children[i] = v.tree()
	}
	return branch("enum "+n.Name, children...)
}
func (n Enum) String() string { return n.tree().String() }

func (n VariantField) tree() *treeNode {
	return leaf(n.Type.String() + " " + n.Name)
}
func (n VariantField) String() string { return n.tree().String() }

func (n Variant) tree() *treeNode {
	children := make([]*treeNode, len(n.Fields))
	for i, f := range n.Fields {
		children[i] = f.tree()
	}
	return branch("variant "+n.Name, children...)
}
func (n Variant) String() string { return n.tree().String() }

func (n Alias) tree() *treeNode {
	return branch("alias "+n.Name, leaf(n.Type.String()))
}
func (n Alias) String() string { return n.tree().String() }

func (n Const) tree() *treeNode {
	return branch("const "+n.Type.String()+" "+n.Name, nodeTree(n.Expr))
}
func (n Const) String() string { return n.tree().String() }
