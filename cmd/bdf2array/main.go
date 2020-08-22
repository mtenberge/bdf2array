package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

// MandatoryOptions defines the mandatory command line options
type MandatoryOptions struct {
	Codepoints []string `short:"r" long:"range" required:"true" valuename:"range(s)" description:"Required: range(s) of codepoints to add, separated with commas. May be specified multiple times. Example: 1,2,0x20,0x30-0x7F"`
}

// OptionalOptions defines the optional command line options
type OptionalOptions struct {
	FontFile    string `short:"i" long:"input" default:"-" valuename:"filename" description:"BDF font file name to use as input"`
	OutputFile  string `short:"o" long:"output" default:"-" valuename:"filename" description:"output file for the generated C code"`
	HeadFile    string `short:"H" long:"headfile" valuename:"filename" description:"If specified, the contents of this file will be prepended to the generated output"`
	StructName  string `short:"s" long:"structname" default:"fontData" valuename:"C identifier" description:"name of the generated struct"`
	Declaration string `short:"d" long:"declaration" default:"const uint8_t <structname>[] PROGMEM" valuename:"C declaration" description:"The declaration of the C data struct. <structname> is substituted for the value specified with --structname."`
	OmitDJTs    bool   `long:"no-djt" description:"Omit all Direct Jump Tables from the generated output code"`
}

// CmdLineOptions defines the available command line options
type CmdLineOptions struct {
	MandatoryOptions `group:"Mandatory Options"`
	OptionalOptions  `group:"Optional Options"`
}

var cmdLineOptions CmdLineOptions
var parser = flags.NewParser(&cmdLineOptions, flags.HelpFlag|flags.PassDoubleDash)

func main() {
	_, err := parser.Parse()
	if err != nil {
		// print Usage information:
		fmt.Fprintf(os.Stderr, "\n")
		parser.WriteHelp(os.Stderr)

		// Print the error text and exit:
		fmt.Fprintf(os.Stderr, "\n\nError while parsing command line options: %s\n\n", err)
		os.Exit(1)
	}

	var numGenerated int
	numGenerated, err = generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\nError during conversion: %s\n\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Successfully generated %d codepoints\n", numGenerated)
}
