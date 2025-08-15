package sqltmpl

import (
	"fmt"
	"text/template/parse"
)

// nodeToText converts argument node to text string.
func nodeToText(node parse.Node) (text string) {
	switch node := node.(type) {
	case *parse.BoolNode:
		text = fmt.Sprintf("%t", node.True)
	case *parse.NumberNode:
		text = node.Text
	case *parse.StringNode:
		text = node.Text
	case *parse.NilNode:
		text = "NULL"
	}
	return
}
