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
	metrics  map[string]int
}

func NewCachedParser() *CachedParser {
	metrics := make(map[string]int)

	metrics["ini_hit"] = 0
	metrics["ini_miss"] = 0
	metrics["fnmatch_hit"] = 0
	metrics["fnmatch_miss"] = 0

	return &CachedParser{
		iniFiles: make(map[string]*ini.File),
		regexps:  make(map[string]*regexp.Regexp),
		metrics:  metrics,
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

		parser.metrics["ini_miss"]++
	} else {
		parser.metrics["ini_hit"]++
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

		parser.metrics["fnmatch_miss"]++
	} else {
		parser.metrics["fnmatch_hit"]++
	}

	return r.MatchString(filename), nil
}

// Metrics returns the metrics from the cache
func (parser *CachedParser) Metrics() map[string]int {
	m := make(map[string]int)

	for k, v := range parser.metrics {
		m[k] = v
	}

	return m
}
