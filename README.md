# Go Editorconfig

A Go parser for [Editorconfig][editorconfig] files.

## Usage

### Parse from file

```go
editorConfig, err := editorconfig.ParseFile("path/to/.editorconfig")
if err != nil {
    log.Fatal(err)
}
```

### Parse from slice of bytes

```go
data := []byte("...")
editorConfig, err := editorconfig.ParseBytes(data)
if err != nil {
    log.Fatal(err)
}
```

### Get definition to a given filename.

This method builds a definition to a given filename.
This definition is a merge of the properties with selectors that matched the
given filename.
The lasts sections of the file have preference over the priors.

```go
def := editorConfig.GetDefinitionForFilename("my/file.go")
```

This definition have the following properties:

```go
type Definition struct {
	Selector string

	Charset                string
	IndentStyle            string
	IndentSize             string
	TabWidth               int
	EndOfLine              string
	TrimTrailingWhitespace bool
	InsertFinalNewline     bool
}
```

## Contributing

To run the tests:

```bash
go test
```

[editorconfig]: http://editorconfig.org/
