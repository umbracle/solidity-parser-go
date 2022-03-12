package solidity

import (
	"fmt"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

func TestGrammar(t *testing.T) {
	n := sitter.Parse([]byte(code), GetLanguage())

	fmt.Println("AST:", n)
	fmt.Println("Root type:", n.Type())
	fmt.Println("Root children:", n.ChildCount())

	/*
		assert.Equal(
			"(source_file (package_clause (package_identifier)))",
			n.String(),
		)
	*/
}

var code = `// pragma
pragma solidity >=0.4.25 <0.7;

// import
import "SomeFile";
import "SomeFile" as b;
import * from "SomeFile";
import * as c from "SomeFile";
import a as a from "SomeFile";
import {a} from "someFile";
import {a as b} from "Somefile";
import {a, c as b} from "someFile";

contract Metacoin {
    
}

enum example {
    optioN,
    option2,
}`
