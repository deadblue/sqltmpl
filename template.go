package sqltmpl

import (
	"fmt"
	"strings"
	"text/template/parse"
)

const (
	_RootName = "root"
)

type Template[P any] interface {
	// Render renders template with given params.
	//
	// The result is rendered query statement, corresponding SQL arguments, and
	// an error that will be not-nil when rendering failed.
	Render(params P) (query string, args []any, err error)

	// MustRender is like [Render] but panics when rendering failed.
	MustRender(params P) (query string, args []any)
}

// Parse parses SQL template and returns a [Template] object if successful.
func Parse[P any](line ...string) (tmpl Template[P], err error) {
	var text string
	if lineCount := len(line); lineCount == 0 {
		err = errEmptyTemplate
		return
	} else if lineCount == 1 {
		text = line[0]
	} else {
		text = strings.Join(line, "\n")
	}
	// Parse template
	treeSet, err := parse.Parse(_RootName, text, "", "")
	if err != nil {
		return
	}
	tree := treeSet[_RootName]
	if hasDynamicNode(tree.Root) {
		tmpl = &_DynamicTemplate[P]{
			root: tree.Root,
		}
	} else {
		tmpl, err = (&_StaticTemplateBuilder[P]{}).Build(tree.Root)
	}
	return
}

// MustParse is like [Parse] but panics when parsing failed.
func MustParse[P any](line ...string) Template[P] {
	if tmpl, err := Parse[P](line...); err != nil {
		panic(fmt.Sprintf(
			"Parse SQL template failed!\nError: %s", err,
		))
	} else {
		return tmpl
	}
}

// hasDynamicNode checks whether template tree has dynamic node.
func hasDynamicNode(root *parse.ListNode) bool {
	for _, node := range root.Nodes {
		switch nodeType := node.Type(); nodeType {
		case parse.NodeIf, parse.NodeRange, parse.NodeWith:
			return true
		}
	}
	return false
}
