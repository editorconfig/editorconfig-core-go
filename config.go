package editorconfig

import (
	"os"
	"path/filepath"

	"github.com/blang/semver"
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
	Parser  Parser
}

func (config *Config) Load(filename string) (*Definition, error) {
	// idiomatic go allows empty struct
	if config.Parser == nil {
		config.Parser = new(SimpleParser)
	}

	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	ecFile := config.Name
	if ecFile == "" {
		ecFile = ConfigNameDefault
	}

	definition := &Definition{}
	definition.Raw = make(map[string]string)

	if config.Version != "" {
		version, err := semver.New(config.Version)
		if err != nil {
			return nil, err
		}
		definition.version = version
	}

	dir := filename
	for dir != filepath.Dir(dir) {
		dir = filepath.Dir(dir)

		ini, err := config.Parser.ParseIni(filepath.Join(dir, ecFile))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		ec, err := newEditorconfig(ini)
		if err != nil {
			return nil, err
		}

		// give it the current config.
		ec.config = config

		relativeFilename := filename
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
