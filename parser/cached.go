package parser

import (
	"fmt"
	"os"
	"regexp"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"gopkg.in/ini.v1"
)

// Cached implements the  interface but caches the definition and
// the regular expressions.
type Cached struct {
	editorconfigs map[string]*editorconfig.Editorconfig
	regexps       map[string]*regexp.Regexp
}

// NewCached initializes the Cached.
func NewCached() *Cached {
	return &Cached{
		editorconfigs: make(map[string]*editorconfig.Editorconfig),
		regexps:       make(map[string]*regexp.Regexp),
	}
}

// ParseIni parses the given filename to a Definition and caches the result.
func (parser *Cached) ParseIni(filename string) (*editorconfig.Editorconfig, error) {
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

		ec, err = editorconfig.NewEditorconfig(iniFile)
		if err != nil {
			return nil, err
		}

		parser.editorconfigs[filename] = ec
	}

	return ec, nil
}

// FnmatchCase calls the module's FnmatchCase and caches the parsed selector.
func (parser *Cached) FnmatchCase(selector string, filename string) (bool, error) {
	r, ok := parser.regexps[selector]
	if !ok {
		p := editorconfig.Translate(selector)

		var err error

		r, err = regexp.Compile(fmt.Sprintf("^%s$", p))
		if err != nil {
			return false, err
		}

		parser.regexps[selector] = r
	}

	return r.MatchString(filename), nil
}
