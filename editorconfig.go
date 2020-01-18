package editorconfig

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/blang/semver"
	"gopkg.in/ini.v1"
)

const (
	// ConfigNameDefault represents the name of the configuration file
	ConfigNameDefault = ".editorconfig"
	// UnsetValue is the value that unsets a preexisting variable
	UnsetValue = "unset"
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
	CharsetUTF8BOM = "utf-8 bom"
)

// Limits for section name, properties, and values.
const (
	MaxPropertyLength = 50
	MaxSectionLength  = 4096
	MaxValueLength    = 255
)

var (
	v0_9_0 = semver.Version{
		Major: 0,
		Minor: 9,
		Patch: 0,
	}
)

// Config holds the configuration
type Config struct {
	Path    string
	Name    string
	Version string
}

// NewDefinition builds a definition from a given config
func NewDefinition(config Config) (*Definition, error) {
	if config.Name == "" {
		config.Name = ConfigNameDefault
	}

	abs, err := filepath.Abs(config.Path)
	if err != nil {
		return nil, err
	}

	config.Path = abs

	return newDefinition(config)
}

// newDefinition recursively builds the definition
func newDefinition(config Config) (*Definition, error) {
	definition := &Definition{}
	definition.Raw = make(map[string]string)

	if config.Version != "" {
		version, err := semver.New(config.Version)
		if err != nil {
			return nil, err
		}
		definition.version = version
	}

	dir := config.Path
	for dir != filepath.Dir(dir) {
		dir = filepath.Dir(dir)
		ecFile := filepath.Join(dir, config.Name)
		fp, err := os.Open(ecFile)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			continue
		}
		defer fp.Close()
		ec, err := Parse(fp)
		if err != nil {
			return nil, err
		}

		relativeFilename := config.Path
		if len(dir) < len(relativeFilename) {
			relativeFilename = relativeFilename[len(dir):]
		}

		def, err := ec.GetDefinitionForFilename(relativeFilename)
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

// Definition represents a definition inside the .editorconfig file.
// E.g. a section of the file.
// The definition is composed of the selector ("*", "*.go", "*.{js.css}", etc),
// plus the properties of the selected files.
type Definition struct {
	Selector string `ini:"-" json:"-"`

	Charset                string            `ini:"charset" json:"charset,omitempty"`
	IndentStyle            string            `ini:"indent_style" json:"indent_style,omitempty"`
	IndentSize             string            `ini:"indent_size" json:"indent_size,omitempty"`
	TabWidth               int               `ini:"-" json:"-"`
	EndOfLine              string            `ini:"end_of_line" json:"end_of_line,omitempty"`
	TrimTrailingWhitespace *bool             `ini:"-" json:"-"`
	InsertFinalNewline     *bool             `ini:"-" json:"-"`
	Raw                    map[string]string `ini:"-" json:"-"`
	version                *semver.Version
}

// Editorconfig represents a .editorconfig file.
// It is composed by a "root" property, plus the definitions defined in the
// file.
type Editorconfig struct {
	Root        bool
	Definitions []*Definition
}

// Parse parses from a reader.
func Parse(r io.Reader) (*Editorconfig, error) {
	iniFile, err := ini.Load(r)
	if err != nil {
		return nil, err
	}

	return newEditorconfig(iniFile)
}

// ParseBytes parses from a slice of bytes.
//
// Deprecated: use Parse instead.
func ParseBytes(data []byte) (*Editorconfig, error) {
	iniFile, err := ini.Load(data)
	if err != nil {
		return nil, err
	}

	return newEditorconfig(iniFile)
}

// ParseFile parses from a file.
//
// Deprecated: use Parse instead.
func ParseFile(path string) (*Editorconfig, error) {
	iniFile, err := ini.Load(path)
	if err != nil {
		return nil, err
	}

	return newEditorconfig(iniFile)
}

// newEditorconfig builds the configuration from an INI file.
func newEditorconfig(iniFile *ini.File) (*Editorconfig, error) {
	editorConfig := &Editorconfig{}

	// Consider mixed-case values for true and false.
	rootKey := iniFile.Section(ini.DefaultSection).Key("root")
	rootKey.SetValue(strings.ToLower(rootKey.Value()))
	editorConfig.Root = rootKey.MustBool(false)

	for _, sectionStr := range iniFile.SectionStrings() {
		if sectionStr == ini.DefaultSection || len(sectionStr) > MaxSectionLength {
			continue
		}

		iniSection := iniFile.Section(sectionStr)
		definition := &Definition{}
		err := iniSection.MapTo(&definition)
		if err != nil {
			return nil, err
		}

		// Shallow copy all properties
		raw := make(map[string]string)
		for k, v := range iniSection.KeysHash() {
			if len(k) > MaxPropertyLength || len(v) > MaxValueLength {
				continue
			}
			raw[strings.ToLower(k)] = v
		}
		definition.Raw = raw

		definition.Selector = sectionStr
		if err := definition.normalize(); err != nil {
			return nil, err
		}
		editorConfig.Definitions = append(editorConfig.Definitions, definition)
	}

	return editorConfig, nil
}

// normalize fixes some values to their lowercaes value
func (d *Definition) normalize() error {
	d.Charset = strings.ToLower(d.Charset)
	d.EndOfLine = strings.ToLower(d.Raw["end_of_line"])
	d.IndentStyle = strings.ToLower(d.Raw["indent_style"])

	trimTrailingWhitespace, ok := d.Raw["trim_trailing_whitespace"]
	if ok && trimTrailingWhitespace != UnsetValue {
		trim, err := strconv.ParseBool(trimTrailingWhitespace)
		if err != nil {
			return fmt.Errorf("trim_trailing_whitespace=%s is not an acceptable value. %s", trimTrailingWhitespace, err)
		}
		d.TrimTrailingWhitespace = &trim
	}

	insertFinalNewline, ok := d.Raw["insert_final_newline"]
	if ok && insertFinalNewline != UnsetValue {
		insert, err := strconv.ParseBool(insertFinalNewline)
		if err != nil {
			return fmt.Errorf("insert_final_newline=%s is not an acceptable value. %s", insertFinalNewline, err)
		}
		d.InsertFinalNewline = &insert
	}

	// tab_width from Raw
	tabWidth, ok := d.Raw["tab_width"]
	if ok && tabWidth != UnsetValue {
		num, err := strconv.Atoi(tabWidth)
		if err != nil {
			return fmt.Errorf("tab_width=%s is not an acceptable value. %s", tabWidth, err)
		}
		d.TabWidth = num
	}

	// tab_width defaults to indent_size:
	// https://github.com/editorconfig/editorconfig/wiki/EditorConfig-Properties#tab_width
	num, err := strconv.Atoi(d.IndentSize)
	if err == nil && d.TabWidth <= 0 {
		d.TabWidth = num
	}

	return nil
}

// merge the parent definition into the child definition
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
	if trimTrailingWhitespace, ok := d.Raw["trim_trailing_whitespace"]; !ok || trimTrailingWhitespace != "unset" {
		if d.TrimTrailingWhitespace == nil {
			d.TrimTrailingWhitespace = md.TrimTrailingWhitespace
		}
	}
	if insertFinalNewline, ok := d.Raw["insert_final_newline"]; !ok || insertFinalNewline != "unset" {
		if d.InsertFinalNewline == nil {
			d.InsertFinalNewline = md.InsertFinalNewline
		}
	}

	for k, v := range md.Raw {
		if _, ok := d.Raw[k]; !ok {
			d.Raw[k] = v
		}
	}
}

func setValues(d *Definition, iniSection *ini.Section, key string, value string) {
	switch key {
	case "insert_final_newline":
		if d.InsertFinalNewline != nil {
			iniSection.NewKey(key, strconv.FormatBool(*d.InsertFinalNewline)) // nolint:errcheck
		} else {
			insertFinalNewline, ok := d.Raw["insert_final_newline"]
			if ok {
				iniSection.NewKey(key, strings.ToLower(insertFinalNewline)) // nolint: errcheck
			}
		}
	case "trim_trailing_whitespace":
		if d.TrimTrailingWhitespace != nil {
			iniSection.NewKey(key, strconv.FormatBool(*d.TrimTrailingWhitespace)) // nolint:errcheck
		} else {
			trimTrailingWhitespace, ok := d.Raw["trim_trailing_whitespace"]
			if ok {
				iniSection.NewKey(key, strings.ToLower(trimTrailingWhitespace)) // nolint:errcheck
			}
		}
	case "charset":
		iniSection.NewKey(key, d.Charset) // nolint:errcheck
	case "end_of_line":
		iniSection.NewKey(key, d.EndOfLine) // nolint:errcheck
	case "indent_style":
		iniSection.NewKey(key, d.IndentStyle) // nolint:errcheck
	case "tab_width":
		tabWidth, ok := d.Raw["tab_width"]
		if ok && tabWidth == UnsetValue {
			iniSection.NewKey(key, tabWidth) // nolint:errcheck
		} else {
			iniSection.NewKey(key, strconv.Itoa(d.TabWidth)) // nolint:errcheck
		}
	case "indent_size":
		iniSection.NewKey(key, d.IndentSize) // nolint:errcheck
	default:
		iniSection.NewKey(key, value) // nolint:errcheck
	}
}

func setRawValues(d *Definition, iniSection *ini.Section) {
	if _, ok := d.Raw["indent_size"]; !ok {
		tabWidth, ok := d.Raw["tab_width"]
		switch {
		case ok && tabWidth == UnsetValue:
			// do nothing
		case d.TabWidth > 0:
			iniSection.NewKey("indent_size", strconv.Itoa(d.TabWidth)) // nolint:errcheck
		case d.IndentStyle == IndentStyleTab && (d.version == nil || d.version.GTE(v0_9_0)):
			iniSection.NewKey("indent_size", IndentStyleTab) // nolint:errcheck
		}
	}

	if _, ok := d.Raw["tab_width"]; !ok {
		if d.IndentSize == UnsetValue {
			iniSection.NewKey("tab_width", d.IndentSize) // nolint:errcheck
		} else {
			_, err := strconv.Atoi(d.IndentSize)
			if err == nil {
				iniSection.NewKey("tab_width", d.Raw["indent_size"]) // nolint:errcheck
			}
		}
	}
}

// InsertToIniFile ... TODO
func (d *Definition) InsertToIniFile(iniFile *ini.File) {
	iniSec := iniFile.Section(d.Selector)
	for k, v := range d.Raw {
		setValues(d, iniSec, k, v)
	}

	setRawValues(d, iniSec)
}

// GetDefinitionForFilename returns a definition for the given filename.
// The result is a merge of the selectors that matched the file.
// The last section has preference over the priors.
func (e *Editorconfig) GetDefinitionForFilename(name string) (*Definition, error) {
	def := &Definition{}
	def.Raw = make(map[string]string)
	for i := len(e.Definitions) - 1; i >= 0; i-- {
		actualDef := e.Definitions[i]
		selector := actualDef.Selector
		if !strings.HasPrefix(selector, "/") {
			if strings.ContainsRune(selector, '/') {
				selector = "/" + selector
			} else {
				selector = "/**/" + selector
			}
		}
		if !strings.HasPrefix(name, "/") {
			name = "/" + name
		}
		ok, err := FnmatchCase(selector, name)
		if err != nil {
			return nil, err
		}
		if ok {
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
	buffer := bytes.NewBuffer(nil)
	err := e.Write(buffer)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Write writes the Editorconfig to the Writer in a compatible INI file.
func (e *Editorconfig) Write(w io.Writer) error {
	var (
		iniFile = ini.Empty()
	)
	iniFile.Section(ini.DefaultSection).Comment = "https://editorconfig.org"
	if e.Root {
		iniFile.Section(ini.DefaultSection).Key("root").SetValue(boolToString(e.Root))
	}
	for _, d := range e.Definitions {
		d.InsertToIniFile(iniFile)
	}
	_, err := iniFile.WriteTo(w)
	return err
}

// Save saves the Editorconfig to a compatible INI file.
func (e *Editorconfig) Save(filename string) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return e.Write(f)
}

// GetDefinitionForFilename given a filename, searches
// for .editorconfig files, starting from the file folder,
// walking through the previous folders, until it reaches a
// folder with `root = true`, and returns the right editorconfig
// definition for the given file.
func GetDefinitionForFilename(filename string) (*Definition, error) {
	return NewDefinition(Config{
		Path: filename,
	})
}

// GetDefinitionForFilenameWithConfigname given a filename and a configname,
// searches for configname files, starting from the file folder,
// walking through the previous folders, until it reaches a
// folder with `root = true`, and returns the right editorconfig
// definition for the given file.
func GetDefinitionForFilenameWithConfigname(filename string, configname string) (*Definition, error) {
	return NewDefinition(Config{
		Path: filename,
		Name: configname,
	})
}
