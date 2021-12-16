package solcparser

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	solAntlr "github.com/umbracle/solidity-parser-go/antlr"
)

// exampleListener is an event-driven callback for the parser.
type exampleListener struct {
	// *solAntlr.BaseSolidityListener

	service reflect.Value
	funcMap map[string]*funcData
}

type funcData struct {
	inNum int
	reqt  []reflect.Type
	fv    reflect.Value
	isDyn bool
}

func (f *funcData) numParams() int {
	return f.inNum - 1
}

func (e *exampleListener) init() {
	e.funcMap = map[string]*funcData{}
	e.service = reflect.ValueOf(e)

	st := reflect.TypeOf(e)
	if st.Kind() == reflect.Struct {
		panic("bad")
	}

	for i := 0; i < st.NumMethod(); i++ {
		mv := st.Method(i)
		if mv.PkgPath != "" {
			// skip unexported methods
			continue
		}
		name := mv.Name
		if name == "Visit" {
			continue
		}
		if !strings.HasPrefix(name, "Visit") {
			continue
		}

		fd := &funcData{
			fv: mv.Func,
		}
		var err error
		if fd.inNum, fd.reqt, err = validateFunc(name, fd.fv, true); err != nil {
			panic(fmt.Sprintf("jsonrpc: %s", err))
		}
		// check if last item is a pointer
		if fd.numParams() != 0 {
			last := fd.reqt[fd.numParams()]
			if last.Kind() == reflect.Ptr {
				fd.isDyn = true
			}
		}
		e.funcMap[name] = fd
	}
}

func validateFunc(funcName string, fv reflect.Value, isMethod bool) (inNum int, reqt []reflect.Type, err error) {
	if funcName == "" {
		err = fmt.Errorf("funcName cannot be empty")
		return
	}

	ft := fv.Type()
	if ft.Kind() != reflect.Func {
		err = fmt.Errorf("function '%s' must be a function instead of %s", funcName, ft)
		return
	}

	inNum = ft.NumIn()
	outNum := ft.NumOut()

	if outNum != 1 {
		err = fmt.Errorf("unexpected number of output arguments in the function '%s': %d. Expected 1", funcName, outNum)
		return
	}

	reqt = make([]reflect.Type, inNum)
	for i := 0; i < inNum; i++ {
		reqt[i] = ft.In(i)
	}
	return
}

func (e *exampleListener) Visit(i antlr.Tree) INode {
	funcName := reflect.TypeOf(i).String()
	if funcName == "*antlr.TerminalNodeImpl" {
		return nil
	}
	funcName = strings.TrimPrefix(funcName, "*solcparser.")
	xx := strings.TrimSuffix(funcName, "Context")
	funcName = "Visit" + strings.TrimSuffix(funcName, "Context")

	fd, ok := e.funcMap[funcName]
	if !ok {
		panic(fmt.Sprintf("BUG: visit not found %s", funcName))
	}

	inArgs := make([]reflect.Value, fd.inNum)
	inArgs[0] = e.service
	inArgs[1] = reflect.ValueOf(i)

	output := fd.fv.Call(inArgs)[0]
	if output.IsNil() {
		return nil
	}

	ii := output.Interface().(INode)
	if !skipNode(xx) {
		ii.SetTypeName(xx)
	}
	return ii
}

func skipNode(i string) bool {
	skip := []string{
		"ContractPart",
		"Statement",
		"SimpleStatement",
		"Expression",
		"TypeName",
		"MappingKey",
		"PrimaryExpression",
		"Parameter",
		"FunctionTypeParameter",
	}
	for _, j := range skip {
		if j == i {
			return true
		}
	}
	return false
}

type Node struct {
	Type string `json:"type"`
}

func (n *Node) IsNode() {}

func (n *Node) SetTypeName(typ string) {
	n.Type = typ
}

func (n *Node) GetType() string {
	return n.Type
}

type INode interface {
	GetType() string
	IsNode()
	SetTypeName(s string)
}

// SourceUnit
type SourceUnit struct {
	Node

	Children []interface{}
}

func (e *exampleListener) VisitSourceUnit(ctx *solAntlr.SourceUnitContext) INode {
	decl := &SourceUnit{
		Children: []interface{}{},
	}
	for _, p := range ctx.GetChildren() {
		decl.Children = append(decl.Children, e.Visit(p))
	}
	return decl
}

type PragmaDirective struct {
	Node

	Name  string
	Value string
}

func (e *exampleListener) VisitPragmaDirective(ctx *solAntlr.PragmaDirectiveContext) INode {
	decl := &PragmaDirective{
		Name:  toText(ctx.PragmaName()),
		Value: toText(ctx.PragmaValue()),
	}
	return decl
}

// ContractDefinition
type ContractDefinition struct {
	Node

	Name          string `json:"name"`
	SubNodes      []interface{}
	BaseContracts []interface{}
	Kind          string
}

func (e *exampleListener) VisitContractDefinition(ctx *solAntlr.ContractDefinitionContext) INode {
	decl := &ContractDefinition{
		Name:          toText(ctx.Identifier()),
		Kind:          toText(ctx.GetChild(0)),
		SubNodes:      []interface{}{},
		BaseContracts: []interface{}{},
	}
	for _, i := range ctx.AllContractPart() {
		decl.SubNodes = append(decl.SubNodes, e.Visit(i))
	}
	for _, i := range ctx.AllInheritanceSpecifier() {
		decl.BaseContracts = append(decl.BaseContracts, e.Visit(i))
	}
	//addMeta(decl, ctx)
	return decl
}

type InheritanceSpecifier struct {
	Node

	BaseName  interface{}
	Arguments []interface{}
}

func (e *exampleListener) VisitInheritanceSpecifier(ctx *solAntlr.InheritanceSpecifierContext) INode {
	decl := &InheritanceSpecifier{
		BaseName:  e.Visit(ctx.UserDefinedTypeName()),
		Arguments: []interface{}{},
	}
	if expr := ctx.ExpressionList(); expr != nil {
		decl.Arguments = append(decl.Arguments, e.Visit(expr))
	}
	return decl
}

func (e *exampleListener) VisitContractPart(ctx *solAntlr.ContractPartContext) interface{} {
	return e.Visit(ctx.GetChild(0))
}

// StructDefinition
type StructDefinition struct {
	Node

	Name    string
	Members []interface{}
}

func (e *exampleListener) VisitStructDefinition(ctx *solAntlr.StructDefinitionContext) INode {
	members := []interface{}{}
	for _, i := range ctx.AllVariableDeclaration() {
		members = append(members, e.Visit(i))
	}
	decl := &StructDefinition{
		Name:    toText(ctx.Identifier()),
		Members: members,
	}
	return decl
}

// VariableDeclaration
type VariableDeclaration struct {
	Node

	Name            string
	TypeName        interface{}
	Identifier      interface{}
	IsIndexed       bool
	IsStateVar      bool
	IsDeclaredConst bool
	StorageLocation string
	Expression      interface{}
	Visibility      string
	Override        []interface{}
}

func (e *exampleListener) VisitVariableDeclaration(ctx *solAntlr.VariableDeclarationContext) INode {
	decl := &VariableDeclaration{
		Name:       toText(ctx.Identifier()),
		Identifier: e.Visit(ctx.Identifier()),
		TypeName:   e.Visit(ctx.TypeName()),
	}
	if ctx.StorageLocation() != nil {
		decl.StorageLocation = toText(ctx.StorageLocation())
	}
	return decl
}

type Identifier struct {
	Node

	Name string
}

func (e *exampleListener) VisitIdentifier(ctx *solAntlr.IdentifierContext) INode {
	decl := &Identifier{
		Name: toText(ctx),
	}
	return decl
}

type StateVariableDeclaration struct {
	Node

	Variables    []interface{}
	InitialValue interface{}
}

type StateVariableDeclarationVariable struct {
	VariableDeclaration
	IsInmutable bool
}

func (e *exampleListener) VisitStateVariableDeclaration(ctx *solAntlr.StateVariableDeclarationContext) INode {

	visibility := "default"
	if hasElem(ctx.AllInternalKeyword()) {
		visibility = "internal"
	} else if hasElem(ctx.AllPublicKeyword()) {
		visibility = "public"
	} else if hasElem(ctx.AllPrivateKeyword()) {
		visibility = "private"
	}

	var override []interface{}
	overSpec := ctx.AllOverrideSpecifier()
	if len(overSpec) != 0 {
		for _, i := range overSpec[0].(*solAntlr.OverrideSpecifierContext).AllUserDefinedTypeName() {
			override = append(override, e.Visit(i))
		}
	}

	vv := &StateVariableDeclarationVariable{
		VariableDeclaration: VariableDeclaration{
			Node:            Node{Type: "VariableDeclaration"},
			Name:            toText(ctx.Identifier()),
			TypeName:        e.Visit(ctx.TypeName()),
			Identifier:      e.Visit(ctx.Identifier()),
			IsDeclaredConst: hasElem(ctx.AllConstantKeyword()),
			Visibility:      visibility,
			IsStateVar:      true,
			Override:        override,
		},
		IsInmutable: hasElem(ctx.AllImmutableKeyword()),
	}
	if ctx.Expression() != nil {
		vv.Expression = e.Visit(ctx.Expression())
	}

	decl := &StateVariableDeclaration{
		Variables: []interface{}{vv},
	}
	return decl
}

// FunctionDefinition
type FunctionDefinition struct {
	Node

	Name             string
	Parameters       []interface{}
	Modifiers        []interface{}
	ReturnParameters interface{}
	Body             interface{}
	StateMutability  string
	Visibility       string
	IsConstructor    bool
	IsReceiveEther   bool
	IsFallback       bool
	IsVirtual        bool
	Override         []interface{}
}

func hasElem(a []antlr.TerminalNode) bool {
	return len(a) > 0
}

func (e *exampleListener) VisitFunctionDefinition(ctx *solAntlr.FunctionDefinitionContext) INode {
	var isConstructor, isFallback, isVirtual, isReceiveEther bool
	visibility := "default"

	var block interface{}
	if ctx.Block() != nil {
		block = e.Visit(ctx.Block())
	}

	modifiers := []interface{}{}
	for _, i := range ctx.ModifierList().(*solAntlr.ModifierListContext).AllModifierInvocation() {
		modifiers = append(modifiers, e.Visit(i))
	}

	modList := ctx.ModifierList().(*solAntlr.ModifierListContext)

	var parameters []interface{}
	params := ctx.ParameterList().(*solAntlr.ParameterListContext).AllParameter()
	for _, param := range params {
		parameters = append(parameters, e.Visit(param))
	}

	var name string
	var returnParameters interface{}
	switch toText(ctx.FunctionDescriptor().(*solAntlr.FunctionDescriptorContext).GetChild(0)) {
	case "constructor":

		if hasElem(modList.AllInternalKeyword()) {
			visibility = "internal"
		} else if hasElem(modList.AllPublicKeyword()) {
			visibility = "public"
		} else {
			visibility = "default"
		}
		isConstructor = true

	case "fallback":
		visibility = "external"
		isFallback = true

	case "receive":
		visibility = "external"
		isReceiveEther = true

	case "function":
		ident := ctx.FunctionDescriptor().(*solAntlr.FunctionDescriptorContext).Identifier()
		if ident != nil {
			name = toText(ident)
		}

		ctxRet := ctx.ReturnParameters()
		if ctxRet != nil {
			returnParameters = e.VisitReturnParameters(ctxRet.(*solAntlr.ReturnParametersContext))
		}

		if hasElem(modList.AllExternalKeyword()) {
			visibility = "external"
		} else if hasElem(modList.AllInternalKeyword()) {
			visibility = "internal"
		} else if hasElem(modList.AllPublicKeyword()) {
			visibility = "public"
		} else if hasElem(modList.AllPrivateKeyword()) {
			visibility = "private"
		}

		isFallback = name == ""
	}

	if hasElem(modList.AllVirtualKeyword()) {
		isVirtual = true
	}

	var override []interface{}
	overrideSpec := modList.AllOverrideSpecifier()
	if len(overrideSpec) != 0 {
		for _, o := range overrideSpec[0].(*solAntlr.OverrideSpecifierContext).AllUserDefinedTypeName() {
			override = append(override, e.Visit(o))
		}
	}

	decl := &FunctionDefinition{
		Name:             name,
		Body:             block,
		Parameters:       parameters,
		ReturnParameters: returnParameters,
		IsConstructor:    isConstructor,
		IsFallback:       isFallback,
		IsVirtual:        isVirtual,
		IsReceiveEther:   isReceiveEther,
		Visibility:       visibility,
		Modifiers:        modifiers,
		Override:         override,
	}
	return decl
}

func (e *exampleListener) VisitParameter(ctx *solAntlr.ParameterContext) INode {
	decl := &VariableDeclaration{
		Node:       Node{Type: "VariableDeclaration"},
		TypeName:   e.Visit(ctx.TypeName()),
		IsStateVar: false,
		IsIndexed:  false,
		Expression: nil,
	}

	if ctx.StorageLocation() != nil {
		decl.StorageLocation = toText(ctx.StorageLocation())
	}
	if iden := ctx.Identifier(); iden != nil {
		decl.Name = toText(iden)
		decl.Identifier = e.Visit(iden)
	}
	return decl
}

type ModifierInvocation struct {
	Node

	Name      string
	Arguments []interface{}
}

func (e *exampleListener) VisitModifierInvocation(ctx *solAntlr.ModifierInvocationContext) INode {
	var args []interface{}
	if expr := ctx.ExpressionList(); expr != nil {
		for _, p := range expr.(*solAntlr.ExpressionListContext).AllExpression() {
			args = append(args, e.Visit(p))
		}
	} else if child := ctx.GetChildren(); len(child) > 1 {
		args = []interface{}{}
	}

	decl := &ModifierInvocation{
		Name:      toText(ctx.Identifier()),
		Arguments: args,
	}
	return decl
}

// Block
type Block struct {
	Node

	Statements []interface{}
}

func (e *exampleListener) VisitBlock(ctx *solAntlr.BlockContext) INode {
	decl := &Block{
		Statements: []interface{}{},
	}
	for _, i := range ctx.AllStatement() {
		decl.Statements = append(decl.Statements, e.Visit(i))
	}
	return decl
}

func (e *exampleListener) VisitStatement(ctx *solAntlr.StatementContext) INode {
	return e.Visit(ctx.GetChild(0))
}

func (e *exampleListener) VisitSimpleStatement(ctx *solAntlr.SimpleStatementContext) INode {
	return e.Visit(ctx.GetChild(0))
}

// ExpressionStatement
type ExpressionStatement struct {
	Node

	Expression interface{}
}

func (e *exampleListener) VisitExpressionStatement(ctx *solAntlr.ExpressionStatementContext) INode {
	if ctx == nil {
		return nil
	}
	decl := &ExpressionStatement{
		Expression: e.Visit(ctx.Expression()),
	}
	return decl
}

// Expression
type NewExpression struct {
	Node

	TypeName interface{}
}

type UnaryOperation struct {
	Node

	Operator      string
	SubExpression interface{}
	IsPrefix      bool
}

type BinaryOperation struct {
	Node

	Operator string
	Left     interface{}
	Right    interface{}
}

type NumberLiteral struct {
	Node

	Number          string
	SubDenomination interface{}
}

type IndexRangeAccess struct {
	Node

	Base       interface{}
	IndexStart interface{}
	IndexEnd   interface{}
}

type Conditional struct {
	Node

	Condition       interface{}
	TrueExpression  interface{}
	FalseExpression interface{}
}

type MemberAccess struct {
	Node

	Expression interface{}
	MemberName string
}

type TupleExpression struct {
	Node

	Components []interface{}
	IsArray    bool
}

type NameValueExpression struct {
	Node

	Expression interface{}
	Arguments  interface{}
}

type IndexAccess struct {
	Node

	Base  interface{}
	Index interface{}
}

type ArrayTypeName struct {
	Node

	BaseTypeName interface{}
	Length       interface{}
}

var (
	unaryOps   = []string{"-", "+", "--", "++", "~", "after", "delete", "!"}
	postFixOps = []string{"++", "--"}
	binaryOps  = []string{"+", "-", "*", "/", "**", "%", "<<", ">>", "&&", "||", ",,", "&", ",", "^", "<", ">", "<=", ">=", "==", "!=", "=", ",=", "^=", "&=", "<<=", ">>=", "+=", "-=", "*=", "/=", "%=", "|", "|="}
)

func contains(list []string, s string) bool {
	for _, op := range list {
		if s == op {
			return true
		}
	}
	return false
}

func (e *exampleListener) VisitExpression(ctx *solAntlr.ExpressionContext) INode {
	children := ctx.GetChildren()
	switch len(children) {
	case 1:
		return e.Visit(children[0])
	case 2:
		op := toText(ctx.GetChild(0))

		// new expression
		if op == "new" {
			decl := &NewExpression{
				Node:     Node{Type: "NewExpression"},
				TypeName: e.Visit(ctx.TypeName()),
			}
			return decl
		}

		// prefix operators
		if contains(unaryOps, op) {
			decl := &UnaryOperation{
				Node:          Node{Type: "UnaryOperation"},
				Operator:      op,
				SubExpression: e.Visit(ctx.Expression(0)),
				IsPrefix:      true,
			}
			return decl
		}

		op = toText(ctx.GetChild(1))

		if contains(postFixOps, op) {
			decl := &UnaryOperation{
				Node:          Node{Type: "UnaryOperation"},
				Operator:      op,
				SubExpression: e.Visit(ctx.Expression(0)),
				IsPrefix:      false,
			}
			return decl
		}
	case 3:
		if toText(ctx.GetChild(0)) == "(" && toText(ctx.GetChild(2)) == ")" {
			decl := &TupleExpression{
				Node:    Node{Type: "TupleExpression"},
				IsArray: false,
				Components: []interface{}{
					e.Visit(ctx.GetChild(1)),
				},
			}
			return decl
		}

		op := toText(ctx.GetChild(1))

		// member access
		if op == "." {
			decl := &MemberAccess{
				Node:       Node{Type: "MemberAccess"},
				Expression: e.Visit(ctx.Expression(0)),
				MemberName: toText(ctx.Identifier()),
			}
			return decl
		}

		if contains(binaryOps, op) {
			decl := &BinaryOperation{
				Node:     Node{Type: "BinaryOperation"},
				Operator: op,
				Left:     e.Visit(ctx.Expression(0)),
				Right:    e.Visit(ctx.Expression(1)),
			}
			return decl
		}
	case 4:
		// function call
		if toText(ctx.GetChild(1)) == "(" && toText(ctx.GetChild(3)) == ")" {
			var names, args, identifiers []interface{}

			ctxArgs := ctx.FunctionCallArguments().(*solAntlr.FunctionCallArgumentsContext)
			if expr := ctxArgs.ExpressionList(); expr != nil {
				for _, p := range expr.(*solAntlr.ExpressionListContext).AllExpression() {
					args = append(args, e.Visit(p))
				}
			} else if expr := ctxArgs.NameValueList(); expr != nil {
				for _, raw := range expr.(*solAntlr.NameValueListContext).AllNameValue() {
					p := raw.(*solAntlr.NameValueContext)
					args = append(args, e.Visit(p.Expression()))
					names = append(names, toText(p.Identifier()))
					identifiers = append(identifiers, e.Visit(p.Identifier()))
				}
			}

			decl := &FunctionCall{
				Node:        Node{Type: "FunctionCall"},
				Expression:  e.Visit(ctx.Expression(0)),
				Names:       names,
				Identifiers: identifiers,
				Arguments:   args,
			}
			return decl
		}

		// index access
		if toText(ctx.GetChild(1)) == "[" && toText(ctx.GetChild(3)) == "]" {
			if toText(ctx.GetChild(2)) == ":" {
				decl := &IndexRangeAccess{
					Node: Node{Type: "IndexRangeAccess"},
					Base: e.Visit(ctx.Expression(0)),
				}
				return decl
			}

			decl := &IndexAccess{
				Node:  Node{Type: "IndexAccess"},
				Base:  e.Visit(ctx.Expression(0)),
				Index: e.Visit(ctx.Expression(1)),
			}
			return decl
		}

		// expression with nameValueList
		if toText(ctx.GetChild(1)) == "{" && toText(ctx.GetChild(3)) == "}" {
			decl := &NameValueExpression{
				Node:       Node{Type: "NameValueExpression"},
				Expression: e.Visit(ctx.Expression(0)),
				Arguments:  e.Visit(ctx.NameValueList()),
			}
			return decl
		}

	case 5:
		// ternary operator
		if toText(ctx.GetChild(1)) == "?" && toText(ctx.GetChild(3)) == ":" {
			decl := &Conditional{
				Node:            Node{Type: "Conditional"},
				Condition:       e.Visit(ctx.Expression(0)),
				TrueExpression:  e.Visit(ctx.Expression(1)),
				FalseExpression: e.Visit(ctx.Expression(2)),
			}
			return decl
		}

		// index range access
		if toText(ctx.GetChild(1)) == "[" && toText(ctx.GetChild(2)) == ":" && toText(ctx.GetChild(4)) == "]" {
			decl := &IndexRangeAccess{
				Node:     Node{Type: "IndexRangeAccess"},
				Base:     e.Visit(ctx.Expression(0)),
				IndexEnd: e.Visit(ctx.Expression(1)),
			}
			return decl

		} else if toText(ctx.GetChild(1)) == "[" && toText(ctx.GetChild(3)) == ":" && toText(ctx.GetChild(4)) == "]" {
			decl := &IndexRangeAccess{
				Node:       Node{Type: "IndexRangeAccess"},
				Base:       e.Visit(ctx.Expression(0)),
				IndexStart: e.Visit(ctx.Expression(1)),
			}
			return decl
		}

	case 6:
		// index range access
		if toText(ctx.GetChild(1)) == "[" && toText(ctx.GetChild(3)) == ":" && toText(ctx.GetChild(5)) == "]" {
			decl := &IndexRangeAccess{
				Node:       Node{Type: "IndexRangeAccess"},
				Base:       e.Visit(ctx.Expression(0)),
				IndexStart: e.Visit(ctx.Expression(1)),
				IndexEnd:   e.Visit(ctx.Expression(2)),
			}
			return decl
		}
	}

	panic("TODO")
}

type Mapping struct {
	Node

	KeyType   interface{}
	ValueType interface{}
}

func (e *exampleListener) VisitMapping(ctx *solAntlr.MappingContext) INode {
	decl := &Mapping{
		KeyType:   e.Visit(ctx.MappingKey()),
		ValueType: e.Visit(ctx.TypeName()),
	}
	return decl
}

func (e *exampleListener) VisitMappingKey(ctx *solAntlr.MappingKeyContext) INode {
	if elem := ctx.ElementaryTypeName(); elem != nil {
		return e.Visit(elem)
	}
	if elem := ctx.UserDefinedTypeName(); elem != nil {
		return e.Visit(elem)
	}
	panic("BUG")
}

type NameValueList struct {
	Node

	Names       []string
	Identifiers []interface{}
	Args        []interface{}
}

func (e *exampleListener) VisitNameValueList(ctx *solAntlr.NameValueListContext) INode {
	decl := &NameValueList{
		Names:       []string{},
		Identifiers: []interface{}{},
		Args:        []interface{}{},
	}
	for _, raw := range ctx.AllNameValue() {
		p := raw.(*solAntlr.NameValueContext)
		decl.Names = append(decl.Names, toText(p.Identifier()))
		decl.Identifiers = append(decl.Identifiers, e.Visit(p.Identifier()))
		decl.Args = append(decl.Args, e.Visit(p.Expression()))
	}
	return decl
}

func (e *exampleListener) VisitTupleExpression(ctx *solAntlr.TupleExpressionContext) INode {
	count := ctx.GetChildCount()
	if count < 2 {
		panic("bad")
	}

	childs := ctx.GetChildren()[1 : count-1]
	childs = mapCommasToNulls(childs)

	components := []interface{}{}
	for _, child := range childs {
		if child == nil {
			components = append(components, nil)
		} else {
			components = append(components, e.Visit(child))
		}
	}

	decl := &TupleExpression{
		Components: components,
		IsArray:    toText(ctx.GetChild(0)) == "[",
	}
	return decl
}

func mapCommasToNulls(tree []antlr.Tree) []antlr.Tree {
	res := []antlr.Tree{}
	if len(tree) == 0 {
		return res
	}

	comma := true
	for _, child := range tree {
		if comma {
			if toText(child) == "," {
				res = append(res, nil)
			} else {
				res = append(res, child)
				comma = false
			}
		} else {
			if toText(child) != "," {
				panic("Comma expected")
			}
			comma = true
		}
	}

	if comma {
		res = append(res, nil)
	}
	return res
}

type TypeNameExpression struct {
	Node

	TypeName interface{}
}

func (e *exampleListener) VisitTypeNameExpression(ctx *solAntlr.TypeNameExpressionContext) INode {
	var typeName interface{}
	if ctxElem := ctx.ElementaryTypeName(); ctxElem != nil {
		typeName = e.Visit(ctxElem)
	} else if ctxUser := ctx.UserDefinedTypeName(); ctxUser != nil {
		typeName = e.Visit(ctxUser)
	} else {
		panic("BAD")
	}

	decl := &TypeNameExpression{
		TypeName: typeName,
	}
	return decl
}

func (e *exampleListener) VisitNumberLiteral(ctx *solAntlr.NumberLiteralContext) INode {
	decl := &NumberLiteral{
		Number:          toText(ctx.GetChild(0)),
		SubDenomination: nil,
	}
	if ctx.GetChildCount() == 2 {
		decl.SubDenomination = toText(ctx.GetChild(1))
	}
	return decl
}

type HexLiteral struct {
	Node

	Value string
	Parts []string
}

func (e *exampleListener) VisitHexLiteral(ctx *solAntlr.HexLiteralContext) INode {
	decl := &HexLiteral{
		Parts: []string{},
	}

	for _, p := range ctx.AllHexLiteralFragment() {
		hex := toText(p)
		hex = strings.TrimPrefix(hex, "hex")
		hex = strings.Trim(hex, "\"")
		decl.Parts = append(decl.Parts, hex)
	}
	decl.Value = strings.Join(decl.Parts, "")
	return decl
}

type BooleanLiteral struct {
	Node

	Value bool
}

func (e *exampleListener) VisitPrimaryExpression(ctx *solAntlr.PrimaryExpressionContext) INode {
	if expr := ctx.BooleanLiteral(); expr != nil {
		decl := &BooleanLiteral{
			Node:  Node{Type: "BooleanLiteral"},
			Value: toText(expr) == "true",
		}
		return decl
	}

	if expr := ctx.HexLiteral(); expr != nil {
		return e.Visit(expr)
	}

	if expr := ctx.StringLiteral(); expr != nil {
		decl := &StringLiteral{
			Node:      Node{Type: "StringLiteral"},
			Parts:     []string{},
			IsUnicode: []bool{},
		}

		values := []string{}
		for _, g := range expr.(*solAntlr.StringLiteralContext).AllStringLiteralFragment() {
			text := toText(g)

			isUnicode := strings.HasPrefix(text, "unicode")
			if isUnicode {
				text = strings.TrimPrefix(text, "unicode")
			}
			// isSingleQuotes := string(text[0]) == "'"

			textWithoutQuotes := text[1 : len(text)-1]
			value := textWithoutQuotes
			values = append(values, value)

			decl.Parts = append(decl.Parts, value)
			decl.IsUnicode = append(decl.IsUnicode, isUnicode)
		}
		decl.Value = strings.Join(values, "")
		return decl
	}

	if expr := ctx.NumberLiteral(); expr != nil {
		return e.Visit(expr)
	}

	if expr := ctx.TypeKeyword(); expr != nil {
		decl := &Identifier{
			Node: Node{Type: "Identifier"},
			Name: "type",
		}
		return decl
	}

	if len(ctx.GetChildren()) == 3 && toText(ctx.GetChild(0)) == "[" && toText(ctx.GetChild(2)) == "]" {
		panic("TODO")
	}

	return e.Visit(ctx.GetChild(0))
}

// ForStatement
type ForStatement struct {
	Node

	InitExpression      interface{}
	ConditionExpression interface{}
	LoopExpression      interface{}
	Body                interface{}
}

func (e *exampleListener) VisitForStatement(ctx *solAntlr.ForStatementContext) INode {
	var conditionExpr interface{}
	if ctx.ExpressionStatement() != nil {
		conditionExpr = e.Visit(ctx.ExpressionStatement()).(*ExpressionStatement).Expression
	}
	decl := &ForStatement{
		Body:                e.Visit(ctx.Statement()),
		ConditionExpression: conditionExpr,
	}
	if ctx.SimpleStatement() != nil {
		decl.InitExpression = e.Visit(ctx.SimpleStatement())
	}

	loopExpr := &ExpressionStatement{
		Node:       Node{Type: "ExpressionStatement"},
		Expression: nil,
	}
	if ctx.Expression() != nil {
		loopExpr.Expression = e.Visit(ctx.Expression())
	}
	decl.LoopExpression = loopExpr
	return decl
}

// VariableDeclarationStatement
type VariableDeclarationStatement struct {
	Node

	InitialValue interface{}
	Variables    []interface{}
}

func (e *exampleListener) VisitVariableDeclarationStatement(ctx *solAntlr.VariableDeclarationStatementContext) INode {
	var variables []interface{}
	if expr := ctx.VariableDeclaration(); expr != nil {
		variables = []interface{}{e.Visit(expr)}
	} else if expr := ctx.IdentifierList(); expr != nil {
		variables = e.buildIdentifierList(expr.(*solAntlr.IdentifierListContext))
	} else if expr := ctx.VariableDeclarationList(); expr != nil {
		variables = e.buildVariableDeclarationList(expr.(*solAntlr.VariableDeclarationListContext))
	}

	decl := &VariableDeclarationStatement{
		Variables: variables,
	}

	if ctx.Expression() != nil {
		decl.InitialValue = e.Visit(ctx.Expression())
	}
	return decl
}

func (e *exampleListener) buildIdentifierList(ctx *solAntlr.IdentifierListContext) []interface{} {
	count := ctx.GetChildCount()
	if count < 2 {
		panic("bad")
	}

	childs := ctx.GetChildren()[1 : count-1]
	childs = mapCommasToNulls(childs)

	identifiers := ctx.AllIdentifier()

	indx := 0
	variables := []interface{}{}

	for _, child := range childs {
		if child == nil {
			variables = append(variables, nil)
		} else {
			iden := identifiers[indx]
			indx++

			variables = append(variables, &VariableDeclaration{
				Node:       Node{Type: "VariableDeclaration"},
				Name:       toText(iden),
				Identifier: e.Visit(iden),
			})
		}
	}
	return variables
}

func (e *exampleListener) buildVariableDeclarationList(ctx *solAntlr.VariableDeclarationListContext) []interface{} {
	childs := mapCommasToNulls(ctx.GetChildren())
	variableDeclarations := ctx.AllVariableDeclaration()

	indx := 0

	variables := []interface{}{}
	for _, child := range childs {
		if child == nil {
			variables = append(variables, nil)
		} else {
			decl := variableDeclarations[indx].(*solAntlr.VariableDeclarationContext)
			indx++

			iden := decl.Identifier()

			storageLocation := ""
			if decl.StorageLocation() != nil {
				storageLocation = toText(decl.StorageLocation())
			}

			variables = append(variables, &VariableDeclaration{
				Node:            Node{Type: "VariableDeclaration"},
				Name:            toText(iden),
				Identifier:      e.Visit(iden),
				TypeName:        e.Visit(decl.TypeName()),
				StorageLocation: storageLocation,
			})
		}
	}
	return variables
}

// WhileStatement
type WhileStatement struct {
	Node

	Condition interface{}
	Body      interface{}
}

func (e *exampleListener) VisitWhileStatement(ctx *solAntlr.WhileStatementContext) INode {
	decl := &WhileStatement{
		Condition: e.Visit(ctx.Expression()),
		Body:      e.Visit(ctx.Statement()),
	}
	return decl
}

type DoWhileStatement struct {
	Node

	Condition interface{}
	Body      interface{}
}

func (e *exampleListener) VisitDoWhileStatement(ctx *solAntlr.DoWhileStatementContext) INode {
	decl := &DoWhileStatement{
		Condition: e.Visit(ctx.Expression()),
		Body:      e.Visit(ctx.Statement()),
	}
	return decl
}

// IfStatement
type IfStatement struct {
	Node

	Condition interface{}
	TrueBody  interface{}
	FalseBody interface{}
}

func (e *exampleListener) VisitIfStatement(ctx *solAntlr.IfStatementContext) INode {
	decl := &IfStatement{
		Condition: e.Visit(ctx.Expression()),
		TrueBody:  e.Visit(ctx.Statement(0)),
	}
	if len(ctx.AllStatement()) > 1 {
		decl.FalseBody = e.Visit(ctx.Statement(1))
	}
	return decl
}

// EventDefinition
type EventDefinition struct {
	Node

	Name       string
	Parameters []interface{}
}

func (e *exampleListener) VisitEventDefinition(ctx *solAntlr.EventDefinitionContext) INode {
	decl := &EventDefinition{
		Name:       toText(ctx.Identifier()),
		Parameters: []interface{}{},
	}

	paramsList := ctx.EventParameterList().(*solAntlr.EventParameterListContext).AllEventParameter()
	for _, p := range paramsList {
		param := p.(*solAntlr.EventParameterContext)

		var identifier interface{}

		name := ""
		if iden := param.Identifier(); iden != nil {
			name = toText(iden)
			identifier = e.Visit(iden)
		}

		varDecl := &VariableDeclaration{
			Node:       Node{Type: "VariableDeclaration"},
			TypeName:   e.Visit(param.TypeName()),
			Name:       name,
			Identifier: identifier,
		}
		if param.IndexedKeyword() != nil {
			varDecl.IsIndexed = true
		}
		decl.Parameters = append(decl.Parameters, varDecl)
	}

	return decl
}

// ImportDirective
type ImportDirective struct {
	Node

	Path        string
	PathLiteral interface{}

	UnitAlias           string
	UnitAliasIdentifier interface{}

	SymbolAliases            [][]string
	SymbolAliasesIdentifiers [][]*Identifier
}

type StringLiteral struct {
	Node

	Value     string
	Parts     []string
	IsUnicode []bool
}

func (e *exampleListener) VisitImportDirective(ctx *solAntlr.ImportDirectiveContext) INode {
	var unitAliasIdentifier interface{}
	var unitAlias string

	var symbolAliases [][]string
	var symbolAliasesIdentifiers [][]*Identifier

	pathString := toText(ctx.ImportPath())
	path := strings.Trim(pathString, "\"")

	importDecl := ctx.AllImportDeclaration()
	if len(importDecl) > 0 {
		for _, i := range importDecl {
			ctx := i.(*solAntlr.ImportDeclarationContext)

			// symbol
			symbol := toText(ctx.Identifier(0))
			symbolIdentifier := e.Visit(ctx.Identifier(0)).(*Identifier)

			// alias
			alias := ""
			var aliasIdentifier *Identifier
			if len(ctx.AllIdentifier()) > 1 {
				alias = toText(ctx.Identifier(1))
				aliasIdentifier = e.Visit(ctx.Identifier(1)).(*Identifier)
			}

			symbolAliases = append(symbolAliases, []string{symbol, alias})
			symbolAliasesIdentifiers = append(symbolAliasesIdentifiers, []*Identifier{
				symbolIdentifier, aliasIdentifier,
			})
		}
	} else {
		idenList := ctx.AllIdentifier()
		var iden antlr.Tree

		switch len(idenList) {
		case 0:
			// nothing
		case 1:
			iden = idenList[0]
		case 2:
			iden = idenList[1]
		default:
			panic("BAD")
		}
		if iden != nil {
			unitAlias = toText(iden)
			unitAliasIdentifier = e.Visit(iden)
		}
	}

	decl := &ImportDirective{
		Path: path,
		PathLiteral: &StringLiteral{
			Node:      Node{Type: "StringLiteral"},
			Value:     path,
			Parts:     []string{path},
			IsUnicode: []bool{false},
		},
		UnitAlias:                unitAlias,
		UnitAliasIdentifier:      unitAliasIdentifier,
		SymbolAliases:            symbolAliases,
		SymbolAliasesIdentifiers: symbolAliasesIdentifiers,
	}
	return decl
}

// EnumDefinition
type EnumDefinition struct {
	Node

	Name    string        `json:"name"`
	Members []interface{} `json:"members"`
}

func (e *exampleListener) VisitEnumDefinition(ctx *solAntlr.EnumDefinitionContext) INode {
	decl := &EnumDefinition{
		Name:    toText(ctx.Identifier()),
		Members: []interface{}{},
	}
	for _, m := range ctx.AllEnumValue() {
		decl.Members = append(decl.Members, e.Visit(m))
	}
	return decl
}

type EnumValue struct {
	Node

	Name string `json:"name"`
}

func (e *exampleListener) VisitEnumValue(ctx *solAntlr.EnumValueContext) INode {
	decl := &EnumValue{
		Name: toText(ctx.Identifier()),
	}
	return decl
}

// BreakStatement
type BreakStatement struct {
	Node
}

func (e *exampleListener) VisitBreakStatement(ctx *solAntlr.BreakStatementContext) INode {
	decl := &BreakStatement{}
	return decl
}

// ContinueStatement
type ContinueStatement struct {
	Node
}

func (e *exampleListener) VisitContinueStatement(ctx *solAntlr.ContinueStatementContext) INode {
	decl := &ContinueStatement{}
	return decl
}

// ReturnStatement
type ReturnStatement struct {
	Node

	Expression interface{}
}

func (e *exampleListener) VisitReturnStatement(ctx *solAntlr.ReturnStatementContext) INode {
	decl := &ReturnStatement{}
	if ctx.Expression() != nil {
		decl.Expression = e.Visit(ctx.Expression())
	}
	return decl
}

// ModifierDefinition
type ModifierDefinition struct {
	Node

	Name       string
	Parameters interface{}
	Body       interface{}
	IsVirtual  bool
	Override   []interface{}
}

func (e *exampleListener) VisitModifierDefinition(ctx *solAntlr.ModifierDefinitionContext) INode {
	decl := &ModifierDefinition{
		Name: toText(ctx.Identifier()),
		//Parameters: []interface{}{},
		IsVirtual: false,
		Override:  []interface{}{},
	}
	if override := ctx.AllOverrideSpecifier(); len(override) > 0 {
		for _, i := range override[0].(*solAntlr.OverrideSpecifierContext).AllUserDefinedTypeName() {
			decl.Override = append(decl.Override, e.Visit(i))
		}
	}

	if len(ctx.AllVirtualKeyword()) > 0 {
		decl.IsVirtual = true
	}
	//if param := ctx.ParameterList(); param != nil {
	// decl.Parameters = e.VisitParameterList(param.(*solAntlr.ParameterListContext))
	//}
	if body := ctx.Block(); body != nil {
		decl.Body = e.Visit(body)
	}
	return decl
}

type OverrideSpecifier struct {
	Node
}

func (e *exampleListener) VisitOverrideSpecifier(ctx *solAntlr.OverrideSpecifierContext) INode {
	decl := &OverrideSpecifier{}
	return decl
}

type FunctionTypeName struct {
	Node

	ParameterTypes  []interface{}
	ReturnTypes     []interface{}
	Visibility      string
	StateMutability string
}

func (e *exampleListener) VisitFunctionTypeName(ctx *solAntlr.FunctionTypeNameContext) INode {
	decl := &FunctionTypeName{
		ParameterTypes: []interface{}{},
		ReturnTypes:    []interface{}{},
	}

	paramList := ctx.AllFunctionTypeParameterList()
	for _, p := range paramList[0].(*solAntlr.FunctionTypeParameterListContext).AllFunctionTypeParameter() {
		decl.ParameterTypes = append(decl.ParameterTypes, e.Visit(p))
	}

	if len(paramList) > 1 {
		for _, p := range paramList[1].(*solAntlr.FunctionTypeParameterListContext).AllFunctionTypeParameter() {
			decl.ReturnTypes = append(decl.ReturnTypes, e.Visit(p))
		}
	}

	visibility := "default"
	if hasElem(ctx.AllInternalKeyword()) {
		visibility = "internal"
	} else if hasElem(ctx.AllExternalKeyword()) {
		visibility = "external"
	}
	decl.Visibility = visibility

	if len(ctx.AllStateMutability()) > 0 {
		decl.StateMutability = toText(ctx.AllStateMutability()[0])
	}
	return decl
}

func (e *exampleListener) VisitFunctionTypeParameter(ctx *solAntlr.FunctionTypeParameterContext) INode {
	decl := &VariableDeclaration{
		Node:     Node{Type: "VariableDeclaration"},
		TypeName: e.Visit(ctx.TypeName()),
	}
	if ctx.StorageLocation() != nil {
		decl.StorageLocation = toText(ctx.StorageLocation())
	}
	return decl
}

// UsingForDeclaration
type UsingForDeclaration struct {
	Node

	LibraryName string
	TypeName    interface{}
}

func (e *exampleListener) VisitUsingForDeclaration(ctx *solAntlr.UsingForDeclarationContext) INode {
	decl := &UsingForDeclaration{
		LibraryName: toText(ctx.Identifier()),
	}
	if typ := ctx.TypeName(); typ != nil {
		decl.TypeName = e.Visit(typ)
	}
	return decl
}

func (e *exampleListener) VisitTypeName(ctx *solAntlr.TypeNameContext) INode {
	if ctx.GetChildren() != nil && ctx.GetChildCount() > 2 {
		var length interface{}
		if ctx.GetChildCount() == 4 {
			expr := ctx.Expression()
			if expr == nil {
				panic("BAD")
			}
			length = e.Visit(expr)
		}
		decl := &ArrayTypeName{
			Node:         Node{Type: "ArrayTypeName"},
			BaseTypeName: e.Visit(ctx.TypeName()),
			Length:       length,
		}
		return decl
	}
	if ctx.GetChildren() != nil && ctx.GetChildCount() == 2 {
		decl := &ElementaryTypeName{
			Node:            Node{Type: "ElementaryTypeName"},
			Name:            toText(ctx.GetChild(0)),
			StateMutability: toText(ctx.GetChild(1)),
		}
		return decl
	}
	if elem := ctx.ElementaryTypeName(); elem != nil {
		return e.Visit(elem)
	}
	if elem := ctx.UserDefinedTypeName(); elem != nil {
		return e.Visit(elem)
	}
	if elem := ctx.Mapping(); elem != nil {
		return e.Visit(elem)
	}
	if elem := ctx.FunctionTypeName(); elem != nil {
		return e.Visit(elem)
	}
	panic("TODO")
}

type InlineAssemblyStatement struct {
	Body interface{}
}

func (e *exampleListener) VisitInlineAssemblyStatement(ctx *solAntlr.InlineAssemblyStatementContext) interface{} {
	decl := &InlineAssemblyStatement{
		Body: e.Visit(ctx.AssemblyBlock()),
	}
	return decl
}

type AssemblyBlock struct {
	Operations []interface{}
}

func (e *exampleListener) VisitAssemblyBlock(ctx *solAntlr.AssemblyBlockContext) interface{} {
	decl := &AssemblyBlock{
		Operations: []interface{}{},
	}
	for _, op := range ctx.AllAssemblyItem() {
		decl.Operations = append(decl.Operations, e.Visit(op))
	}
	return decl
}

func (e *exampleListener) VisitAssemblyItem(ctx *solAntlr.AssemblyItemContext) interface{} {
	return e.Visit(ctx.GetChild(0))
}

func (e *exampleListener) VisitAssemblyExpression(ctx *solAntlr.AssemblyExpressionContext) interface{} {
	return e.Visit(ctx.GetChild(0))
}

type AssemblyCall struct {
	Arguments []interface{}
}

func (e *exampleListener) VisitAssemblyCall(ctx *solAntlr.AssemblyCallContext) interface{} {
	decl := &AssemblyCall{
		Arguments: []interface{}{},
	}
	for _, i := range ctx.AllAssemblyExpression() {
		decl.Arguments = append(decl.Arguments, e.Visit(i))
	}
	return decl
}

type AssemblyLiteral struct {
}

func (e *exampleListener) VisitAssemblyLiteral(ctx *solAntlr.AssemblyLiteralContext) interface{} {
	decl := &AssemblyLiteral{}
	return decl
}

type AssemblySwitch struct {
	Expression interface{}
	Cases      []interface{}
}

func (e *exampleListener) VisitAssemblySwitch(ctx *solAntlr.AssemblySwitchContext) interface{} {
	decl := &AssemblySwitch{
		Expression: e.Visit(ctx.AssemblyExpression()),
		Cases:      []interface{}{},
	}
	for _, c := range ctx.AllAssemblyCase() {
		decl.Cases = append(decl.Cases, e.Visit(c))
	}
	return decl
}

type AssemblyCase struct {
	Block interface{}
}

func (e *exampleListener) VisitAssemblyCase(ctx *solAntlr.AssemblyCaseContext) interface{} {
	decl := &AssemblyCase{
		Block: e.Visit(ctx.AssemblyBlock()),
	}
	return decl
}

type AssemblyLocalDefinition struct {
	Expression interface{}
}

func (e *exampleListener) VisitAssemblyLocalDefinition(ctx *solAntlr.AssemblyLocalDefinitionContext) interface{} {
	decl := &AssemblyLocalDefinition{}
	if expr := ctx.AssemblyExpression(); expr != nil {
		decl.Expression = e.Visit(expr)
	}
	return decl
}

type AssemblyFunctionDefinition struct {
	Body interface{}
}

func (e *exampleListener) VisitAssemblyFunctionDefinition(ctx *solAntlr.AssemblyFunctionDefinitionContext) interface{} {
	decl := &AssemblyFunctionDefinition{
		Body: e.Visit(ctx.AssemblyBlock()),
	}
	return decl
}

type AssemblyAssignment struct {
	Expression interface{}
}

func (e *exampleListener) VisitAssemblyAssignment(ctx *solAntlr.AssemblyAssignmentContext) interface{} {
	decl := &AssemblyAssignment{
		Expression: e.Visit(ctx.AssemblyExpression()),
	}
	return decl
}

type AssemblyFor struct {
	Pre       interface{}
	Condition interface{}
	Post      interface{}
	Body      interface{}
}

func (e *exampleListener) VisitAssemblyFor(ctx *solAntlr.AssemblyForContext) interface{} {
	decl := &AssemblyFor{
		Pre:       e.Visit(ctx.GetChild(0)),
		Condition: e.Visit(ctx.GetChild(1)),
		Post:      e.Visit(ctx.GetChild(2)),
		Body:      e.Visit(ctx.GetChild(3)),
	}
	return decl
}

type AssemblyIf struct {
	Condition interface{}
	Body      interface{}
}

func (e *exampleListener) VisitAssemblyIf(ctx *solAntlr.AssemblyIfContext) interface{} {
	decl := &AssemblyIf{
		Condition: e.Visit(ctx.AssemblyExpression()),
		Body:      e.Visit(ctx.AssemblyBlock()),
	}
	return decl
}

type AssemblyMember struct {
	Expression interface{}
	MemberName interface{}
}

func (e *exampleListener) VisitAssemblyMember(ctx *solAntlr.AssemblyMemberContext) interface{} {
	decl := &AssemblyMember{
		Expression: e.Visit(ctx.Identifier(0)),
		MemberName: e.Visit(ctx.Identifier(1)),
	}
	return decl
}

type EmitStatement struct {
	Node

	EventCall interface{}
}

func (e *exampleListener) VisitEmitStatement(ctx *solAntlr.EmitStatementContext) INode {
	decl := &EmitStatement{
		EventCall: e.Visit(ctx.FunctionCall()),
	}
	return decl
}

type ThrowStatement struct {
	Node
}

func (e *exampleListener) VisitThrowStatement(ctx *solAntlr.ThrowStatementContext) INode {
	decl := &ThrowStatement{}
	return decl
}

type FunctionCall struct {
	Node

	Arguments   []interface{}
	Names       []interface{}
	Identifiers []interface{}
	Expression  interface{}
}

func (e *exampleListener) VisitFunctionCall(ctx *solAntlr.FunctionCallContext) INode {
	decl := &FunctionCall{
		Expression:  e.Visit(ctx.Expression()),
		Arguments:   []interface{}{},
		Identifiers: []interface{}{},
		Names:       []interface{}{},
	}

	ctxArgs := ctx.FunctionCallArguments().(*solAntlr.FunctionCallArgumentsContext)
	ctxArgsExpr := ctxArgs.ExpressionList()
	ctxArgsName := ctxArgs.NameValueList()

	if ctxArgsExpr != nil {
		for _, expr := range ctxArgsExpr.(*solAntlr.ExpressionListContext).AllExpression() {
			decl.Arguments = append(decl.Arguments, e.Visit(expr))
		}
	} else if ctxArgsName != nil {
		for _, raw := range ctxArgsName.(*solAntlr.NameValueListContext).AllNameValue() {
			nameValue := raw.(*solAntlr.NameValueContext)
			decl.Arguments = append(decl.Arguments, e.Visit(nameValue.Expression()))
			decl.Names = append(decl.Names, toText(nameValue.Identifier()))
			decl.Identifiers = append(decl.Identifiers, e.Visit(nameValue.Identifier()))
		}
	}

	return decl
}

type CustomErrorDefinition struct {
	Node

	Name       string
	Parameters []interface{}
}

func (e *exampleListener) VisitCustomErrorDefinition(ctx *solAntlr.CustomErrorDefinitionContext) INode {
	decl := &CustomErrorDefinition{
		Name:       toText(ctx.Identifier()),
		Parameters: e.VisitParameterList(ctx.ParameterList().(*solAntlr.ParameterListContext)),
	}
	return decl
}

// TryStatement
type TryStatement struct {
	Node

	Expression       interface{}
	CatchClause      []interface{}
	Body             interface{}
	ReturnParameters interface{}
}

func (e *exampleListener) VisitTryStatement(ctx *solAntlr.TryStatementContext) INode {
	catchClause := []interface{}{}
	for _, i := range ctx.AllCatchClause() {
		catchClause = append(catchClause, e.Visit(i))
	}

	decl := &TryStatement{
		CatchClause: catchClause,
		Expression:  e.Visit(ctx.Expression()),
		Body:        e.Visit(ctx.Block()),
	}

	if ret := ctx.ReturnParameters(); ret != nil {
		decl.ReturnParameters = e.VisitReturnParameters(ret.(*solAntlr.ReturnParametersContext))
	}
	return decl
}

type CatchClause struct {
	Node

	IsReasonType bool
	Kind         string
	Parameters   interface{}
	Body         interface{}
}

func (e *exampleListener) VisitCatchClause(ctx *solAntlr.CatchClauseContext) INode {

	kind := ""
	if iden := ctx.Identifier(); iden != nil {
		kind = toText(iden)
		if kind != "Error" && kind != "Panic" {
			panic("Expected 'Error' or 'Panic'")
		}
	}

	decl := &CatchClause{
		Body:         e.Visit(ctx.Block()),
		IsReasonType: kind == "Error",
		Kind:         kind,
	}
	if param := ctx.ParameterList(); param != nil {
		decl.Parameters = e.VisitParameterList(param.(*solAntlr.ParameterListContext))
	}
	return decl
}

func (e *exampleListener) VisitReturnParameters(ctx *solAntlr.ReturnParametersContext) []interface{} {
	return e.VisitParameterList(ctx.ParameterList().(*solAntlr.ParameterListContext))
}

func (e *exampleListener) VisitParameterList(ctx *solAntlr.ParameterListContext) []interface{} {
	var parameters []interface{}
	for _, p := range ctx.AllParameter() {
		parameters = append(parameters, e.Visit(p))
	}
	return parameters
}

type FileLevelConstant struct {
	Node

	Name            string
	InitialValue    interface{}
	TypeName        interface{}
	IsDeclaredConst bool
	IsImmutable     bool
}

func (e *exampleListener) VisitFileLevelConstant(ctx *solAntlr.FileLevelConstantContext) INode {
	decl := &FileLevelConstant{
		Name:            toText(ctx.Identifier()),
		TypeName:        e.Visit(ctx.TypeName()),
		InitialValue:    e.Visit(ctx.Expression()),
		IsDeclaredConst: true,
	}
	return decl
}

type UncheckedStatement struct {
	Node

	Block interface{}
}

func (e *exampleListener) VisitUncheckedStatement(ctx *solAntlr.UncheckedStatementContext) INode {
	decl := &UncheckedStatement{
		Block: e.Visit(ctx.Block()),
	}
	return decl
}

type TypeDefinition struct {
	Node

	Name       string
	Definition interface{}
}

func (e *exampleListener) VisitTypeDefinition(ctx *solAntlr.TypeDefinitionContext) INode {
	decl := &TypeDefinition{
		Name:       toText(ctx.Identifier()),
		Definition: e.Visit(ctx.ElementaryTypeName()),
	}
	return decl
}

type ElementaryTypeName struct {
	Node

	Name            string
	StateMutability string
}

func (e *exampleListener) VisitElementaryTypeName(ctx *solAntlr.ElementaryTypeNameContext) INode {
	decl := &ElementaryTypeName{
		Name: toText(ctx),
	}
	return decl
}

type UserDefinedTypeName struct {
	Node

	NamePath string
}

func (e *exampleListener) VisitUserDefinedTypeName(ctx *solAntlr.UserDefinedTypeNameContext) INode {
	decl := &UserDefinedTypeName{
		NamePath: toText(ctx),
	}
	return decl
}

type RevertStatement struct {
	Node

	RevertCall interface{}
}

func (e *exampleListener) VisitRevertStatement(ctx *solAntlr.RevertStatementContext) INode {
	decl := &RevertStatement{
		RevertCall: e.Visit(ctx.FunctionCall()),
	}
	return decl
}

type antlrToText interface {
	GetText() string
}

func toText(i interface{}) string {
	if obj, ok := i.(antlrToText); ok {
		return obj.GetText()
	}
	panic("BUG: toText expr not found " + reflect.TypeOf(i).String())
}

type Parser struct {
	Result INode
	Errors []*SyntaxError
}

func (p *Parser) Json() (string, error) {
	data, err := json.Marshal(p.Result)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func Parse(s string) *Parser {
	// Setup the input
	is := antlr.NewInputStream(s)

	// Create the Lexer
	lexer := solAntlr.NewSolidityLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// Create the Parser
	parserErrors := &CustomErrorListener{}
	p := solAntlr.NewSolidityParser(stream)
	p.BuildParseTrees = true
	p.AddErrorListener(parserErrors)

	tree := p.SourceUnit()
	lis := &exampleListener{}
	lis.init()

	result := lis.Visit(tree)

	pp := &Parser{
		Result: result,
		Errors: parserErrors.Errors,
	}
	return pp
}

type SyntaxError struct {
	line, column int
	msg          string
}

func (c *SyntaxError) Error() string {
	return c.msg
}

type CustomErrorListener struct {
	*antlr.DefaultErrorListener
	Errors []*SyntaxError
}

func (c *CustomErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	c.Errors = append(c.Errors, &SyntaxError{
		line:   line,
		column: column,
		msg:    msg,
	})
}
