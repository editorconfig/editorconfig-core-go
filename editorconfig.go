// Package editorconfig can be used to parse and generate editorconfig files.
// For more information about editorconfig, see http://editorconfig.org/
package editorconfig

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

// IndentStyle possible values
const (
	IndentStyleTab    = "tab"
	IndentStyleSpaces = "space"
)

// EndOfLine possible values
const (
	EndOfLineLf   = "lf"
	EndOfLineCr   = "cr"
	EndOfLineCrLf = "crlf"
)

// Charset possible values
const (
	CharsetLatin1  = "latin1"
	CharsetUTF8    = "utf-8"
	CharsetUTF16BE = "utf-16be"
	CharsetUTF16LE = "utf-16le"
)

// Definition represents a definition inside the .editorconfig file.
// E.g. a section of the file.
// The definition is composed of the selector ("*", "*.go", "*.{js.css}", etc),
// plus the properties of the selected files.
type Definition struct {
	Selector string

	Charset                string `ini:"charset"`
	IndentStyle            string `ini:"indent_style"`
	IndentSize             string `ini:"indent_size"`
	TabWidth               int    `ini:"tab_width"`
	EndOfLine              string `ini:"end_of_line"`
	TrimTrailingWhitespace bool   `ini:"trim_trailing_whitespace"`
	InsertFinalNewline     bool   `ini:"insert_final_newline"`
}

// Editorconfig represents a .editorconfig file.
// It is composed by a "root" property, plus the definitions defined in the
// file.
type Editorconfig struct {
	Root        bool
	Definitions []*Definition
}

// ParseBytes parses from a slice of bytes.
func ParseBytes(data []byte) (*Editorconfig, error) {
	iniFile, err := ini.Load(data)
	if err != nil {
		return nil, err
	}

	editorConfig := &Editorconfig{}
	editorConfig.Root = iniFile.Section(ini.DEFAULT_SECTION).Key("root").MustBool(false)
	for _, sectionStr := range iniFile.SectionStrings() {
		if sectionStr == ini.DEFAULT_SECTION {
			continue
		}
		var (
			iniSection = iniFile.Section(sectionStr)
			definition = &Definition{}
		)
		err := iniSection.MapTo(&definition)
		if err != nil {
			return nil, err
		}

		// tab_width defaults to indent_size:
		// https://github.com/editorconfig/editorconfig/wiki/EditorConfig-Properties#tab_width
		if definition.TabWidth <= 0 {
			if num, err := strconv.Atoi(definition.IndentSize); err == nil {
				definition.TabWidth = num
			}
		}

		definition.Selector = sectionStr
		editorConfig.Definitions = append(editorConfig.Definitions, definition)
	}
	return editorConfig, nil
}

// ParseFile parses from a file.
func ParseFile(f string) (*Editorconfig, error) {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}

	return ParseBytes(data)
}

var (
	regexpMultiExtension = regexp.MustCompile("\\.{.*}$")
)

func filenameMatches(pattern, name string) bool {
	// basic match
	matched, _ := filepath.Match(pattern, name)
	if matched {
		return true
	}
	// foo/bar/main.go should match main.go
	matched, _ = filepath.Match(pattern, filepath.Base(name))
	if matched {
		return true
	}
	// foo should match foo/main.go
	matched, _ = filepath.Match(filepath.Join(pattern, "*"), name)
	if matched {
		return true
	}
	// *.{js,go} should match main.go
	if str := regexpMultiExtension.FindString(pattern); len(str) > 0 {
		// remote initial ".{" and final "}"
		str = str[2 : len(str)-1]

		for _, ext := range strings.Split(str, ",") {
			matched, _ = filepath.Match(fmt.Sprintf("*.%s", ext), filepath.Base(name))
			if matched {
				return true
			}
		}
	}
	return false
}

func (d *Definition) merge(md *Definition) {
	if len(d.Charset) == 0 {
		d.Charset = md.Charset
	}
	if len(d.IndentStyle) == 0 {
		d.IndentStyle = md.IndentStyle
	}
	if len(d.IndentSize) == 0 {
		d.IndentSize = md.IndentSize
	}
	if d.TabWidth <= 0 {
		d.TabWidth = md.TabWidth
	}
	if len(d.EndOfLine) == 0 {
		d.EndOfLine = md.EndOfLine
	}
	if !d.TrimTrailingWhitespace {
		d.TrimTrailingWhitespace = md.TrimTrailingWhitespace
	}
	if !d.InsertFinalNewline {
		d.InsertFinalNewline = md.InsertFinalNewline
	}
}

// GetDefinitionForFilename returns a definition for the given filename.
// The result is a merge of the selectors that matched the file.
// The last section has preference over the priors.
func (e *Editorconfig) GetDefinitionForFilename(name string) *Definition {
	def := &Definition{}
	for i := len(e.Definitions) - 1; i >= 0; i-- {
		actualDef := e.Definitions[i]
		if filenameMatches(actualDef.Selector, name) {
			def.merge(actualDef)
		}
	}
	return def
}
