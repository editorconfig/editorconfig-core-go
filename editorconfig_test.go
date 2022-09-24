package editorconfig //nolint:testpackage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/iotest"

	"github.com/hashicorp/go-multierror"

	"github.com/editorconfig/editorconfig-core-go/v2/internal/assert"
)

const (
	testFile = "testdata/.editorconfig"
)

func testParse(t *testing.T, ec *Editorconfig) {
	t.Helper()

	assert.Equal(t, true, ec.Root)
	assert.Equal(t, 3, len(ec.Definitions))

	def := ec.Definitions[0]
	assert.Equal(t, "*", def.Selector)
	assert.Equal(t, EndOfLineLf, def.EndOfLine)
	assert.Equal(t, true, *def.InsertFinalNewline)
	assert.Equal(t, CharsetUTF8, def.Charset)
	assert.Equal(t, true, *def.TrimTrailingWhitespace)
	assert.Equal(t, "8", def.IndentSize)

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

func TestParseFile(t *testing.T) {
	t.Parallel()

	ec, err := ParseFile(testFile)
	assert.Nil(t, err)

	testParse(t, ec)
}

func TestParseBytes(t *testing.T) { //nolint:paralleltest
	data, err := os.ReadFile(testFile)
	assert.Nil(t, err)

	ec, err := ParseBytes(data)
	assert.Nil(t, err)

	testParse(t, ec)
}

func TestParseReader(t *testing.T) { //nolint:paralleltest
	f, err := os.Open(testFile)
	assert.Nil(t, err)

	defer f.Close()

	ec, err := Parse(f)
	assert.Nil(t, err)

	testParse(t, ec)
}

func TestParseReaderTimeoutError(t *testing.T) { //nolint:paralleltest
	f, err := os.Open(testFile)
	assert.Nil(t, err)

	defer f.Close()

	_, err = Parse(iotest.TimeoutReader(f))
	assert.Equal(t, true, errors.Is(err, iotest.ErrTimeout))
}

func TestGetDefinition(t *testing.T) {
	t.Parallel()

	ec, err := ParseFile(testFile)
	if err != nil {
		t.Errorf("Couldn't parse file: %v", err)
	}

	def, err := ec.GetDefinitionForFilename("main.go")
	if err != nil {
		t.Errorf("Couldn't parse file: %v", err)
	}

	assert.Equal(t, IndentStyleTab, def.IndentStyle)
	assert.Equal(t, "4", def.IndentSize)
	assert.Equal(t, 4, def.TabWidth)
	assert.Equal(t, true, *def.TrimTrailingWhitespace)
	assert.Equal(t, CharsetUTF8, def.Charset)
	assert.Equal(t, true, *def.InsertFinalNewline)
	assert.Equal(t, EndOfLineLf, def.EndOfLine)
}

func TestWrite(t *testing.T) { //nolint:paralleltest
	ec, err := ParseFile(testFile)
	if err != nil {
		t.Errorf("Couldn't parse file: %v", err)
	}

	tempFile := filepath.Join(os.TempDir(), ".editorconfig")

	f, err := os.OpenFile(tempFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	assert.Nil(t, err)

	defer func() {
		f.Close()
		os.Remove(tempFile)
	}()

	err = ec.Write(f)
	assert.Nil(t, err)

	savedEc, err := ParseFile(tempFile)
	assert.Nil(t, err)

	testParse(t, savedEc)
}

func TestSave(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	def, err := GetDefinitionForFilename("testdata/root/src/dummy.go")
	assert.Nil(t, err)
	assert.Equal(t, "4", def.IndentSize)
	assert.Equal(t, IndentStyleTab, def.IndentStyle)
	assert.Equal(t, true, *def.InsertFinalNewline)
	assert.Equal(t, (*bool)(nil), def.TrimTrailingWhitespace)
}

func TestPublicTestDefinitionForFilenameWithConfigname(t *testing.T) {
	t.Parallel()

	def, warning, err := GetDefinitionForFilenameWithConfignameGraceful("testdata/root/src/dummy.go", "a.ini")

	if merr, ok := warning.(*multierror.Error); ok { //nolint:errorlint
		// 3 errors are expected
		assert.Equal(t, 3, len(merr.Errors))
	}

	assert.Nil(t, err)
	assert.Equal(t, "5", def.IndentSize)
	assert.Equal(t, IndentStyleSpaces, def.IndentStyle)
	assert.Equal(t, false, *def.InsertFinalNewline)
	assert.Equal(t, false, *def.TrimTrailingWhitespace)
}
