package editorconfig

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"gopkg.in/ini.v1"
)

// CachedParser implements the Parser interface but caches the definition and
// the regular expressions.
type CachedParser struct {
	editorconfigs map[string]*Editorconfig
	regexps       map[string]*regexp.Regexp
}

// NewCachedParser initializes the CachedParser.
func NewCachedParser() *CachedParser {
	return &CachedParser{
		editorconfigs: make(map[string]*Editorconfig),
		regexps:       make(map[string]*regexp.Regexp),
	}
}

// ParseIni parses the given filename to a Definition and caches the result.
func (parser *CachedParser) ParseIni(filename string) (*Editorconfig, error) {
	ec, warning, err := parser.ParseIniGraceful(filename)
	if err != nil {
		return nil, err
	}

	return ec, warning
}

// ParseIniGraceful parses the given filename to a Definition and caches the result.
func (parser *CachedParser) ParseIniGraceful(filename string) (*Editorconfig, error, error) {
	var warning error

	ec, ok := parser.editorconfigs[filename]
	if !ok {
		fp, err := os.Open(filename)
		if err != nil {
			return nil, nil, fmt.Errorf("error opening %q: %w", filename, err)
		}

		defer fp.Close()

		iniFile, err := ini.Load(fp)
		if err != nil {
			return nil, nil, fmt.Errorf("error loading ini file %q: %w", filename, err)
		}

		var warn error

		ec, warn, err = newEditorconfig(iniFile)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating config: %w", err)
		}

		if warn != nil {
			warning = errors.Join(warning, warn)
		}

		parser.editorconfigs[filename] = ec
	}

	return ec, warning, nil
}

// FnmatchCase calls the module's FnmatchCase and caches the parsed selector.
func (parser *CachedParser) FnmatchCase(selector string, filename string) (bool, error) {
	r, ok := parser.regexps[selector]
	if !ok {
		p := translate(selector)

		var err error

		r, err = regexp.Compile(fmt.Sprintf("^%s$", p))
		if err != nil {
			return false, fmt.Errorf("error compiling selector %q: %w", selector, err)
		}

		parser.regexps[selector] = r
	}

	return r.MatchString(filename), nil
}
