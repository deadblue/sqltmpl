package sqltmpl

import (
	"reflect"
	"strings"
	"text/template/parse"
)

type _StaticTemplate[P any] struct {
	// Query statement
	query string
	// Argument fields
	argFields [][]string
}

func (t *_StaticTemplate[P]) Render(params P) (query string, args []any, err error) {
	query = t.query
	rv := reflect.ValueOf(params)
	args = make([]any, len(t.argFields))
	for index, fields := range t.argFields {
		args[index], err = toRawValue(getValue(rv, fields...))
		if err != nil {
			return
		}
	}
	return
}

func (t *_StaticTemplate[P]) MustRender(params P) (query string, args []any) {
	query, args, err := t.Render(params)
	if err != nil {
		panic(err)
	}
	return
}

type _StaticTemplateBuilder[P any] struct {
	// Statement buffer
	buf strings.Builder
	// Reference fields
	argFields [][]string
}

func (b *_StaticTemplateBuilder[P]) renderActionNode(node *parse.ActionNode) (err error) {
	pipe := node.Pipe
	if err = assertSupportedPipeNode(pipe); err != nil {
		return
	}
	arg := pipe.Cmds[0].Args[0]
	switch nodeType := arg.Type(); nodeType {
	case parse.NodeBool, parse.NodeNumber, parse.NodeNil, parse.NodeString:
		b.buf.WriteString(nodeToText(arg))
	case parse.NodeDot:
		b.buf.Write(_ParamPlaceholder)
		b.argFields = append(b.argFields, _EmtypStringSlice)
	case parse.NodeField:
		b.buf.Write(_ParamPlaceholder)
		arg := arg.(*parse.FieldNode)
		b.argFields = append(b.argFields, arg.Ident)
	default:
		err = raiseUnsupportedNode(nodeType)
	}
	return
}

func (b *_StaticTemplateBuilder[P]) Build(root *parse.ListNode) (tmpl *_StaticTemplate[P], err error) {
	for _, node := range root.Nodes {
		switch node := node.(type) {
		case *parse.TextNode:
			b.buf.Write(node.Text)
		case *parse.ActionNode:
			b.renderActionNode(node)
		case *parse.CommentNode:
			// Skip comment node
		default:
			err = raiseUnsupportedNode(node.Type())
			return
		}
	}
	tmpl = &_StaticTemplate[P]{
		query: b.buf.String(),
	}
	tmpl.argFields = append(tmpl.argFields, b.argFields...)
	return
}
