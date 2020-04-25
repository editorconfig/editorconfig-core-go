package parser

import (
	"fmt"
	"os"
	"regexp"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/hashicorp/golang-lru"
	"gopkg.in/ini.v1"
)

// LRU implements the  interface but caches the definition and
// the regular expressions using Hashicorp's LRU cache.
//
// https://github.com/hashicorp/golang-lru
type LRU struct {
	editorconfigs *lru.Cache
	regexps       *lru.Cache
}

// NewLRU initializes the LRU.
func NewLRU() *LRU {
	c, _ := lru.New(64)
	r, _ := lru.New(256)
	return &LRU{
		editorconfigs: c,
		regexps:       r,
	}
}

// ParseIni parses the given filename to a Definition and caches the result.
func (parser *LRU) ParseIni(filename string) (*editorconfig.Editorconfig, error) {
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

		ec, err = editorconfig.NewEditorconfig(iniFile)
		if err != nil {
			return nil, err
		}

		parser.editorconfigs.Add(filename, ec)
	}

	return ec.(*editorconfig.Editorconfig), nil
}

// FnmatchCase calls the module's FnmatchCase and caches the parsed selector.
func (parser *LRU) FnmatchCase(selector string, filename string) (bool, error) {
	r, ok := parser.regexps.Get(selector)
	if !ok {
		p := editorconfig.Translate(selector)

		var err error

		r, err = regexp.Compile(fmt.Sprintf("^%s$", p))
		if err != nil {
			return false, err
		}

		parser.regexps.Add(selector, r)
	}

	return r.(*regexp.Regexp).MatchString(filename), nil
}
