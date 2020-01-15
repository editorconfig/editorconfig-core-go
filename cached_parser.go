package editorconfig

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/ini.v1"
)

type CachedParser struct {
	iniFiles map[string]*ini.File
	regexps  map[string]*regexp.Regexp
}

func NewCachedParser() *CachedParser {
	return &CachedParser{
		iniFiles: make(map[string]*ini.File),
		regexps:  make(map[string]*regexp.Regexp),
	}
}

func (parser *CachedParser) ParseIni(filename string) (*ini.File, error) {
	iniFile, ok := parser.iniFiles[filename]
	if !ok {
		fp, err := os.Open(filename)
		if err != nil {
			return nil, err
		}

		defer fp.Close()

		iniFile, err = ini.Load(fp)
		if err != nil {
			return nil, err
		}

		parser.iniFiles[filename] = iniFile
	}

	return iniFile, nil
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
