package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"

	"github.com/editorconfig/editorconfig-core-go/v2"
)

const (
	// Version indicates the current version number
	Version = "2.1.2"
)

func main() {
	var (
		configName      string
		configVersion   string
		showVersionFlag bool
	)
	flag.StringVar(&configName, "f", editorconfig.ConfigNameDefault, "Specify conf filename other than '.editorconfig'")
	flag.StringVar(&configVersion, "b", "", "Specify version (used by devs to test compatibility)")
	flag.BoolVar(&showVersionFlag, "v", false, "Display version information")
	flag.BoolVar(&showVersionFlag, "version", false, "Display version information")
	flag.Parse()

	if showVersionFlag {
		fmt.Printf("EditorConfig Core Go, Version %s\n", Version)
		os.Exit(0)
	}

	rest := flag.Args()

	if len(rest) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	for _, file := range rest {
		absolutePath, err := filepath.Abs(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		def, err := editorconfig.GetDefinitionForFilenameWithConfigname(absolutePath, configName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var (
			iniFile = ini.Empty()
			buffer  = bytes.NewBuffer(nil)
		)
		ini.PrettyFormat = false
		if len(rest) < 2 {
			def.Selector = ini.DEFAULT_SECTION
		} else {
			def.Selector = absolutePath
		}
		def.InsertToIniFile(iniFile)
		_, err = iniFile.WriteTo(buffer)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(buffer.String())
	}
}
