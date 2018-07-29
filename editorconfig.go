// Package editorconfig can be used to parse and generate editorconfig files.
// For more information about editorconfig, see http://editorconfig.org/
package editorconfig

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
	"github.com/danwakefield/fnmatch"
)

const (
	ConfigNameDefault = ".editorconfig"
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
	Selector string `ini:"-" json:"-"`

	Charset                string `ini:"charset" json:"charset,omitempty"`
	IndentStyle            string `ini:"indent_style" json:"indent_style,omitempty"`
	IndentSize             string `ini:"indent_size" json:"indent_size,omitempty"`
	TabWidth               int    `ini:"tab_width" json:"tab_width,omitempty"`
	EndOfLine              string `ini:"end_of_line" json:"end_of_line,omitempty"`
	TrimTrailingWhitespace bool   `ini:"trim_trailing_whitespace" json:"trim_trailing_whitespace,omitempty"`
	InsertFinalNewline     bool   `ini:"insert_final_newline" json:"insert_final_newline,omitempty"`

	Raw map[string]string `ini:"-" json:"-"`
}

// Editorconfig represents a .editorconfig file.
// It is composed by a "root" property, plus the definitions defined in the
// file.
type Editorconfig struct {
	Root        bool
	Path        string
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
			raw  = make(map[string]string)
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

		// Shallow copy all properties
		for k, v := range iniSection.KeysHash() {
			raw[k] = v
		}

		definition.Selector = sectionStr
		definition.Raw = raw
		editorConfig.Definitions = append(editorConfig.Definitions, definition)
	}
	return editorConfig, nil
}

// ParseFile parses from a file.
func ParseFile(f string) (*Editorconfig, error) {
	abs, err := filepath.Abs(f)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(abs)
	if err != nil {
		return nil, err
	}

	ec, err := ParseBytes(data)
	if err != nil {
		return nil, err
	}
	ec.Path = abs
	return ec, nil
}

var (
	regexpBraces = regexp.MustCompile("^(|.*?[^\\\\])({[^}]*,[^}]*[^\\\\}]})(.*)$")
)

func filenameMatches(pattern, str string) bool {
	// Expand brace like {a,b,c}.js into a.js, b.js and c.js .
	// {single}.b should match "{single}.b"
	if braceMatch := regexpBraces.FindStringSubmatch(pattern); braceMatch != nil {
		candidates := strings.TrimPrefix(braceMatch[2], "{")
		candidates = strings.TrimSuffix(candidates, "}")

		for _, candidate := range strings.Split(candidates, ",") {
			newPattern := regexpBraces.ReplaceAllString(pattern, "${1}" + candidate + "${3}")
			matched := filenameMatches(newPattern, str)
			if matched {
				return true
			}
		}
		return false
	}

	// basic match
	matched := fnmatch.Match(pattern, str, 0)
	if matched {
		return true
	}
	// if singleBrace := regexpSingleBrace.FindString(pattern); len(singleBrace) > 0 {
	// 	// Replace { and } to \{ and \}
	// 	singleBrace = regexp.MustCompile("^{").ReplaceAllString(singleBrace, "\\{")
	// 	singleBrace = regexp.MustCompile("}$").ReplaceAllString(singleBrace, "\\}")

	// 	newPattern := regexpSingleBrace.ReplaceAllString(pattern, singleBrace)
	// 	matched = filenameMatches(newPattern, str)
	// 	if matched {
	// 		return true
	// 	}
	// }
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

	for k, v := range md.Raw {
		if _, ok := d.Raw[k]; !ok {
			d.Raw[k] = v
		}
	}
}

func (d *Definition) InsertToIniFile(iniFile *ini.File) {
	iniSec := iniFile.Section(d.Selector)
	for k, v := range d.Raw {
		iniSec.Key(k).SetValue(v)
	}
}

// GetDefinitionForFilename returns a definition for the given filename.
// The result is a merge of the selectors that matched the file.
// The last section has preference over the priors.
func (e *Editorconfig) GetDefinitionForFilename(name string) (*Definition, error) {
	abs, err := filepath.Abs(name)
	if err != nil {
		return nil, err
	}

	def := &Definition{}
	def.Raw = make(map[string]string)
	for i := len(e.Definitions) - 1; i >= 0; i-- {
		actualDef := e.Definitions[i]
		pattern := actualDef.Selector

		targetName := abs

		// If path separator not in pattern, just check file basename
		// Otherwise pattern is relative to .editorconfig base directory
		if p := strings.Index(pattern, string(filepath.Separator)); p == -1 {
			targetName = filepath.Base(targetName)
		} else {
			pattern = filepath.Join(filepath.Dir(e.Path), pattern)
		}

		if filenameMatches(pattern, targetName) {
			def.merge(actualDef)
		}
	}
	return def, nil
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// Serialize converts the Editorconfig to a slice of bytes, containing the
// content of the file in the INI format.
func (e *Editorconfig) Serialize() ([]byte, error) {
	var (
		iniFile = ini.Empty()
		buffer  = bytes.NewBuffer(nil)
	)
	iniFile.Section(ini.DEFAULT_SECTION).Comment = "http://editorconfig.org"
	if e.Root {
		iniFile.Section(ini.DEFAULT_SECTION).Key("root").SetValue(boolToString(e.Root))
	}
	for _, d := range e.Definitions {
		d.InsertToIniFile(iniFile)
	}
	_, err := iniFile.WriteTo(buffer)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Save saves the Editorconfig to a compatible INI file.
func (e *Editorconfig) Save(filename string) error {
	data, err := e.Serialize()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0666)
}

// GetDefinitionForFilename given a filename, searches
// for .editorconfig files, starting from the file folder,
// walking through the previous folders, until it reaches a
// folder with `root = true`, and returns the right editorconfig
// definition for the given file.
func GetDefinitionForFilename(filename string) (*Definition, error) {
	return GetDefinitionForFilenameWithConfigname(filename, ConfigNameDefault)
}

func GetDefinitionForFilenameWithConfigname(filename string, configname string) (*Definition, error) {
	abs, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	definition := &Definition{}
	definition.Raw = make(map[string]string)

	dir := abs
	for dir != filepath.Dir(dir) {
		dir = filepath.Dir(dir)
		ecFile := filepath.Join(dir, configname)
		if _, err := os.Stat(ecFile); os.IsNotExist(err) {
			continue
		}
		ec, err := ParseFile(ecFile)
		if err != nil {
			return nil, err
		}
		def, err := ec.GetDefinitionForFilename(filename)
		if err != nil {
			return nil, err
		}
		definition.merge(def)
		if ec.Root {
			break
		}
	}
	return definition, nil
}
