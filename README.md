# Elastigo

Create Elasticsearch mappings from Go structs

## Elasticsearch data types

There is a small DSL for mapping Go types to Elasticsearch types. Like the `json` tags
found on structs, it is possible to provide custom behavior for Go struct fields.

To specify a custom data type, such as `date` for a `uint64`, put it right after the
`es` tag:

```go
type MyStruct struct {
    ts          uint64 `es:"date"`
    description string `es:"text"`
}
```

Additional custom properties must be specified _after_ the first item:

```go
type MyStruct struct {
    ts          uint64 `es:"date,epoch_ms"`
    description string `es:",indexignore"`
}
```

## Currently supported customizations

| Property name | Description |
| --- | --- |
| `epoch_seconds` | Only to be used with a `date` data type, this specifies that the timestamp is in seconds. |
| `epoch_ms` | Similar to `epoch_seconds`, but with milliseconds instead. |
| `indexignore` | Tells Elasticsearch not to index (make searchable) this field. |
| `eager_global_ordinals` | Tells Elasticsearch to load ordinals eagerly rather than lazily (for search speed). |

## Tests

To run tests:

```bash
go test .
```

