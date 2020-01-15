package editorconfig

import (
	"gopkg.in/ini.v1"
)

// Parser interface is responsible for the parsing of the ini file and the
// globbing patterns.
type Parser interface {
	ParseIni(string) (*ini.File, error)
	FnmatchCase(string, string) (bool, error)
}
