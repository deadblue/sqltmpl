package sqltmpl_test

import (
	"fmt"

	"github.com/deadblue/sqltmpl"
)

type Params struct {
	Ids  []int
	Type string
}

func Example() {
	tmpl := sqltmpl.MustParse[Params](
		"SELECT a, b, c FROM table WHERE id IN (",
		"{{- range $index, $elem := .Ids -}}",
		"{{ if $index }}, {{ end }}{{ $elem }}",
		"{{- end -}}",
		") AND type = {{ .Type }}",
	)

	query, args := tmpl.MustRender(Params{
		Ids:  []int{1, 2, 3},
		Type: "foobar",
	})
	fmt.Printf("Query: %s\n", query)
	fmt.Printf("Arguments: %v", args)

	// Output:
	// Query: SELECT a, b, c FROM table WHERE id IN (?, ?, ?) AND type = ?
	// Arguments: [1 2 3 foobar]
}
