package editorconfig

import (
	"os"

	"gopkg.in/ini.v1"
)

type SimpleParser struct{}

// ParseIni calls go-ini's Load on the file.
func (parser *SimpleParser) ParseIni(filename string) (*ini.File, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer fp.Close()

	return ini.Load(fp)
}

// FnmatchCase calls the module's FnmatchCase.
func (parser *SimpleParser) FnmatchCase(selector string, filename string) (bool, error) {
	return FnmatchCase(selector, filename)
}
