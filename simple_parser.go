package editorconfig

import (
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

// SimpleParser implements the Parser interface but without doing any caching.
type SimpleParser struct{}

// ParseInit calls go-ini's Load on the file.
func (parser *SimpleParser) ParseIni(filename string) (*Editorconfig, error) {
	ec, warning, err := parser.ParseIniGraceful(filename)
	if err != nil {
		return nil, err
	}

	return ec, warning
}

// ParseIni calls go-ini's Load on the file and keep warnings in a separate error.
func (parser *SimpleParser) ParseIniGraceful(filename string) (*Editorconfig, error, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, nil, err //nolint:wrapcheck
	}

	defer fp.Close()

	iniFile, err := ini.Load(fp)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot load %q: %w", filename, err)
	}

	return newEditorconfig(iniFile)
}

// FnmatchCase calls the module's FnmatchCase.
func (parser *SimpleParser) FnmatchCase(selector string, filename string) (bool, error) {
	return FnmatchCase(selector, filename)
}
