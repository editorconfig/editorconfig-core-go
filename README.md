# Editorconfig Core Go

A [Editorconfig][editorconfig] file parser and manipulator for Go.

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

### Get definition to a given filename

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

### Generating a .editorconfig file

You can easily convert a Editorconfig struct to a compatible INI file:

```go
// serialize to slice of bytes
data, err := editorConfig.Serialize()
if err != nil {
    log.Fatal(err)
}

// save directly to file
err := editorConfig.Save("path/to/.editorconfig")
if err != nil {
    log.Fatal(err)
}
```

## Contributing

To run the tests:

```bash
go test
```

[editorconfig]: http://editorconfig.org/
