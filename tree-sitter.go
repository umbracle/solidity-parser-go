package solcparser

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"
	treesitter "github.com/umbracle/solidity-parser-go/tree-sitter"
)

func NewTreeSitter(code string) *sitter.Node {
	n, _ := sitter.ParseCtx(context.Background(), []byte(code), treesitter.GetLanguage())
	return n
}
