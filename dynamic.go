package sqltmpl

import (
	"maps"
	"reflect"
	"strings"
	"text/template/parse"
)

type _DynamicTemplate[P any] struct {
	root *parse.ListNode
}

func (t *_DynamicTemplate[P]) Render(params P) (query string, args []any, err error) {
	renderer := _DynamicTemplateRenderer[P]{}
	if err = renderer.render(t.root, map[string]reflect.Value{
		_DotName: reflect.ValueOf(params),
	}); err == nil {
		query = renderer.buf.String()
		args = append(args, renderer.args...)
	}
	return
}

func (t *_DynamicTemplate[P]) MustRender(params P) (query string, args []any) {
	query, args, err := t.Render(params)
	if err != nil {
		panic(err)
	}
	return
}

type _DynamicTemplateRenderer[P any] struct {
	buf  strings.Builder
	args []any
}

func (r *_DynamicTemplateRenderer[P]) render(
	root *parse.ListNode, values map[string]reflect.Value,
) (err error) {
	if root == nil {
		return
	}
	for _, node := range root.Nodes {
		switch node := node.(type) {
		case *parse.TextNode:
			r.buf.Write(node.Text)
		case *parse.ActionNode:
			err = r.renderActionNode(node, values)
		case *parse.IfNode:
			err = r.renderIfNode(node, values)
		case *parse.RangeNode:
			err = r.renderRangeNode(node, values)
		case *parse.WithNode:
			err = r.renderWithNode(node, values)
		case *parse.CommentNode:
			// Skip comment node
		}
		if err != nil {
			break
		}
	}
	return
}

func (r *_DynamicTemplateRenderer[P]) renderActionNode(
	root *parse.ActionNode, values map[string]reflect.Value,
) (err error) {
	if err = assertSupportedPipeNode(root.Pipe); err != nil {
		return
	}
	arg := root.Pipe.Cmds[0].Args[0]
	switch nodeType := arg.Type(); nodeType {
	case parse.NodeBool, parse.NodeNumber, parse.NodeNil, parse.NodeString:
		r.buf.WriteString(nodeToText(arg))
	case parse.NodeDot, parse.NodeField, parse.NodeVariable:
		if argValue, err := toRawValue(calcArgNode(arg, values)); err == nil {
			r.args = append(r.args, argValue)
		} else {
			return err
		}
		r.buf.Write(_ParamPlaceholder)
	default:
		err = raiseUnsupportedNode(nodeType)
	}
	return
}

func (r *_DynamicTemplateRenderer[P]) renderIfNode(
	root *parse.IfNode, values map[string]reflect.Value,
) (err error) {
	value, err := calcPipeNode(root.Pipe, values)
	if err != nil {
		return
	}
	if toBoolValue(value) {
		r.render(root.List, values)
	} else {
		r.render(root.ElseList, values)
	}
	return
}

func (r *_DynamicTemplateRenderer[P]) renderRangeNode(
	root *parse.RangeNode, values map[string]reflect.Value,
) (err error) {
	pipe := root.Pipe
	rangeValue, err := calcPipeNode(root.Pipe, values)
	if err != nil {
		return
	}
	if err = assertRangeValue(rangeValue); err != nil {
		return
	}
	if rangeValue.Len() == 0 {
		return r.render(root.ElseList, values)
	}

	// Prepare values for range block
	var indexKey, elemKey string
	switch varCount := len(pipe.Decl); varCount {
	case 0:
		elemKey = _DotName
	case 1:
		elemKey = pipe.Decl[0].Ident[0]
	default:
		indexKey = pipe.Decl[0].Ident[0]
		elemKey = pipe.Decl[1].Ident[0]
	}
	rangeValues := maps.Clone(values)
	for indexValue, elemValue := range rangeValue.Seq2() {
		if indexKey != "" {
			rangeValues[indexKey] = indexValue
		}
		rangeValues[elemKey] = elemValue
		if err = r.render(root.List, rangeValues); err != nil {
			return
		}
	}
	return
}

func (r *_DynamicTemplateRenderer[P]) renderWithNode(
	root *parse.WithNode, values map[string]reflect.Value,
) (err error) {
	pipe := root.Pipe
	withValue, err := calcPipeNode(pipe, values)
	if err != nil {
		return
	}
	if !toBoolValue(withValue) {
		return r.render(root.ElseList, values)
	}

	// Determine value key
	withKey := _DotName
	if len(root.Pipe.Decl) > 0 {
		withKey = pipe.Decl[0].Ident[0]
	}
	// Prepare values for with block
	withValues := maps.Clone(values)
	withValues[withKey] = withValue
	// Render with content
	return r.render(root.List, withValues)
}

func calcPipeNode(
	pipe *parse.PipeNode, values map[string]reflect.Value,
) (result reflect.Value, err error) {
	// TODO: Support function?
	if err = assertSupportedPipeNode(pipe); err != nil {
		return
	}
	result = calcArgNode(pipe.Cmds[0].Args[0], values)
	return
}

func calcArgNode(
	node parse.Node,
	values map[string]reflect.Value,
) (result reflect.Value) {
	switch node := node.(type) {
	case *parse.BoolNode:
		result = reflect.ValueOf(node.True)
	case *parse.StringNode:
		result = reflect.ValueOf(node.Text)
	case *parse.NilNode:
		result = reflect.ValueOf(nil)
	case *parse.NumberNode:
		if node.IsInt {
			result = reflect.ValueOf(node.Int64)
		} else if node.IsUint {
			result = reflect.ValueOf(node.Uint64)
		} else if node.IsFloat {
			result = reflect.ValueOf(node.Float64)
		} else if node.IsComplex {
			result = reflect.ValueOf(node.Complex128)
		}
	case *parse.DotNode:
		result = values["."]
	case *parse.FieldNode:
		dotValue := values["."]
		result = getValue(dotValue, node.Ident...)
	case *parse.VariableNode:
		varName := node.Ident[0]
		varValue := values[varName]
		result = getValue(varValue, node.Ident[1:]...)
	}
	return
}
