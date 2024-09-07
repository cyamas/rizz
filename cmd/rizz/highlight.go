package main

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

var highlights = map[string]string{
	"package_clause":       "blue",
	"import_declaration":   "green",
	"function_declaration": "purple",
}

func printTree(n *sitter.Node, gen int) {
	tabs := strings.Repeat("  ", gen)
	for i := range n.ChildCount() {
		child := n.Child(int(i))
		color := "white"
		if c, ok := highlights[child.Type()]; ok {
			color = c
		}
		fmt.Println(tabs, child.Type(), color, child.StartPoint(), "-", child.EndPoint())
		fmt.Println("")
		if child.ChildCount() > 0 {
			gen++
			printTree(child, gen)
		}
	}
}
