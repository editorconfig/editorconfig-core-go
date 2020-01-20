package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/ini.v1"

	"github.com/editorconfig/editorconfig-core-go/v2"
)

const (
	// Version indicates the current version number
	Version = "2.2.2"
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
		def, err := editorconfig.NewDefinition(editorconfig.Config{
			Path:    file,
			Name:    configName,
			Version: configVersion,
		})
		if err != nil {
			log.Fatal(err)
		}

		var (
			iniFile = ini.Empty()
		)
		ini.PrettyFormat = false
		if len(rest) < 2 {
			def.Selector = ini.DefaultSection
		} else {
			def.Selector = file
		}
		def.InsertToIniFile(iniFile)
		_, err = iniFile.WriteTo(os.Stdout)
		if err != nil {
			log.Fatal(err)
		}
	}
}
