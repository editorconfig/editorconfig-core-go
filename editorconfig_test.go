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
	assertFilenameNotMatch("main.go", "foo/bar/main.go")
	assertFilenameMatch("foo/bar/main.go", "foo/bar/main.go")
	assertFilenameMatch("/foo/bar/main.go", "/foo/bar/main.go")
	assertFilenameNotMatch("foo", "foo/main.go")
	assertFilenameMatch("*/main.go", "foo/main.go")
	assertFilenameMatch("*/*.go", "foo/main.go")
	assertFilenameMatch("**/*.go", "foo/bar/main.go")
	assertFilenameMatch("/foo/**/*.go", "/foo/bar/baz/main.go")

	assertFilenameMatch("*.{go,css}", "main.go")
	assertFilenameNotMatch("*.{js,css}", "main.go")
	assertFilenameMatch("*.{css,less}", "file.less")
	assertFilenameMatch("{file,a}.{css,less}", "file.less")
	assertFilenameNotMatch("{file,a}.{css,less}", "file.html")

	assertFilenameMatch("{foo}.go", "{foo}.go")
	assertFilenameNotMatch("{foo}.go", "bar/baz/foo.go")
	assertFilenameNotMatch("{}.go", "foo.go")
	assertFilenameNotMatch("{}.go", "bar.go")
	assertFilenameNotMatch("a{b,c}.go", "ad.go")
	assertFilenameMatch("a{b,c}.go", "ab.go")
	assertFilenameMatch("a{a,b,,d}.go", "ad.go")
	assertFilenameNotMatch("a{a,b,,d}.go", "ac.go")
	assertFilenameNotMatch("a{b,c,d}.go", "a.go")
	assertFilenameMatch("a{b,c,,d}.go", "a.go")

	assertFilenameMatch("a{1..3}.go", "a2.go")
	assertFilenameNotMatch("a{1..3}.go", "a4.go")
	assertFilenameMatch("a{-3..3}.go", "a-2.go")
	assertFilenameMatch("a{-3..3}.{go,css}", "a2.go")

	assertFilenameMatch("[abc].js", "b.js")
	assertFilenameMatch("[abc]b.js", "ab.js")
	assertFilenameMatch("a[a-d].go", "ac.go")
	assertFilenameMatch("a[a-d].go", "ac.go")
	assertFilenameMatch("a[a-d].go", "ac.go")
	assertFilenameMatch("[abd-g].go", "e.go")
	assertFilenameMatch("[!abc].js", "d.js")
	assertFilenameMatch("[!abc]b.js", "db.js")
	assertFilenameMatch("/dir/[!abc].js", "/dir/f.js")
	assertFilenameMatch("[!a-c].js", "d.js")

	assertFilenameNotMatch("[!abc].js", "b.js")
	assertFilenameNotMatch("[!abc]b.js", "bb.js")
	assertFilenameNotMatch("[!abc]b.js", "ab.js")
	assertFilenameNotMatch("/dir/[!abc].js", "/dir/a.js")
	assertFilenameNotMatch("a[a-d].go", "af.go")
	assertFilenameNotMatch("[!a-c].js", "a.js")


}

func TestGetDefinition(t *testing.T) {
	ec, err := ParseFile(testFile)
	if err != nil {
		t.Errorf("Couldn't parse file: %v", err)
	}

	def, err := ec.GetDefinitionForFilename("testdata/main.go")
	assert.Nil(t, err)
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

func TestPublicTestDefinitionForFilename(t *testing.T) {
	def, err := GetDefinitionForFilename("testdata/root/src/dummy.go")
	if err != nil {
		t.Errorf("Couldn't get file definition: %v", err)
	}
	assert.Nil(t, err)
	assert.Equal(t, "4", def.IndentSize)
	assert.Equal(t, IndentStyleTab, def.IndentStyle)
	assert.Equal(t, true, def.InsertFinalNewline)
	assert.Equal(t, false, def.TrimTrailingWhitespace)
}

func TestPublicTestDefinitionForFilenameWithConfigname(t *testing.T) {
	def, err := GetDefinitionForFilenameWithConfigname("testdata/root/src/dummy.go", "a.ini")
	assert.Nil(t, err)
	assert.Equal(t, "5", def.IndentSize)
	assert.Equal(t, IndentStyleSpaces, def.IndentStyle)
	assert.Equal(t, false, def.InsertFinalNewline)
	assert.Equal(t, false, def.TrimTrailingWhitespace)
}
