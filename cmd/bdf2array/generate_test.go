package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testFont = "../../testdata/5x7.bdf"
const testOutput = "testdata/output.c"
const testHead = "testdata/head.c"
const testStructName = "fontData"
const testDeclaration = "const uint8_t <structname>[] PROGMEM"
const expectedWithHead = "testdata/expectedWithHead.c"
const expectedWithoutHead = "testdata/expectedWithoutHead.c"

func TestGenerateSuccessAllFiles(t *testing.T) {
	// prepare config:
	cmdLineOptions.Codepoints = []string{"0x30-0x32, 0x60"}
	cmdLineOptions.FontFile = testFont
	cmdLineOptions.OutputFile = testOutput
	cmdLineOptions.HeadFile = testHead
	cmdLineOptions.StructName = testStructName
	cmdLineOptions.Declaration = testDeclaration
	cmdLineOptions.OmitDJTs = false

	assert.NoError(t, generate())

	// compare the result:
	expected, err := ioutil.ReadFile(expectedWithHead)
	assert.NoError(t, err)
	actual, err := ioutil.ReadFile(testOutput)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestGenerateSuccessFilesWithoutHead(t *testing.T) {
	// prepare config:
	cmdLineOptions.Codepoints = []string{"0x30-0x32, 0x60"}
	cmdLineOptions.FontFile = testFont
	cmdLineOptions.OutputFile = testOutput
	cmdLineOptions.HeadFile = ""
	cmdLineOptions.StructName = testStructName
	cmdLineOptions.Declaration = testDeclaration
	cmdLineOptions.OmitDJTs = false

	assert.NoError(t, generate())

	// compare the result:
	expected, err := ioutil.ReadFile(expectedWithoutHead)
	assert.NoError(t, err)
	actual, err := ioutil.ReadFile(testOutput)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestGenerateSuccessFontFromStdin(t *testing.T) {
	// prepare config:
	cmdLineOptions.Codepoints = []string{"0x30-0x32, 0x60"}
	cmdLineOptions.FontFile = "-"
	cmdLineOptions.OutputFile = testOutput
	cmdLineOptions.HeadFile = testHead
	cmdLineOptions.StructName = testStructName
	cmdLineOptions.Declaration = testDeclaration
	cmdLineOptions.OmitDJTs = false

	// fake stdin:
	stdin, err := os.Open(testFont)
	assert.NoError(t, err)
	defer func() {
		err = stdin.Close()
		assert.NoError(t, err)
	}()

	origStdin := os.Stdin
	os.Stdin = stdin
	defer func() {
		os.Stdin = origStdin
	}()

	assert.NoError(t, generate())

	// compare the result:
	expected, err := ioutil.ReadFile(expectedWithHead)
	assert.NoError(t, err)
	actual, err := ioutil.ReadFile(testOutput)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestGenerateSuccessHeadFromStdin(t *testing.T) {
	// prepare config:
	cmdLineOptions.Codepoints = []string{"0x30-0x32, 0x60"}
	cmdLineOptions.FontFile = testFont
	cmdLineOptions.OutputFile = testOutput
	cmdLineOptions.HeadFile = "-"
	cmdLineOptions.StructName = testStructName
	cmdLineOptions.Declaration = testDeclaration
	cmdLineOptions.OmitDJTs = false

	// fake stdin:
	stdin, err := os.Open(testHead)
	assert.NoError(t, err)
	defer func() {
		err = stdin.Close()
		assert.NoError(t, err)
	}()

	origStdin := os.Stdin
	os.Stdin = stdin
	defer func() {
		os.Stdin = origStdin
	}()

	assert.NoError(t, generate())

	// compare the result:
	expected, err := ioutil.ReadFile(expectedWithHead)
	assert.NoError(t, err)
	actual, err := ioutil.ReadFile(testOutput)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestGenerateSuccessOutputToStdout(t *testing.T) {
	// prepare config:
	cmdLineOptions.Codepoints = []string{"0x30-0x32, 0x60"}
	cmdLineOptions.FontFile = testFont
	cmdLineOptions.OutputFile = "-"
	cmdLineOptions.HeadFile = testHead
	cmdLineOptions.StructName = testStructName
	cmdLineOptions.Declaration = testDeclaration
	cmdLineOptions.OmitDJTs = false

	outfile, err := os.Create(testOutput)
	assert.NoError(t, err)

	origStdout := os.Stdout
	os.Stdout = outfile
	defer func() {
		os.Stdout = origStdout
	}()

	assert.NoError(t, generate())

	assert.NoError(t, outfile.Close())

	// compare the result:
	expected, err := ioutil.ReadFile(expectedWithHead)
	assert.NoError(t, err)
	actual, err := ioutil.ReadFile(testOutput)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestGenerateErrorBothInputsFromStdin(t *testing.T) {
	// prepare config:
	cmdLineOptions.Codepoints = []string{"0x30-0x32, 0x60"}
	cmdLineOptions.FontFile = "-"
	cmdLineOptions.OutputFile = testOutput
	cmdLineOptions.HeadFile = "-"
	cmdLineOptions.StructName = testStructName
	cmdLineOptions.Declaration = testDeclaration
	cmdLineOptions.OmitDJTs = false

	// fake stdin:
	stdin, err := os.Open(testFont)
	assert.NoError(t, err)
	defer func() {
		err = stdin.Close()
		assert.NoError(t, err)
	}()

	origStdin := os.Stdin
	os.Stdin = stdin
	defer func() {
		os.Stdin = origStdin
	}()

	err = generate()

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "'input' and 'headfile' cannot both be read from '-' (stdin)")
	}
}

func TestGenerateErrorFontFileNotExist(t *testing.T) {
	// prepare config:
	cmdLineOptions.Codepoints = []string{"0x30-0x32, 0x60"}
	cmdLineOptions.FontFile = "non-existent-file.bdf"
	cmdLineOptions.OutputFile = testOutput
	cmdLineOptions.HeadFile = testHead
	cmdLineOptions.StructName = testStructName
	cmdLineOptions.Declaration = testDeclaration
	cmdLineOptions.OmitDJTs = false

	err := generate()

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "no such file or directory")
	}
}

func TestGenerateErrorInvalidCodepoints(t *testing.T) {
	// prepare config:
	cmdLineOptions.Codepoints = []string{"NaN"}
	cmdLineOptions.FontFile = testFont
	cmdLineOptions.OutputFile = testOutput
	cmdLineOptions.HeadFile = testHead
	cmdLineOptions.StructName = testStructName
	cmdLineOptions.Declaration = testDeclaration
	cmdLineOptions.OmitDJTs = false

	err := generate()

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "invalid syntax")
	}
}
