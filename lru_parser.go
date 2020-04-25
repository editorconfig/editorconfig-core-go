package editorconfig

import (
	"fmt"
	"os"
	"regexp"

	"github.com/hashicorp/golang-lru"
	"gopkg.in/ini.v1"
)

// LRUParser implements the Parser interface but caches the definition and
// the regular expressions using Hashicorp's LRU cache.
//
// https://github.com/hashicorp/golang-lru
type LRUParser struct {
	editorconfigs *lru.Cache
	regexps       *lru.Cache
}

// NewLRUParser initializes the LRUParser.
func NewLRUParser() *LRUParser {
	c, _ := lru.New(64)
	r, _ := lru.New(256)
	return &LRUParser{
		editorconfigs: c,
		regexps:       r,
	}
}

// ParseIni parses the given filename to a Definition and caches the result.
func (parser *LRUParser) ParseIni(filename string) (*Editorconfig, error) {
	ec, ok := parser.editorconfigs.Get(filename)
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

		parser.editorconfigs.Add(filename, ec)
	}

	return ec.(*Editorconfig), nil
}

// FnmatchCase calls the module's FnmatchCase and caches the parsed selector.
func (parser *LRUParser) FnmatchCase(selector string, filename string) (bool, error) {
	r, ok := parser.regexps.Get(selector)
	if !ok {
		p := translate(selector)

		var err error

		r, err = regexp.Compile(fmt.Sprintf("^%s$", p))
		if err != nil {
			return false, err
		}

		parser.regexps.Add(selector, r)
	}

	return r.(*regexp.Regexp).MatchString(filename), nil
}
