package editorconfig

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

const (
	testFile = "testdata/.editorconfig"
)

func TestParse(t *testing.T) {
	ec, err := ParseFile(testFile)
	if err != nil {
		t.Errorf("Couldn't parse file: %v", err)
	}

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
	assert.Equal(t, IdentStyleTab, def.IndentStyle)
	assert.Equal(t, "4", def.IndentSize)

	def = ec.Definitions[2]
	assert.Equal(t, "*.{js,css,less,htm,html}", def.Selector)
	assert.Equal(t, IdentStyleSpaces, def.IndentStyle)
	assert.Equal(t, "2", def.IndentSize)
}
