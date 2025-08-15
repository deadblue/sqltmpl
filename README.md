# SQL-Template

A template engine for SQL in Golang.

## Example

```go
import (
    "github.com/deadblue/sqltmpl"
)

func main() {
    tmpl := sqltmpl.MustParse[string](
        "SELECT * FROM table WHERE id = {{ . }}",
    )
    var query string
    var args []any

    query, args = tmpl.MustRender("a")
    // query: SELECT * FROM table WHERE id = ?
    // args: ["a"]

    query, args = tmpl.MustRender("b")
    // query: SELECT * FROM table WHERE id = ?
    // args: ["b"]
}
```

## Specification

The SQL template follows the specification of standard package "text/template" 
with some limitations in below:

- Function/Method calling is unsupported.
- Chained pipelines is unsupported.
- Variable definition is unsupported.

## License

MIT
