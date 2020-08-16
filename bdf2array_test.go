package bdf2array

import (
	"bytes"
	"image"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestAll(t *testing.T) {
	suite.Run(t, new(testSuite))
}

type testSuite struct {
	suite.Suite
	c *Converter
}

func (suite *testSuite) SetupTest() {
	suite.c = NewConverter()
}

func (suite *testSuite) TestLoadErrorNonExistentFile() {
	suite.Error(suite.c.LoadFontFromFile("non-existent filename"))
}

func (suite *testSuite) TestGlyph5x7() {
	suite.NoError(suite.c.LoadFontFromFile("testdata/5x7.bdf"))
	suite.NoError(suite.c.IncludeCodepoint('!'))
	suite.NoError(suite.c.IncludeCodepoint('a'))

	// check some properties of the '!':
	suite.Equal("(0,0)-(5,7)", suite.c.Glyphs['!'].Alpha.Bounds().String())
	suite.Equal("(2,0)-(3,6)", suite.c.Glyphs['!'].BoundingBox.String())
	suite.Equal(5, suite.c.Glyphs['!'].Advance)
	suite.Equal(0, suite.c.Glyphs['!'].Character.LowerPoint[0])
	suite.Equal(-1, suite.c.Glyphs['!'].Character.LowerPoint[1])
	suite.Equal(image.Pt(2, 0), suite.c.Glyphs['!'].TopLeftOffset)
	if suite.Len(suite.c.Glyphs['!'].EncodedData, 1) {
		suite.Equal([]byte{0x2F}, suite.c.Glyphs['!'].EncodedData)
	}

	// check some properties of the 'a':
	suite.Equal("(0,0)-(5,7)", suite.c.Glyphs['a'].Alpha.Bounds().String())
	suite.Equal("(0,2)-(4,6)", suite.c.Glyphs['a'].BoundingBox.String())
	suite.Equal(5, suite.c.Glyphs['a'].Advance)
	suite.Equal(0, suite.c.Glyphs['a'].Character.LowerPoint[0])
	suite.Equal(-1, suite.c.Glyphs['a'].Character.LowerPoint[1])
	suite.Equal(image.Pt(0, 2), suite.c.Glyphs['a'].TopLeftOffset)
	if suite.Len(suite.c.Glyphs['a'].EncodedData, 4) {
		suite.Equal([]byte{0x06, 0x09, 0x05, 0x0f}, suite.c.Glyphs['a'].EncodedData)
	}
}

func (suite *testSuite) TestGlyph5x7Generate() {
	suite.NoError(suite.c.LoadFontFromFile("testdata/5x7.bdf"))
	for i := 32; i < 127; i++ {
		suite.NoError(suite.c.IncludeCodepoint(i))
	}

	suite.c.SetIncludeDirectJumpTables(true, true, true)

	// generate C code:
	output := new(bytes.Buffer)
	suite.NoError(suite.c.GenerateCStruct(output, "const static uint8_t fontData[] __attribute((\"PROGMEM\"))"))

	// compare the results:
	expected, err := ioutil.ReadFile("testdata/CStruct5x7.c")
	suite.NoError(err)
	suite.Equal(expected, output.Bytes())
}

func (suite *testSuite) TestGlyph6x13() {
	suite.NoError(suite.c.LoadFontFromFile("testdata/6x13.bdf"))
	suite.NoError(suite.c.IncludeCodepoint('!'))
	suite.NoError(suite.c.IncludeCodepoint('a'))

	// check some properties of the '!':
	suite.Equal("(0,0)-(6,13)", suite.c.Glyphs['!'].Alpha.Bounds().String())
	suite.Equal("(2,2)-(3,11)", suite.c.Glyphs['!'].BoundingBox.String())
	suite.Equal(6, suite.c.Glyphs['!'].Advance)
	suite.Equal(0, suite.c.Glyphs['!'].Character.LowerPoint[0])
	suite.Equal(-2, suite.c.Glyphs['!'].Character.LowerPoint[1])
	suite.Equal(image.Pt(2, 2), suite.c.Glyphs['!'].TopLeftOffset)
	if suite.Len(suite.c.Glyphs['!'].EncodedData, 2) {
		suite.Equal([]byte{0x7F, 0x01}, suite.c.Glyphs['!'].EncodedData)
	}

	// check some properties of the 'a':
	suite.Equal("(0,0)-(6,13)", suite.c.Glyphs['a'].Alpha.Bounds().String())
	suite.Equal("(0,5)-(5,11)", suite.c.Glyphs['a'].BoundingBox.String())
	suite.Equal(6, suite.c.Glyphs['a'].Advance)
	suite.Equal(0, suite.c.Glyphs['a'].Character.LowerPoint[0])
	suite.Equal(-2, suite.c.Glyphs['a'].Character.LowerPoint[1])
	suite.Equal(image.Pt(0, 5), suite.c.Glyphs['a'].TopLeftOffset)
	if suite.Len(suite.c.Glyphs['a'].EncodedData, 5) {
		suite.Equal([]byte{0x18, 0x25, 0x25, 0x15, 0x3e}, suite.c.Glyphs['a'].EncodedData)
	}
}

func (suite *testSuite) TestGlyph6x13Generate() {
	suite.NoError(suite.c.LoadFontFromFile("testdata/6x13.bdf"))
	for i := 32; i < 127; i++ {
		suite.NoError(suite.c.IncludeCodepoint(i))
	}

	suite.c.SetIncludeDirectJumpTables(true, true, true)

	// generate C code:
	output := new(bytes.Buffer)
	suite.NoError(suite.c.GenerateCStruct(output, "const static uint8_t fontData[] __attribute((\"PROGMEM\"))"))

	// compare the results:
	expected, err := ioutil.ReadFile("testdata/CStruct6x13.c")
	suite.NoError(err)
	suite.Equal(expected, output.Bytes())
}
