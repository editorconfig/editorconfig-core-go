package editorconfig

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

const (
	testFile = "testdata/.editorconfig"
)

func testParse(t *testing.T, ec *Editorconfig) {
	assert.Equal(t, true, ec.Root)
	assert.Equal(t, 3, len(ec.Definitions))

	def := ec.Definitions[0]
	assert.Equal(t, "*", def.Selector)
	assert.Equal(t, EndOfLineLf, def.EndOfLine)
	assert.Equal(t, true, def.InsertFinalNewline)
	assert.Equal(t, CharsetUTF8, def.Charset)
	assert.Equal(t, true, def.TrimTrailingWhitespace)

	def = ec.Definitions[1]
	assert.Equal(t, "*.go", def.Selector)
	assert.Equal(t, IndentStyleTab, def.IndentStyle)
	assert.Equal(t, "4", def.IndentSize)
	assert.Equal(t, 4, def.TabWidth)

	def = ec.Definitions[2]
	assert.Equal(t, "*.{js,css,less,htm,html}", def.Selector)
	assert.Equal(t, IndentStyleSpaces, def.IndentStyle)
	assert.Equal(t, "2", def.IndentSize)
	assert.Equal(t, 2, def.TabWidth)
}

func TestParse(t *testing.T) {
	ec, err := ParseFile(testFile)
	if err != nil {
		t.Errorf("Couldn't parse file: %v", err)
	}

	testParse(t, ec)
}

func TestFilenameMatches(t *testing.T) {
	assertFilenameMatch := func(pattern, name string) {
		assert.Equal(t, true, filenameMatches(pattern, name), "\"%s\" should match \"%s\"", name, pattern)
	}
	assertFilenameNotMatch := func(pattern, name string) {
		assert.Equal(t, false, filenameMatches(pattern, name), "\"%s\" should not match \"%s\"", name, pattern)
	}
	assertFilenameMatch("*", "main.go")
	assertFilenameMatch("*.go", "main.go")
	assertFilenameNotMatch("*.js", "main.go")
	assertFilenameMatch("main.go", "main.go")
	assertFilenameMatch("main.go", "foo/bar/main.go")
	assertFilenameMatch("foo/bar/main.go", "foo/bar/main.go")
	assertFilenameMatch("foo", "foo/main.go")

	assertFilenameMatch("*.{go,css}", "main.go")
	assertFilenameNotMatch("*.{js,css}", "main.go")
	assertFilenameMatch("*.{css,less}", "foo/bar/file.less")

	assertFilenameMatch("{foo}.go", "foo.go")
	assertFilenameMatch("{foo}.go", "bar/baz/foo.go")
	assertFilenameMatch("{}.go", "foo.go")
	assertFilenameMatch("{}.go", "bar.go")
	assertFilenameNotMatch("a{b,c}.go", "ad.go")
	assertFilenameMatch("a{b,c}.go", "ab.go")
	assertFilenameMatch("a{a,b,,d}.go", "ad.go")
	assertFilenameNotMatch("a{a,b,,d}.go", "ac.go")
	assertFilenameNotMatch("a{b,c,d}.go", "a.go")
	assertFilenameMatch("a{b,c,,d}.go", "a.go")
}

func TestGetDefinition(t *testing.T) {
	ec, err := ParseFile(testFile)
	if err != nil {
		t.Errorf("Couldn't parse file: %v", err)
	}

	def := ec.GetDefinitionForFilename("main.go")
	assert.Equal(t, IndentStyleTab, def.IndentStyle)
	assert.Equal(t, "4", def.IndentSize)
	assert.Equal(t, 4, def.TabWidth)
	assert.Equal(t, true, def.TrimTrailingWhitespace)
	assert.Equal(t, CharsetUTF8, def.Charset)
	assert.Equal(t, true, def.InsertFinalNewline)
	assert.Equal(t, EndOfLineLf, def.EndOfLine)
}

func TestSave(t *testing.T) {
	ec, err := ParseFile(testFile)
	if err != nil {
		t.Errorf("Couldn't parse file: %v", err)
	}

	tempFile := filepath.Join(os.TempDir(), ".editorconfig")
	defer os.Remove(tempFile)

	err = ec.Save(tempFile)
	assert.Nil(t, err)

	savedEc, err := ParseFile(tempFile)
	assert.Nil(t, err)

	testParse(t, savedEc)
}
