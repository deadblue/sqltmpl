package sqltmpl

import (
	"errors"
	"fmt"
	"reflect"
	"text/template/parse"
)

var (
	errEmptyTemplate = errors.New("template is empty")

	errChainedPipeline = errors.New("chained pipeline is unsupported")
	errFunctionCalling = errors.New("function or method calling is unsupported")

	errInvalidRangeValue = errors.New("range value should be array, slice or map")
)

func raiseUnsupportedNode(nt parse.NodeType) error {
	return fmt.Errorf("unsupportted node type: %d", nt)
}

func assertSupportedPipeNode(pipe *parse.PipeNode) error {
	if len(pipe.Cmds) > 1 {
		return errChainedPipeline
	}
	if len(pipe.Cmds[0].Args) > 1 {
		return errFunctionCalling
	}
	return nil
}

func assertRangeValue(value reflect.Value) error {
	switch value.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice:
		return nil
	default:
		return errInvalidRangeValue
	}
}
