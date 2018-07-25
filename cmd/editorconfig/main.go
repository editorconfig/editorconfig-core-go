package main

import (
	"bytes"
	"path/filepath"
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/ini.v1"

	editorconfig "github.com/editorconfig/editorconfig-core-go"
)

const (
	Version = "1.2.0"
)

func main() {
	var (
		configname string
		configversion string
		showVersionFlag bool
	)
	flag.StringVar(&configname, "f", editorconfig.ConfigNameDefault, "Specify conf filename other than '.editorconfig'")
	flag.StringVar(&configversion, "b", "", "Specify version (used by devs to test compatibility)")
	flag.BoolVar(&showVersionFlag, "v", false, "Display version information")
	flag.BoolVar(&showVersionFlag, "version", false, "Display version information")
	flag.Parse()

	if showVersionFlag {
		fmt.Printf("EditorConfig Core Go, Version %s\n", Version)
		os.Exit(0)
	}

	rest := flag.Args()

	for _, file := range rest {
		if !filepath.IsAbs(file) {
			fmt.Fprintf(os.Stderr, "Input file must be a full path name: %s\n", file)
			os.Exit(1)
		}
		def, err := editorconfig.GetDefinitionForFilenameWithConfigname(file, configname)
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
			def.Selector = file
		}
		def.InsertToIniFile(iniFile)
		_, err = iniFile.WriteTo(buffer)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", buffer.String())
	}
}
