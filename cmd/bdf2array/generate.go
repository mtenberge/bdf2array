package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mtenberge/bdf2array"
)

func generate() (err error) {
	var argCodepoints intrange
	for _, cp := range cmdLineOptions.Codepoints {
		err = argCodepoints.Set(cp)
		if err != nil {
			return
		}
	}
	codepoints := argCodepoints.GetSlice()
	if len(codepoints) == 0 {
		err = fmt.Errorf("no codepoints specified")
		parser.WriteHelp(os.Stderr)
		return
	}

	c := bdf2array.NewConverter()

	// open input file:
	var fontFile *os.File
	if cmdLineOptions.FontFile == "-" {
		fontFile = os.Stdin
	} else {
		fontFile, err = os.Open(cmdLineOptions.FontFile)
		if err != nil {
			return
		}
		defer func() {
			err = fontFile.Close()
		}()
	}

	err = c.LoadFont(fontFile)
	if err != nil {
		return
	}

	for _, codepoint := range codepoints {
		err = c.IncludeCodepoint(codepoint)
		if err != nil {
			return
		}
	}

	// open head file:
	var headFile *os.File
	switch cmdLineOptions.HeadFile {
	case "":

	case "-":
		if cmdLineOptions.FontFile == "-" {
			err = fmt.Errorf("'input' and 'headfile' cannot both be read from '-' (stdin)")
			return
		}
		headFile = os.Stdin

	default:
		headFile, err = os.Open(cmdLineOptions.HeadFile)
		if err != nil {
			return
		}
		defer func() {
			err = headFile.Close()
		}()
	}

	// Open output file for writing:
	var outputFile *os.File
	if cmdLineOptions.OutputFile == "-" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.Create(cmdLineOptions.OutputFile)
		if err != nil {
			return
		}
		defer func() {
			err = outputFile.Close()
		}()
	}

	// copy headfile to the output:
	if headFile != nil {
		_, err = io.Copy(outputFile, headFile)
		if err != nil {
			return
		}
	}

	// for now, do not offer the possibility to toggle individual DJTs:
	djt := !cmdLineOptions.OmitDJTs
	c.SetIncludeDirectJumpTables(djt, djt, djt)

	// preprocess the declaration string:
	cmdLineOptions.Declaration = strings.ReplaceAll(cmdLineOptions.Declaration, "<structname>", cmdLineOptions.StructName)

	err = c.GenerateCStruct(outputFile, cmdLineOptions.Declaration)
	if err != nil {
		return
	}

	return nil
}
