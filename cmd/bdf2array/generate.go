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
			err2 := fontFile.Close()
			if err == nil {
				err = err2
			}
		}()
	}

	err = c.LoadFont(fontFile)
	if err != nil {
		return
	}

	for _, codepoint := range codepoints {
		err = c.IncludeCodepoint(codepoint)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			err = nil // continue
		}
	}

	if len(c.Glyphs) == 0 {
		err = fmt.Errorf("none of the specified codepoints are present in the input font file")
		return
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
			err2 := headFile.Close()
			if err == nil {
				err = err2
			}
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
			err2 := outputFile.Close()
			if err == nil {
				err = err2
			}
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
