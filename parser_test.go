package solcparser

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type parserCase []struct {
	code   interface{}
	result interface{}
}

func TestParser(t *testing.T) {
	cases := parserCase{
		{
			parseNode(t, "enum Hello { A, B, C }"),
			&EnumDefinition{
				Node: Node{Type: "EnumDefinition"},
				Name: "Hello",
				Members: []interface{}{
					&EnumValue{
						Node: Node{Type: "EnumValue"},
						Name: "A",
					},
					&EnumValue{
						Node: Node{Type: "EnumValue"},
						Name: "B",
					},
					&EnumValue{
						Node: Node{Type: "EnumValue"},
						Name: "C",
					},
				},
			},
		},
		{
			parseNode(t, "using Lib for uint;"),
			&UsingForDeclaration{
				Node:        Node{Type: "UsingForDeclaration"},
				LibraryName: "Lib",
				TypeName: &ElementaryTypeName{
					Node: Node{Type: "ElementaryTypeName"},
					Name: "uint",
				},
			},
		},
		{
			parseNode(t, "using Lib for *;"),
			&UsingForDeclaration{
				Node:        Node{Type: "UsingForDeclaration"},
				LibraryName: "Lib",
				TypeName:    nil,
			},
		},
		{
			parseNode(t, "using Lib for S;"),
			&UsingForDeclaration{
				Node:        Node{Type: "UsingForDeclaration"},
				LibraryName: "Lib",
				TypeName: &UserDefinedTypeName{
					Node:     Node{Type: "UserDefinedTypeName"},
					NamePath: "S",
				},
			},
		},
		{
			parseContract(t, "pragma experimental ABIEncoderV2;"),
			&PragmaDirective{
				Node:  Node{Type: "PragmaDirective"},
				Name:  "experimental",
				Value: "ABIEncoderV2",
			},
		},
		{
			parseContract(t, "pragma abicoder v1"),
			&PragmaDirective{
				Node:  Node{Type: "PragmaDirective"},
				Name:  "abicoder",
				Value: "v1",
			},
		},
		{
			parseContract(t, "pragma abicoder v2"),
			&PragmaDirective{
				Node:  Node{Type: "PragmaDirective"},
				Name:  "abicoder",
				Value: "v2",
			},
		},

		{
			parseContract(t, "contract test {}"),
			&ContractDefinition{
				Node:          Node{Type: "ContractDefinition"},
				Name:          "test",
				SubNodes:      []interface{}{},
				BaseContracts: []interface{}{},
				Kind:          "contract",
			},
		},
		{
			parseContract(t, "contract test is foo, bar {}"),
			&ContractDefinition{
				Node:     Node{Type: "ContractDefinition"},
				Name:     "test",
				SubNodes: []interface{}{},
				BaseContracts: []interface{}{
					&InheritanceSpecifier{
						Node: Node{Type: "InheritanceSpecifier"},
						BaseName: &UserDefinedTypeName{
							Node:     Node{Type: "UserDefinedTypeName"},
							NamePath: "foo",
						},
						Arguments: []interface{}{},
					},
					&InheritanceSpecifier{
						Node: Node{Type: "InheritanceSpecifier"},
						BaseName: &UserDefinedTypeName{
							Node:     Node{Type: "UserDefinedTypeName"},
							NamePath: "bar",
						},
						Arguments: []interface{}{},
					},
				},
				Kind: "contract",
			},
		},
		{
			parseContract(t, "library test {}"),
			&ContractDefinition{
				Node:          Node{Type: "ContractDefinition"},
				Name:          "test",
				SubNodes:      []interface{}{},
				BaseContracts: []interface{}{},
				Kind:          "library",
			},
		},
		{
			parseContract(t, "interface test {}"),
			&ContractDefinition{
				Node:          Node{Type: "ContractDefinition"},
				Name:          "test",
				SubNodes:      []interface{}{},
				BaseContracts: []interface{}{},
				Kind:          "interface",
			},
		},

		{
			parseStatement(t, "return;"),
			&ReturnStatement{
				Node:       Node{Type: "ReturnStatement"},
				Expression: nil,
			},
		},
		{
			parseStatement(t, "return 2;"),
			&ReturnStatement{
				Node: Node{Type: "ReturnStatement"},
				Expression: &NumberLiteral{
					Node:   Node{Type: "NumberLiteral"},
					Number: "2",
				},
			},
		},
		{
			parseStatement(t, "return ();"),
			&ReturnStatement{
				Node: Node{Type: "ReturnStatement"},
				Expression: &TupleExpression{
					Node:       Node{Type: "TupleExpression"},
					IsArray:    false,
					Components: []interface{}{},
				},
			},
		},

		{
			parseStatement(t, "while (true) {}"),
			&WhileStatement{
				Node: Node{Type: "WhileStatement"},
				Condition: &BooleanLiteral{
					Node:  Node{Type: "BooleanLiteral"},
					Value: true,
				},
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
			},
		},
		{
			parseStatement(t, "do {} while (true);"),
			&DoWhileStatement{
				Node: Node{Type: "DoWhileStatement"},
				Condition: &BooleanLiteral{
					Node:  Node{Type: "BooleanLiteral"},
					Value: true,
				},
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
			},
		},
		{
			parseStatement(t, "if (true) {}"),
			&IfStatement{
				Node: Node{Type: "IfStatement"},
				Condition: &BooleanLiteral{
					Node:  Node{Type: "BooleanLiteral"},
					Value: true,
				},
				TrueBody: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				FalseBody: nil,
			},
		},
		{
			parseStatement(t, "if (true) {} else {}"),
			&IfStatement{
				Node: Node{Type: "IfStatement"},
				Condition: &BooleanLiteral{
					Node:  Node{Type: "BooleanLiteral"},
					Value: true,
				},
				TrueBody: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				FalseBody: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
			},
		},

		// modifier
		{
			parseNode(t, "modifier onlyOwner {}"),
			&ModifierDefinition{
				Node: Node{Type: "ModifierDefinition"},
				Name: "onlyOwner",
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Override: []interface{}{},
			},
		},
		{
			parseNode(t, "modifier onlyOwner() {}"),
			&ModifierDefinition{
				Node: Node{Type: "ModifierDefinition"},
				Name: "onlyOwner",
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Override: []interface{}{},
			},
		},
		{
			parseNode(t, "modifier foo1() virtual {}"),
			&ModifierDefinition{
				Node: Node{Type: "ModifierDefinition"},
				Name: "foo1",
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				IsVirtual: true,
				Override:  []interface{}{},
			},
		},
		{
			parseNode(t, "modifier foo2() virtual;"),
			&ModifierDefinition{
				Node:      Node{Type: "ModifierDefinition"},
				Name:      "foo2",
				Body:      nil,
				IsVirtual: true,
				Override:  []interface{}{},
			},
		},
		{
			parseNode(t, "modifier foo3() override {}"),
			&ModifierDefinition{
				Node: Node{Type: "ModifierDefinition"},
				Name: "foo3",
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Override: []interface{}{},
			},
		},
		{
			parseNode(t, "modifier foo4() override(Base) {}"),
			&ModifierDefinition{
				Node: Node{Type: "ModifierDefinition"},
				Name: "foo4",
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Override: []interface{}{
					&UserDefinedTypeName{
						Node:     Node{Type: "UserDefinedTypeName"},
						NamePath: "Base",
					},
				},
			},
		},
		{
			parseNode(t, "modifier foo5() override(Base1, Base2) {}"),
			&ModifierDefinition{
				Node: Node{Type: "ModifierDefinition"},
				Name: "foo5",
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Override: []interface{}{
					&UserDefinedTypeName{
						Node:     Node{Type: "UserDefinedTypeName"},
						NamePath: "Base1",
					},
					&UserDefinedTypeName{
						Node:     Node{Type: "UserDefinedTypeName"},
						NamePath: "Base2",
					},
				},
			},
		},

		// unchecked blocks

		// revert
		{
			parseStatement(t, "revert MyCustomError();"),
			&RevertStatement{
				Node: Node{Type: "RevertStatement"},
				RevertCall: &FunctionCall{
					Node:        Node{Type: "FunctionCall"},
					Arguments:   []interface{}{},
					Names:       []interface{}{},
					Identifiers: []interface{}{},
					Expression: &Identifier{
						Node: Node{Type: "Identifier"},
						Name: "MyCustomError",
					},
				},
			},
		},
		{
			parseStatement(t, "revert MyCustomError(3);"),
			&RevertStatement{
				Node: Node{Type: "RevertStatement"},
				RevertCall: &FunctionCall{
					Node:        Node{Type: "FunctionCall"},
					Names:       []interface{}{},
					Identifiers: []interface{}{},
					Arguments: []interface{}{
						&NumberLiteral{
							Node:   Node{Type: "NumberLiteral"},
							Number: "3",
						},
					},
					Expression: &Identifier{
						Node: Node{Type: "Identifier"},
						Name: "MyCustomError",
					},
				},
			},
		},

		// import

		{
			parseContract(t, `import "./abc.sol";`),
			&ImportDirective{
				Node: Node{Type: "ImportDirective"},
				Path: "./abc.sol",
				PathLiteral: &StringLiteral{
					Node:      Node{Type: "StringLiteral"},
					Value:     "./abc.sol",
					Parts:     []string{"./abc.sol"},
					IsUnicode: []bool{false},
				},
			},
		},
		{
			parseContract(t, `import "./abc.sol" as x;`),
			&ImportDirective{
				Node: Node{Type: "ImportDirective"},
				Path: "./abc.sol",
				PathLiteral: &StringLiteral{
					Node:      Node{Type: "StringLiteral"},
					Value:     "./abc.sol",
					Parts:     []string{"./abc.sol"},
					IsUnicode: []bool{false},
				},
				UnitAlias: "x",
				UnitAliasIdentifier: &Identifier{
					Node: Node{Type: "Identifier"},
					Name: "x",
				},
			},
		},
		{
			parseContract(t, `import * as x from "./abc.sol";`),
			&ImportDirective{
				Node: Node{Type: "ImportDirective"},
				Path: "./abc.sol",
				PathLiteral: &StringLiteral{
					Node:      Node{Type: "StringLiteral"},
					Value:     "./abc.sol",
					Parts:     []string{"./abc.sol"},
					IsUnicode: []bool{false},
				},
				UnitAlias: "x",
				UnitAliasIdentifier: &Identifier{
					Node: Node{Type: "Identifier"},
					Name: "x",
				},
			},
		},
		{
			parseContract(t, `import { a as b, c as d, f } from "./abc.sol";`),
			&ImportDirective{
				Node: Node{Type: "ImportDirective"},
				Path: "./abc.sol",
				PathLiteral: &StringLiteral{
					Node:      Node{Type: "StringLiteral"},
					Value:     "./abc.sol",
					Parts:     []string{"./abc.sol"},
					IsUnicode: []bool{false},
				},
				SymbolAliases: [][]string{
					{"a", "b"},
					{"c", "d"},
					{"f", ""},
				},
				SymbolAliasesIdentifiers: [][]*Identifier{
					{
						{
							Node: Node{Type: "Identifier"},
							Name: "a",
						},
						{
							Node: Node{Type: "Identifier"},
							Name: "b",
						},
					},
					{
						{
							Node: Node{Type: "Identifier"},
							Name: "c",
						},
						{
							Node: Node{Type: "Identifier"},
							Name: "d",
						},
					},
					{
						{
							Node: Node{Type: "Identifier"},
							Name: "f",
						},
						nil,
					},
				},
			},
		},

		// expression

		{
			parseExpression(t, "new MyContract"),
			&NewExpression{
				Node: Node{Type: "NewExpression"},
				TypeName: &UserDefinedTypeName{
					Node:     Node{Type: "UserDefinedTypeName"},
					NamePath: "MyContract",
				},
			},
		},
		{
			parseExpression(t, "!true"),
			&UnaryOperation{
				Node:     Node{Type: "UnaryOperation"},
				Operator: "!",
				SubExpression: &BooleanLiteral{
					Node:  Node{Type: "BooleanLiteral"},
					Value: true,
				},
				IsPrefix: true,
			},
		},
		{
			parseExpression(t, "i++"),
			&UnaryOperation{
				Node:     Node{Type: "UnaryOperation"},
				Operator: "++",
				SubExpression: &Identifier{
					Node: Node{Type: "Identifier"},
					Name: "i",
				},
				IsPrefix: false,
			},
		},
		{
			parseExpression(t, "--i"),
			&UnaryOperation{
				Node:     Node{Type: "UnaryOperation"},
				Operator: "--",
				SubExpression: &Identifier{
					Node: Node{Type: "Identifier"},
					Name: "i",
				},
				IsPrefix: true,
			},
		},

		// for

		{
			parseStatement(t, "for (i = 0; i < 10; i++) {}"),
			&ForStatement{
				Node: Node{Type: "ForStatement"},
				InitExpression: &ExpressionStatement{
					Node: Node{Type: "ExpressionStatement"},
					Expression: &BinaryOperation{
						Node:     Node{Type: "BinaryOperation"},
						Operator: "=",
						Left: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "i",
						},
						Right: &NumberLiteral{
							Node:            Node{Type: "NumberLiteral"},
							Number:          "0",
							SubDenomination: nil,
						},
					},
				},
				ConditionExpression: &BinaryOperation{
					Node:     Node{Type: "BinaryOperation"},
					Operator: "<",
					Left: &Identifier{
						Node: Node{Type: "Identifier"},
						Name: "i",
					},
					Right: &NumberLiteral{
						Node:            Node{Type: "NumberLiteral"},
						Number:          "10",
						SubDenomination: nil,
					},
				},
				LoopExpression: &ExpressionStatement{
					Node: Node{Type: "ExpressionStatement"},
					Expression: &UnaryOperation{
						Node:     Node{Type: "UnaryOperation"},
						Operator: "++",
						SubExpression: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "i",
						},
						IsPrefix: false,
					},
				},
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
			},
		},
		{
			parseStatement(t, "for (;; i++) {}"),
			&ForStatement{
				Node:                Node{Type: "ForStatement"},
				InitExpression:      nil,
				ConditionExpression: nil,
				LoopExpression: &ExpressionStatement{
					Node: Node{Type: "ExpressionStatement"},
					Expression: &UnaryOperation{
						Node:     Node{Type: "UnaryOperation"},
						Operator: "++",
						SubExpression: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "i",
						},
						IsPrefix: false,
					},
				},
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
			},
		},

		// function calls

		{
			parseNode(t, "constructor(uint a) public {}"),
			&FunctionDefinition{
				Node: Node{Type: "FunctionDefinition"},
				Name: "",
				Parameters: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						Name: "a",
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint",
						},
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "a",
						},
					},
				},
				ReturnParameters: nil,
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Visibility:    "public",
				Modifiers:     []interface{}{},
				IsConstructor: true,
			},
		},

		{
			parseNode(t, "constructor(uint a) {}"),
			&FunctionDefinition{
				Node: Node{Type: "FunctionDefinition"},
				Name: "",
				Parameters: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						Name: "a",
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint",
						},
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "a",
						},
					},
				},
				ReturnParameters: nil,
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Visibility:    "default",
				Modifiers:     []interface{}{},
				IsConstructor: true,
			},
		},

		{
			parseNode(t, "fallback () external {}"),
			&FunctionDefinition{
				Node:       Node{Type: "FunctionDefinition"},
				Parameters: nil,
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				IsFallback: true,
				Modifiers:  []interface{}{},
				Visibility: "external",
			},
		},

		{
			parseNode(t, "fallback () external payable virtual {}"),
			&FunctionDefinition{
				Node:       Node{Type: "FunctionDefinition"},
				Parameters: nil,
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				IsFallback: true,
				Modifiers:  []interface{}{},
				IsVirtual:  true,
				Visibility: "external",
			},
		},

		{
			parseNode(t, "receive () external payable virtual {}"),
			&FunctionDefinition{
				Node:       Node{Type: "FunctionDefinition"},
				Parameters: nil,
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				IsReceiveEther: true,
				Modifiers:      []interface{}{},
				IsVirtual:      true,
				Visibility:     "external",
			},
		},

		{
			parseNode(t, "function () external {}"),
			&FunctionDefinition{
				Node:       Node{Type: "FunctionDefinition"},
				Parameters: nil,
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Modifiers:  []interface{}{},
				IsFallback: true,
				Visibility: "external",
			},
		},

		{
			parseNode(t, "receive () external payable {}"),
			&FunctionDefinition{
				Node:       Node{Type: "FunctionDefinition"},
				Parameters: nil,
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Modifiers:      []interface{}{},
				IsReceiveEther: true,
				Visibility:     "external",
			},
		},

		{
			parseNode(t, "function foo() public override {}"),
			&FunctionDefinition{
				Node:       Node{Type: "FunctionDefinition"},
				Name:       "foo",
				Parameters: nil,
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Modifiers:  []interface{}{},
				Visibility: "public",
			},
		},

		{
			parseNode(t, "function foo() public override(Base) {}"),
			&FunctionDefinition{
				Node:       Node{Type: "FunctionDefinition"},
				Name:       "foo",
				Parameters: nil,
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Visibility: "public",
				Modifiers:  []interface{}{},
				Override: []interface{}{
					&UserDefinedTypeName{
						Node:     Node{Type: "UserDefinedTypeName"},
						NamePath: "Base",
					},
				},
			},
		},

		{
			parseNode(t, "function foo() public override(Base1, Base2) {}"),
			&FunctionDefinition{
				Node:       Node{Type: "FunctionDefinition"},
				Name:       "foo",
				Parameters: nil,
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Visibility: "public",
				Modifiers:  []interface{}{},
				Override: []interface{}{
					&UserDefinedTypeName{
						Node:     Node{Type: "UserDefinedTypeName"},
						NamePath: "Base1",
					},
					&UserDefinedTypeName{
						Node:     Node{Type: "UserDefinedTypeName"},
						NamePath: "Base2",
					},
				},
			},
		},

		{
			parseNode(t, "function foo(uint a) pure {}"),
			&FunctionDefinition{
				Node: Node{Type: "FunctionDefinition"},
				Name: "foo",
				Parameters: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						Name: "a",
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint",
						},
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "a",
						},
					},
				},
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Visibility: "default",
				Modifiers:  []interface{}{},
			},
		},

		{
			parseNode(t, "function foo() virtual public {}"),
			&FunctionDefinition{
				Node:       Node{Type: "FunctionDefinition"},
				Name:       "foo",
				Parameters: nil,
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Visibility: "public",
				Modifiers:  []interface{}{},
				IsVirtual:  true,
			},
		},

		{
			parseNode(t, "function foo(uint a) pure returns (uint256) {}"),
			&FunctionDefinition{
				Node: Node{Type: "FunctionDefinition"},
				Name: "foo",
				Parameters: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						Name: "a",
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint",
						},
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "a",
						},
					},
				},
				ReturnParameters: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint256",
						},
						Identifier: nil,
					},
				},
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Visibility: "default",
				Modifiers:  []interface{}{},
			},
		},

		/*
			{
				parseNode(t, "function foo(uint a) onlyOwner {}"),
			},
			{
				parseNode(t, "function foo(uint a) onlyOwner() {}"),
			},
			{
				parseNode(t, "function foo(uint a) bar(true, 1) {}"),
			},
		*/

		/*
			{
				parseStatement(t, "uint(a);"),
				&TypeNameExpression{
					TypeName: &ElementaryTypeName{
						Node: Node{Type: "ElementaryTypeName"},
						Name: "uint",
					},
				},
			},
		*/

		{
			parseStatement(t, "try f(1, 2) returns (uint a) {} catch (bytes memory a) {}"),
			&TryStatement{
				Node: Node{Type: "TryStatement"},
				Expression: &FunctionCall{
					Node: Node{Type: "FunctionCall"},
					Expression: &Identifier{
						Node: Node{Type: "Identifier"},
						Name: "f",
					},
					Arguments: []interface{}{
						&NumberLiteral{
							Node:   Node{Type: "NumberLiteral"},
							Number: "1",
						},
						&NumberLiteral{
							Node:   Node{Type: "NumberLiteral"},
							Number: "2",
						},
					},
				},
				ReturnParameters: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						Name: "a",
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "a",
						},
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint",
						},
					},
				},
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				CatchClause: []interface{}{
					&CatchClause{
						Node: Node{Type: "CatchClause"},
						Body: &Block{
							Node:       Node{Type: "Block"},
							Statements: []interface{}{},
						},
						Parameters: []interface{}{
							&VariableDeclaration{
								Node: Node{Type: "VariableDeclaration"},
								Name: "a",
								Identifier: &Identifier{
									Node: Node{Type: "Identifier"},
									Name: "a",
								},
								StorageLocation: "memory",
								TypeName: &ElementaryTypeName{
									Node: Node{Type: "ElementaryTypeName"},
									Name: "bytes",
								},
							},
						},
					},
				},
			},
		},

		{
			parseStatement(t, "try f(1, 2) returns (uint a) {} catch Error(string memory b) {} catch (bytes memory c) {}"),
			&TryStatement{
				Node: Node{Type: "TryStatement"},
				Expression: &FunctionCall{
					Node: Node{Type: "FunctionCall"},
					Expression: &Identifier{
						Node: Node{Type: "Identifier"},
						Name: "f",
					},
					Arguments: []interface{}{
						&NumberLiteral{
							Node:   Node{Type: "NumberLiteral"},
							Number: "1",
						},
						&NumberLiteral{
							Node:   Node{Type: "NumberLiteral"},
							Number: "2",
						},
					},
				},
				ReturnParameters: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						Name: "a",
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "a",
						},
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint",
						},
					},
				},
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				CatchClause: []interface{}{
					&CatchClause{
						Node: Node{Type: "CatchClause"},
						Kind: "Error",
						Body: &Block{
							Node:       Node{Type: "Block"},
							Statements: []interface{}{},
						},
						IsReasonType: true,
						Parameters: []interface{}{
							&VariableDeclaration{
								Node: Node{Type: "VariableDeclaration"},
								Name: "b",
								Identifier: &Identifier{
									Node: Node{Type: "Identifier"},
									Name: "b",
								},
								StorageLocation: "memory",
								TypeName: &ElementaryTypeName{
									Node: Node{Type: "ElementaryTypeName"},
									Name: "string",
								},
							},
						},
					},
					&CatchClause{
						Node: Node{Type: "CatchClause"},
						Kind: "",
						Body: &Block{
							Node:       Node{Type: "Block"},
							Statements: []interface{}{},
						},
						Parameters: []interface{}{
							&VariableDeclaration{
								Node: Node{Type: "VariableDeclaration"},
								Name: "c",
								Identifier: &Identifier{
									Node: Node{Type: "Identifier"},
									Name: "c",
								},
								StorageLocation: "memory",
								TypeName: &ElementaryTypeName{
									Node: Node{Type: "ElementaryTypeName"},
									Name: "bytes",
								},
							},
						},
					},
				},
			},
		},

		{
			parseStatement(t, "try f(1, 2) returns (uint a) {} catch Panic(uint errorCode) {} catch (bytes memory c) {}"),
			&TryStatement{
				Node: Node{Type: "TryStatement"},
				Expression: &FunctionCall{
					Node: Node{Type: "FunctionCall"},
					Expression: &Identifier{
						Node: Node{Type: "Identifier"},
						Name: "f",
					},
					Arguments: []interface{}{
						&NumberLiteral{
							Node:   Node{Type: "NumberLiteral"},
							Number: "1",
						},
						&NumberLiteral{
							Node:   Node{Type: "NumberLiteral"},
							Number: "2",
						},
					},
				},
				ReturnParameters: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint",
						},
						Name: "a",
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "a",
						},
					},
				},
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				CatchClause: []interface{}{
					&CatchClause{
						Node: Node{Type: "CatchClause"},
						Kind: "Panic",
						Body: &Block{
							Node:       Node{Type: "Block"},
							Statements: []interface{}{},
						},
						IsReasonType: false,
						Parameters: []interface{}{
							&VariableDeclaration{
								Node: Node{Type: "VariableDeclaration"},
								Name: "errorCode",
								Identifier: &Identifier{
									Node: Node{Type: "Identifier"},
									Name: "errorCode",
								},
								TypeName: &ElementaryTypeName{
									Node: Node{Type: "ElementaryTypeName"},
									Name: "uint",
								},
							},
						},
					},
					&CatchClause{
						Node: Node{Type: "CatchClause"},
						Body: &Block{
							Node:       Node{Type: "Block"},
							Statements: []interface{}{},
						},
						Parameters: []interface{}{
							&VariableDeclaration{
								Node: Node{Type: "VariableDeclaration"},
								Name: "c",
								Identifier: &Identifier{
									Node: Node{Type: "Identifier"},
									Name: "c",
								},
								StorageLocation: "memory",
								TypeName: &ElementaryTypeName{
									Node: Node{Type: "ElementaryTypeName"},
									Name: "bytes",
								},
							},
						},
					},
				},
			},
		},

		{
			parseStatement(t, "try f(1, 2) returns (uint a) {} catch Error(string memory b) {} catch Panic(uint errorCode) {} catch (bytes memory c) {}"),
			&TryStatement{
				Node: Node{Type: "TryStatement"},
				Expression: &FunctionCall{
					Node: Node{Type: "FunctionCall"},
					Expression: &Identifier{
						Node: Node{Type: "Identifier"},
						Name: "f",
					},
					Arguments: []interface{}{
						&NumberLiteral{
							Node:   Node{Type: "NumberLiteral"},
							Number: "1",
						},
						&NumberLiteral{
							Node:   Node{Type: "NumberLiteral"},
							Number: "2",
						},
					},
				},
				ReturnParameters: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint",
						},
						Name: "a",
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "a",
						},
					},
				},
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				CatchClause: []interface{}{
					&CatchClause{
						Node: Node{Type: "CatchClause"},
						Kind: "Error",
						Body: &Block{
							Node:       Node{Type: "Block"},
							Statements: []interface{}{},
						},
						IsReasonType: true,
						Parameters: []interface{}{
							&VariableDeclaration{
								Node: Node{Type: "VariableDeclaration"},
								Name: "b",
								Identifier: &Identifier{
									Node: Node{Type: "Identifier"},
									Name: "b",
								},
								StorageLocation: "memory",
								TypeName: &ElementaryTypeName{
									Node: Node{Type: "ElementaryTypeName"},
									Name: "string",
								},
							},
						},
					},
					&CatchClause{
						Node: Node{Type: "CatchClause"},
						Kind: "Panic",
						Body: &Block{
							Node:       Node{Type: "Block"},
							Statements: []interface{}{},
						},
						IsReasonType: false,
						Parameters: []interface{}{
							&VariableDeclaration{
								Node: Node{Type: "VariableDeclaration"},
								Name: "errorCode",
								Identifier: &Identifier{
									Node: Node{Type: "Identifier"},
									Name: "errorCode",
								},
								TypeName: &ElementaryTypeName{
									Node: Node{Type: "ElementaryTypeName"},
									Name: "uint",
								},
							},
						},
					},
					&CatchClause{
						Node: Node{Type: "CatchClause"},
						Body: &Block{
							Node:       Node{Type: "Block"},
							Statements: []interface{}{},
						},
						Parameters: []interface{}{
							&VariableDeclaration{
								Node: Node{Type: "VariableDeclaration"},
								Name: "c",
								Identifier: &Identifier{
									Node: Node{Type: "Identifier"},
									Name: "c",
								},
								StorageLocation: "memory",
								TypeName: &ElementaryTypeName{
									Node: Node{Type: "ElementaryTypeName"},
									Name: "bytes",
								},
							},
						},
					},
				},
			},
		},

		// expressions

		{
			parseExpression(t, "a[:]"),
			&IndexRangeAccess{
				Node: Node{Type: "IndexRangeAccess"},
				Base: &Identifier{
					Node: Node{Type: "Identifier"},
					Name: "a",
				},
			},
		},

		{
			parseExpression(t, "a[3:]"),
			&IndexRangeAccess{
				Node: Node{Type: "IndexRangeAccess"},
				Base: &Identifier{
					Node: Node{Type: "Identifier"},
					Name: "a",
				},
				IndexStart: &NumberLiteral{
					Node:   Node{Type: "NumberLiteral"},
					Number: "3",
				},
			},
		},

		{
			parseExpression(t, "a[:20]"),
			&IndexRangeAccess{
				Node: Node{Type: "IndexRangeAccess"},
				Base: &Identifier{
					Node: Node{Type: "Identifier"},
					Name: "a",
				},
				IndexEnd: &NumberLiteral{
					Node:   Node{Type: "NumberLiteral"},
					Number: "20",
				},
			},
		},

		{
			parseExpression(t, "a[0:4]"),
			&IndexRangeAccess{
				Node: Node{Type: "IndexRangeAccess"},
				Base: &Identifier{
					Node: Node{Type: "Identifier"},
					Name: "a",
				},
				IndexStart: &NumberLiteral{
					Node:   Node{Type: "NumberLiteral"},
					Number: "0",
				},
				IndexEnd: &NumberLiteral{
					Node:   Node{Type: "NumberLiteral"},
					Number: "4",
				},
			},
		},

		// units

		{
			parseStatement(t, "a = 1 wei;"),
			&ExpressionStatement{
				Node: Node{Type: "ExpressionStatement"},
				Expression: &BinaryOperation{
					Node: Node{Type: "BinaryOperation"},
					Left: &Identifier{
						Node: Node{Type: "Identifier"},
						Name: "a",
					},
					Operator: "=",
					Right: &NumberLiteral{
						Node:            Node{Type: "NumberLiteral"},
						Number:          "1",
						SubDenomination: "wei",
					},
				},
			},
		},
		{
			parseStatement(t, "a = 1 gwei;"),
			&ExpressionStatement{
				Node: Node{Type: "ExpressionStatement"},
				Expression: &BinaryOperation{
					Node: Node{Type: "BinaryOperation"},
					Left: &Identifier{
						Node: Node{Type: "Identifier"},
						Name: "a",
					},
					Operator: "=",
					Right: &NumberLiteral{
						Node:            Node{Type: "NumberLiteral"},
						Number:          "1",
						SubDenomination: "gwei",
					},
				},
			},
		},
		{
			parseStatement(t, "a = 1 seconds;"),
			&ExpressionStatement{
				Node: Node{Type: "ExpressionStatement"},
				Expression: &BinaryOperation{
					Node: Node{Type: "BinaryOperation"},
					Left: &Identifier{
						Node: Node{Type: "Identifier"},
						Name: "a",
					},
					Operator: "=",
					Right: &NumberLiteral{
						Node:            Node{Type: "NumberLiteral"},
						Number:          "1",
						SubDenomination: "seconds",
					},
				},
			},
		},

		// event definition

		{
			parseNode(t, "event Foo(address indexed a, uint b);"),
			&EventDefinition{
				Node: Node{Type: "EventDefinition"},
				Name: "Foo",
				Parameters: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						Name: "a",
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "address",
						},
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "a",
						},
						IsIndexed: true,
					},
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						Name: "b",
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint",
						},
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "b",
						},
					},
				},
			},
		},

		// throw

		{
			parseStatement(t, "throw;"),
			&ThrowStatement{
				Node: Node{Type: "ThrowStatement"},
			},
		},

		// variable declarations

		{
			parseNode(t, "uint a;"),
			&StateVariableDeclaration{
				Node: Node{Type: "StateVariableDeclaration"},
				Variables: []interface{}{
					&StateVariableDeclarationVariable{
						VariableDeclaration: VariableDeclaration{
							Node: Node{Type: "VariableDeclaration"},
							TypeName: &ElementaryTypeName{
								Node: Node{Type: "ElementaryTypeName"},
								Name: "uint",
							},
							Name: "a",
							Identifier: &Identifier{
								Node: Node{Type: "Identifier"},
								Name: "a",
							},
							Visibility: "default",
							IsStateVar: true,
						},
					},
				},
			},
		},

		{
			parseNode(t, "uint immutable foo;"),
			&StateVariableDeclaration{
				Node: Node{Type: "StateVariableDeclaration"},
				Variables: []interface{}{
					&StateVariableDeclarationVariable{
						VariableDeclaration: VariableDeclaration{
							Node: Node{Type: "VariableDeclaration"},
							TypeName: &ElementaryTypeName{
								Node: Node{Type: "ElementaryTypeName"},
								Name: "uint",
							},
							Name: "foo",
							Identifier: &Identifier{
								Node: Node{Type: "Identifier"},
								Name: "foo",
							},
							Visibility: "default",
							IsStateVar: true,
						},
						IsInmutable: true,
					},
				},
			},
		},

		{
			parseNode(t, "uint public override(Base) foo;"),
			&StateVariableDeclaration{
				Node: Node{Type: "StateVariableDeclaration"},
				Variables: []interface{}{
					&StateVariableDeclarationVariable{
						VariableDeclaration: VariableDeclaration{
							Node: Node{Type: "VariableDeclaration"},
							TypeName: &ElementaryTypeName{
								Node: Node{Type: "ElementaryTypeName"},
								Name: "uint",
							},
							Name: "foo",
							Identifier: &Identifier{
								Node: Node{Type: "Identifier"},
								Name: "foo",
							},
							Visibility: "public",
							Override: []interface{}{
								&UserDefinedTypeName{
									Node:     Node{Type: "UserDefinedTypeName"},
									NamePath: "Base",
								},
							},
							IsStateVar: true,
						},
					},
				},
			},
		},

		{
			parseNode(t, "uint public override(Base1, Base2) foo;"),
			&StateVariableDeclaration{
				Node: Node{Type: "StateVariableDeclaration"},
				Variables: []interface{}{
					&StateVariableDeclarationVariable{
						VariableDeclaration: VariableDeclaration{
							Node: Node{Type: "VariableDeclaration"},
							TypeName: &ElementaryTypeName{
								Node: Node{Type: "ElementaryTypeName"},
								Name: "uint",
							},
							Name: "foo",
							Identifier: &Identifier{
								Node: Node{Type: "Identifier"},
								Name: "foo",
							},
							Visibility: "public",
							Override: []interface{}{
								&UserDefinedTypeName{
									Node:     Node{Type: "UserDefinedTypeName"},
									NamePath: "Base1",
								},
								&UserDefinedTypeName{
									Node:     Node{Type: "UserDefinedTypeName"},
									NamePath: "Base2",
								},
							},
							IsStateVar: true,
						},
					},
				},
			},
		},

		{
			parse(t, "uint constant EXPONENT = 10;").(*SourceUnit).Children[0],
			&FileLevelConstant{
				Node: Node{Type: "FileLevelConstant"},
				InitialValue: &NumberLiteral{
					Node:   Node{Type: "NumberLiteral"},
					Number: "10",
				},
				Name: "EXPONENT",
				TypeName: &ElementaryTypeName{
					Node: Node{Type: "ElementaryTypeName"},
					Name: "uint",
				},
				IsDeclaredConst: true,
			},
		},

		{
			parseStatement(t, "var (a,,b) = 0;"),
			&VariableDeclarationStatement{
				Node: Node{Type: "VariableDeclarationStatement"},
				InitialValue: &NumberLiteral{
					Node:   Node{Type: "NumberLiteral"},
					Number: "0",
				},
				Variables: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						Name: "a",
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "a",
						},
					},
					nil,
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						Name: "b",
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "b",
						},
					},
				},
			},
		},

		{
			parseStatement(t, "(uint a,, uint b) = 0;"),
			&VariableDeclarationStatement{
				Node: Node{Type: "VariableDeclarationStatement"},
				InitialValue: &NumberLiteral{
					Node:   Node{Type: "NumberLiteral"},
					Number: "0",
				},
				Variables: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						Name: "a",
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "a",
						},
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint",
						},
					},
					nil,
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						Name: "b",
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "b",
						},
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint",
						},
					},
				},
			},
		},

		// expression

		{
			parseExpression(t, "(a,) = (1,2)").(*BinaryOperation).Left,
			&TupleExpression{
				Node: Node{Type: "TupleExpression"},
				Components: []interface{}{
					&Identifier{
						Node: Node{Type: "Identifier"},
						Name: "a",
					},
					nil,
				},
			},
		},

		{
			parseExpression(t, "(a) = (1,)").(*BinaryOperation).Left,
			&TupleExpression{
				Node: Node{Type: "TupleExpression"},
				Components: []interface{}{
					&Identifier{
						Node: Node{Type: "Identifier"},
						Name: "a",
					},
				},
			},
		},

		{
			parseExpression(t, "(a,,b,) = (1,2,1)").(*BinaryOperation).Left,
			&TupleExpression{
				Node: Node{Type: "TupleExpression"},
				Components: []interface{}{
					&Identifier{
						Node: Node{Type: "Identifier"},
						Name: "a",
					},
					nil,
					&Identifier{
						Node: Node{Type: "Identifier"},
						Name: "b",
					},
					nil,
				},
			},
		},

		{
			parseExpression(t, "a"),
			&Identifier{
				Node: Node{Type: "Identifier"},
				Name: "a",
			},
		},
		{
			parseExpression(t, "calldata"),
			&Identifier{
				Node: Node{Type: "Identifier"},
				Name: "calldata",
			},
		},

		{
			parseExpression(t, "(,a,, b,,)"),
			&TupleExpression{
				Node: Node{Type: "TupleExpression"},
				Components: []interface{}{
					nil,
					&Identifier{
						Node: Node{Type: "Identifier"},
						Name: "a",
					},
					nil,
					&Identifier{
						Node: Node{Type: "Identifier"},
						Name: "b",
					},
					nil,
					nil,
				},
			},
		},

		{
			parseExpression(t, "[a, b]"),
			&TupleExpression{
				Node: Node{Type: "TupleExpression"},
				Components: []interface{}{
					&Identifier{
						Node: Node{Type: "Identifier"},
						Name: "a",
					},
					&Identifier{
						Node: Node{Type: "Identifier"},
						Name: "b",
					},
				},
				IsArray: true,
			},
		},

		// statements

		{
			parseNode(t, "uint256[2] a;").(*StateVariableDeclaration).Variables[0].(*StateVariableDeclarationVariable).TypeName,
			&ArrayTypeName{
				Node: Node{Type: "ArrayTypeName"},
				BaseTypeName: &ElementaryTypeName{
					Node: Node{Type: "ElementaryTypeName"},
					Name: "uint256",
				},
				Length: &NumberLiteral{
					Node:   Node{Type: "NumberLiteral"},
					Number: "2",
				},
			},
		},

		{
			parseNode(t, "uint256[] a;").(*StateVariableDeclaration).Variables[0].(*StateVariableDeclarationVariable).TypeName,
			&ArrayTypeName{
				Node: Node{Type: "ArrayTypeName"},
				BaseTypeName: &ElementaryTypeName{
					Node: Node{Type: "ElementaryTypeName"},
					Name: "uint256",
				},
			},
		},

		// A[]
		{
			parseNode(t, "A[] a;").(*StateVariableDeclaration).Variables[0].(*StateVariableDeclarationVariable).TypeName,
			&ArrayTypeName{
				Node: Node{Type: "ArrayTypeName"},
				BaseTypeName: &UserDefinedTypeName{
					Node:     Node{Type: "UserDefinedTypeName"},
					NamePath: "A",
				},
			},
		},

		{
			parseNode(t, "address payable a;").(*StateVariableDeclaration).Variables[0].(*StateVariableDeclarationVariable).TypeName,
			&ElementaryTypeName{
				Node:            Node{Type: "ElementaryTypeName"},
				Name:            "address",
				StateMutability: "payable",
			},
		},

		{
			parseNode(t, "Foo.Bar a;").(*StateVariableDeclaration).Variables[0].(*StateVariableDeclarationVariable).TypeName,
			&UserDefinedTypeName{
				Node:     Node{Type: "UserDefinedTypeName"},
				NamePath: "Foo.Bar",
			},
		},

		{
			parseStatement(t, "true;"),
			&ExpressionStatement{
				Node: Node{Type: "ExpressionStatement"},
				Expression: &BooleanLiteral{
					Node:  Node{Type: "BooleanLiteral"},
					Value: true,
				},
			},
		},

		{
			parseNode(t, "function (uint, uint) returns(bool) a;").(*StateVariableDeclaration).Variables[0].(*StateVariableDeclarationVariable).TypeName,
			&FunctionTypeName{
				Node: Node{Type: "FunctionTypeName"},
				ParameterTypes: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint",
						},
					},
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint",
						},
					},
				},
				ReturnTypes: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "bool",
						},
					},
				},
				Visibility: "default",
			},
		},

		{
			parseStatement(t, "emit EventCalled(1);"),
			&EmitStatement{
				Node: Node{Type: "EmitStatement"},
				EventCall: &FunctionCall{
					Node: Node{Type: "FunctionCall"},
					Arguments: []interface{}{
						&NumberLiteral{
							Node:   Node{Type: "NumberLiteral"},
							Number: "1",
						},
					},
					Expression: &Identifier{
						Node: Node{Type: "Identifier"},
						Name: "EventCalled",
					},
					Names:       []interface{}{},
					Identifiers: []interface{}{},
				},
			},
		},
		{
			parseStatement(t, "emit EventCalled({x : 1});"),
			&EmitStatement{
				Node: Node{Type: "EmitStatement"},
				EventCall: &FunctionCall{
					Node: Node{Type: "FunctionCall"},
					Arguments: []interface{}{
						&NumberLiteral{
							Node:   Node{Type: "NumberLiteral"},
							Number: "1",
						},
					},
					Names: []interface{}{"x"},
					Identifiers: []interface{}{
						&Identifier{
							Node: Node{Type: "Identifier"},
							Name: "x",
						},
					},
					Expression: &Identifier{
						Node: Node{Type: "Identifier"},
						Name: "EventCalled",
					},
				},
			},
		},

		{
			parseNode(t, "modifier foo() virtual {}"),
			&ModifierDefinition{
				Node: Node{Type: "ModifierDefinition"},
				Name: "foo",
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				IsVirtual: true,
				Override:  []interface{}{},
			},
		},

		{
			parseNode(t, "modifier foo() virtual;"),
			&ModifierDefinition{
				Node:      Node{Type: "ModifierDefinition"},
				Name:      "foo",
				Body:      nil,
				IsVirtual: true,
				Override:  []interface{}{},
			},
		},

		{
			parseNode(t, "modifier foo() override {}"),
			&ModifierDefinition{
				Node: Node{Type: "ModifierDefinition"},
				Name: "foo",
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Override: []interface{}{},
			},
		},

		{
			parseNode(t, "modifier foo() override(Base) {}"),
			&ModifierDefinition{
				Node: Node{Type: "ModifierDefinition"},
				Name: "foo",
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Override: []interface{}{
					&UserDefinedTypeName{
						Node:     Node{Type: "UserDefinedTypeName"},
						NamePath: "Base",
					},
				},
			},
		},
		{
			parseNode(t, "modifier foo() override(Base1, Base2) {}"),
			&ModifierDefinition{
				Node: Node{Type: "ModifierDefinition"},
				Name: "foo",
				Body: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
				Override: []interface{}{
					&UserDefinedTypeName{
						Node:     Node{Type: "UserDefinedTypeName"},
						NamePath: "Base1",
					},
					&UserDefinedTypeName{
						Node:     Node{Type: "UserDefinedTypeName"},
						NamePath: "Base2",
					},
				},
			},
		},

		// errors

		{
			parseNode(t, "error MyCustomError();"),
			&CustomErrorDefinition{
				Node: Node{Type: "CustomErrorDefinition"},
				Name: "MyCustomError",
			},
		},

		{
			parseNode(t, "error MyCustomError(uint a);"),
			&CustomErrorDefinition{
				Node: Node{Type: "CustomErrorDefinition"},
				Name: "MyCustomError",
				Parameters: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint",
						},
						Name: "a",
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "a",
						},
					},
				},
			},
		},

		{
			parseNode(t, "error MyCustomError(string);"),
			&CustomErrorDefinition{
				Node: Node{Type: "CustomErrorDefinition"},
				Name: "MyCustomError",
				Parameters: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "string",
						},
					},
				},
			},
		},

		{
			parseExpression(t, "2 ether"),
			&NumberLiteral{
				Node:            Node{Type: "NumberLiteral"},
				Number:          "2",
				SubDenomination: "ether",
			},
		},
		{
			parseExpression(t, "2.3e5"),
			&NumberLiteral{
				Node:   Node{Type: "NumberLiteral"},
				Number: "2.3e5",
			},
		},
		{
			parseExpression(t, ".1"),
			&NumberLiteral{
				Node:   Node{Type: "NumberLiteral"},
				Number: ".1",
			},
		},
		{
			parseExpression(t, "1_000_000"),
			&NumberLiteral{
				Node:   Node{Type: "NumberLiteral"},
				Number: "1_000_000",
			},
		},

		// string

		{
			parseExpression(t, `"Hello"`),
			&StringLiteral{
				Node:      Node{Type: "StringLiteral"},
				Value:     "Hello",
				Parts:     []string{"Hello"},
				IsUnicode: []bool{false},
			},
		},
		{
			parseExpression(t, `'Hello'`),
			&StringLiteral{
				Node:      Node{Type: "StringLiteral"},
				Value:     "Hello",
				Parts:     []string{"Hello"},
				IsUnicode: []bool{false},
			},
		},

		{
			parseExpression(t, `hex"fafafa"`),
			&HexLiteral{
				Node:  Node{Type: "HexLiteral"},
				Value: "fafafa",
				Parts: []string{"fafafa"},
			},
		},

		{
			parseExpression(t, `hex""`),
			&HexLiteral{
				Node:  Node{Type: "HexLiteral"},
				Value: "",
				Parts: []string{""},
			},
		},

		{
			parseExpression(t, "false"),
			&BooleanLiteral{
				Node:  Node{Type: "BooleanLiteral"},
				Value: false,
			},
		},

		{
			parseNode(t, "mapping(uint => address) a;"),
			&StateVariableDeclaration{
				Node: Node{Type: "StateVariableDeclaration"},
				Variables: []interface{}{
					&StateVariableDeclarationVariable{
						VariableDeclaration: VariableDeclaration{
							Node: Node{Type: "VariableDeclaration"},
							Name: "a",
							Identifier: &Identifier{
								Node: Node{Type: "Identifier"},
								Name: "a",
							},
							TypeName: &Mapping{
								Node: Node{Type: "Mapping"},
								KeyType: &ElementaryTypeName{
									Node: Node{Type: "ElementaryTypeName"},
									Name: "uint",
								},
								ValueType: &ElementaryTypeName{
									Node: Node{Type: "ElementaryTypeName"},
									Name: "address",
								},
							},
							Visibility: "default",
							IsStateVar: true,
						},
					},
				},
			},
		},

		{
			parseNode(t, "mapping(Foo => address) a;"),
			&StateVariableDeclaration{
				Node: Node{Type: "StateVariableDeclaration"},
				Variables: []interface{}{
					&StateVariableDeclarationVariable{
						VariableDeclaration: VariableDeclaration{
							Node: Node{Type: "VariableDeclaration"},
							Name: "a",
							Identifier: &Identifier{
								Node: Node{Type: "Identifier"},
								Name: "a",
							},
							TypeName: &Mapping{
								Node: Node{Type: "Mapping"},
								KeyType: &UserDefinedTypeName{
									Node:     Node{Type: "UserDefinedTypeName"},
									NamePath: "Foo",
								},
								ValueType: &ElementaryTypeName{
									Node: Node{Type: "ElementaryTypeName"},
									Name: "address",
								},
							},
							Visibility: "default",
							IsStateVar: true,
						},
					},
				},
			},
		},

		{
			parseExpression(t, "new MyContract"),
			&NewExpression{
				Node: Node{Type: "NewExpression"},
				TypeName: &UserDefinedTypeName{
					Node:     Node{Type: "UserDefinedTypeName"},
					NamePath: "MyContract",
				},
			},
		},

		// function call

		{
			parseExpression(t, "f(1, 2)"),
			&FunctionCall{
				Node: Node{Type: "FunctionCall"},
				Expression: &Identifier{
					Node: Node{Type: "Identifier"},
					Name: "f",
				},
				Arguments: []interface{}{
					&NumberLiteral{
						Node:   Node{Type: "NumberLiteral"},
						Number: "1",
					},
					&NumberLiteral{
						Node:   Node{Type: "NumberLiteral"},
						Number: "2",
					},
				},
			},
		},

		{
			parseExpression(t, "type(MyContract)"),
			&FunctionCall{
				Node: Node{Type: "FunctionCall"},
				Expression: &Identifier{
					Node: Node{Type: "Identifier"},
					Name: "type",
				},
				Arguments: []interface{}{
					&Identifier{
						Node: Node{Type: "Identifier"},
						Name: "MyContract",
					},
				},
			},
		},

		{
			parseExpression(t, "f{value: 10}(1, 2)"),
			&FunctionCall{
				Node: Node{Type: "FunctionCall"},
				Expression: &NameValueExpression{
					Node: Node{Type: "NameValueExpression"},
					Expression: &Identifier{
						Node: Node{Type: "Identifier"},
						Name: "f",
					},
					Arguments: &NameValueList{
						Node:  Node{Type: "NameValueList"},
						Names: []string{"value"},
						Identifiers: []interface{}{
							&Identifier{
								Node: Node{Type: "Identifier"},
								Name: "value",
							},
						},
						Args: []interface{}{
							&NumberLiteral{
								Node:   Node{Type: "NumberLiteral"},
								Number: "10",
							},
						},
					},
				},
				Arguments: []interface{}{
					&NumberLiteral{
						Node:   Node{Type: "NumberLiteral"},
						Number: "1",
					},
					&NumberLiteral{
						Node:   Node{Type: "NumberLiteral"},
						Number: "2",
					},
				},
			},
		},

		{
			parseExpression(t, "f({x: 1, y: 2})"),
			&FunctionCall{
				Node: Node{Type: "FunctionCall"},
				Expression: &Identifier{
					Node: Node{Type: "Identifier"},
					Name: "f",
				},
				Arguments: []interface{}{
					&NumberLiteral{
						Node:   Node{Type: "NumberLiteral"},
						Number: "1",
					},
					&NumberLiteral{
						Node:   Node{Type: "NumberLiteral"},
						Number: "2",
					},
				},
				Names: []interface{}{"x", "y"},
				Identifiers: []interface{}{
					&Identifier{
						Node: Node{Type: "Identifier"},
						Name: "x",
					},
					&Identifier{
						Node: Node{Type: "Identifier"},
						Name: "y",
					},
				},
			},
		},

		{
			parseExpression(t, "payable(recipient)"),
			&FunctionCall{
				Node: Node{Type: "FunctionCall"},
				Expression: &Identifier{
					Node: Node{Type: "Identifier"},
					Name: "payable",
				},
				Arguments: []interface{}{
					&Identifier{
						Node: Node{Type: "Identifier"},
						Name: "recipient",
					},
				},
			},
		},

		// structs

		{
			parseNode(t, "struct hello { uint a; }"),
			&StructDefinition{
				Node: Node{Type: "StructDefinition"},
				Name: "hello",
				Members: []interface{}{
					&VariableDeclaration{
						Node: Node{Type: "VariableDeclaration"},
						Name: "a",
						TypeName: &ElementaryTypeName{
							Node: Node{Type: "ElementaryTypeName"},
							Name: "uint",
						},
						Identifier: &Identifier{
							Node: Node{Type: "Identifier"},
							Name: "a",
						},
					},
				},
			},
		},

		{
			parseNode(t, "type Price is uint128;"),
			&TypeDefinition{
				Node: Node{Type: "TypeDefinition"},
				Name: "Price",
				Definition: &ElementaryTypeName{
					Node: Node{Type: "ElementaryTypeName"},
					Name: "uint128",
				},
			},
		},

		// unchecked
		{
			parseStatement(t, "unchecked { }"),
			&UncheckedStatement{
				Node: Node{Type: "UncheckedStatement"},
				Block: &Block{
					Node:       Node{Type: "Block"},
					Statements: []interface{}{},
				},
			},
		},
		{
			parseStatement(t, "unchecked { x++; }"),
			&UncheckedStatement{
				Node: Node{Type: "UncheckedStatement"},
				Block: &Block{
					Node: Node{Type: "Block"},
					Statements: []interface{}{
						&ExpressionStatement{
							Node: Node{Type: "ExpressionStatement"},
							Expression: &UnaryOperation{
								Node:     Node{Type: "UnaryOperation"},
								Operator: "++",
								SubExpression: &Identifier{
									Node: Node{Type: "Identifier"},
									Name: "x",
								},
							},
						},
					},
				},
			},
		},
	}

	testSolidityCase(t, cases)
}

func testSolidityCase(t *testing.T, cases parserCase) {
	for indx, c := range cases {
		fmt.Println("-----------------")
		if !reflect.DeepEqual(c.result, c.code) {

			print := func(c interface{}) {
				data, err := json.Marshal(c)
				if err != nil {
					t.Fatal(err)
				}
				t.Log(reflect.TypeOf(c))
				t.Log(string(data))
			}

			print(c.code)
			print(c.result)
			t.Fatalf("bad %d", indx)

		}
	}
}

func parse(t *testing.T, source string) interface{} {
	p := Parse(source)
	return p.Result
}

func parseContract(t *testing.T, source string) interface{} {
	return parse(t, source).(*SourceUnit).Children[0]
}

func parseNode(t *testing.T, source string) interface{} {
	p := parseContract(t, "contract test { "+source+" }")
	return p.(*ContractDefinition).SubNodes[0]
}

func parseStatement(t *testing.T, source string) interface{} {
	p := parseNode(t, "function () { "+source+" }")
	return p.(*FunctionDefinition).Body.(*Block).Statements[0]
}

func parseExpression(t *testing.T, source string) interface{} {
	p := parseNode(t, "function () { "+source+"; }")
	return p.(*FunctionDefinition).Body.(*Block).Statements[0].(*ExpressionStatement).Expression
}
