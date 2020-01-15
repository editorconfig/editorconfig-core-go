package editorconfig

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/blang/semver"
	"gopkg.in/ini.v1"
)

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

// NewDefinition builds a definition from a given config
func NewDefinition(config Config) (*Definition, error) {
	return config.Load(config.Path)
}

// normalize fixes some values to their lowercaes value
func (d *Definition) normalize() error {
	d.Charset = strings.ToLower(d.Charset)
	d.EndOfLine = strings.ToLower(d.Raw["end_of_line"])
	d.IndentStyle = strings.ToLower(d.Raw["indent_style"])

	trimTrailingWhitespace, ok := d.Raw["trim_trailing_whitespace"]
	if ok && trimTrailingWhitespace != "unset" {
		trim, err := strconv.ParseBool(trimTrailingWhitespace)
		if err != nil {
			return fmt.Errorf("trim_trailing_whitespace=%s is not an acceptable value. %s", trimTrailingWhitespace, err)
		}
		d.TrimTrailingWhitespace = &trim
	}

	insertFinalNewline, ok := d.Raw["insert_final_newline"]
	if ok && insertFinalNewline != "unset" {
		insert, err := strconv.ParseBool(insertFinalNewline)
		if err != nil {
			return fmt.Errorf("insert_final_newline=%s is not an acceptable value. %s", insertFinalNewline, err)
		}
		d.InsertFinalNewline = &insert
	}

	// tab_width from Raw
	tabWidth, ok := d.Raw["tab_width"]
	if ok && tabWidth != "unset" {
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

// InsertToIniFile ... TODO
func (d *Definition) InsertToIniFile(iniFile *ini.File) {
	iniSec := iniFile.Section(d.Selector)
	for k, v := range d.Raw {
		if k == "insert_final_newline" {
			if d.InsertFinalNewline != nil {
				iniSec.NewKey(k, strconv.FormatBool(*d.InsertFinalNewline))
			} else {
				insertFinalNewline, ok := d.Raw["insert_final_newline"]
				if ok {
					iniSec.NewKey(k, strings.ToLower(insertFinalNewline))
				}
			}
		} else if k == "trim_trailing_whitespace" {
			if d.TrimTrailingWhitespace != nil {
				iniSec.NewKey(k, strconv.FormatBool(*d.TrimTrailingWhitespace))
			} else {
				trimTrailingWhitespace, ok := d.Raw["trim_trailing_whitespace"]
				if ok {
					iniSec.NewKey(k, strings.ToLower(trimTrailingWhitespace))
				}
			}
		} else if k == "charset" {
			iniSec.NewKey(k, d.Charset)
		} else if k == "end_of_line" {
			iniSec.NewKey(k, d.EndOfLine)
		} else if k == "indent_style" {
			iniSec.NewKey(k, d.IndentStyle)
		} else if k == "tab_width" {
			tabWidth, ok := d.Raw["tab_width"]
			if ok && tabWidth == "unset" {
				iniSec.NewKey(k, tabWidth)
			} else {
				iniSec.NewKey(k, strconv.Itoa(d.TabWidth))
			}
		} else if k == "indent_size" {
			iniSec.NewKey(k, d.IndentSize)
		} else {
			iniSec.NewKey(k, v)
		}
	}

	if _, ok := d.Raw["indent_size"]; !ok {
		tabWidth, ok := d.Raw["tab_width"]
		if ok && tabWidth == "unset" {
			// do nothing
		} else if d.TabWidth > 0 {
			iniSec.NewKey("indent_size", strconv.Itoa(d.TabWidth))
		} else if d.IndentStyle == IndentStyleTab && (d.version == nil || d.version.GTE(v0_9_0)) {
			iniSec.NewKey("indent_size", IndentStyleTab)
		}
	}

	if _, ok := d.Raw["tab_width"]; !ok {
		if d.IndentSize == "unset" {
			iniSec.NewKey("tab_width", d.IndentSize)
		} else {
			_, err := strconv.Atoi(d.IndentSize)
			if err == nil {
				iniSec.NewKey("tab_width", d.Raw["indent_size"])
			}
		}
	}
}
