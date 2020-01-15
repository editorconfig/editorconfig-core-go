package editorconfig

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/ini.v1"
)

type CachedParser struct {
	editorconfigs map[string]*Editorconfig
	regexps       map[string]*regexp.Regexp
}

func NewCachedParser() *CachedParser {
	return &CachedParser{
		editorconfigs: make(map[string]*Editorconfig),
		regexps:       make(map[string]*regexp.Regexp),
	}
}

func (parser *CachedParser) ParseIni(filename string) (*Editorconfig, error) {
	ec, ok := parser.editorconfigs[filename]
	if !ok {
		fp, err := os.Open(filename)
		if err != nil {
			return nil, err
		}

		defer fp.Close()

		iniFile, err := ini.Load(fp)
		if err != nil {
			return nil, err
		}

		ec, err = newEditorconfig(iniFile)
		if err != nil {
			return nil, err
		}

		parser.editorconfigs[filename] = ec
	}

	return ec, nil
}

// FnmatchCase calls the module's FnmatchCase.
func (parser *CachedParser) FnmatchCase(selector string, filename string) (bool, error) {
	r, ok := parser.regexps[selector]
	if !ok {
		p := translate(selector)

		var err error
		r, err = regexp.Compile(fmt.Sprintf("^%s$", p))
		if err != nil {
			return false, err
		}

		parser.regexps[selector] = r
	}

	return r.MatchString(filename), nil
}
