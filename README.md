# bdf2array
Converts BDF font data to a C array.

The output format is especially suited for small 1 bit-per-pixel displays attached to small microcontrollers, and can be rendered quickly, without a lot of CPU overhead (however, this is a subjective matter, others may think of this differently).

See `FORMAT.md` for a description of the supported encoding format.


## Usage
This module can be used as a library, and it has a command-line tool.


### As library
Follow these steps:

Import the library:
```go
import "github.com/mtenberge/bdf2array"
```

Create a Converter:
```go
c := bdf2array.NewConverter()
```

Load a font (must be a valid BDF font file):
```go
c.LoadFontFromFile("font.bdf")
```

Select the codepoints you want to include in the encoded output:
```go
c.IncludeCodepoint('a')
// or:
for i := 0x20; i < 0x7F; i++ {
  c.IncludeCodepoint(i)
}
```

Enable the Direct Jump Tables if you want to have them included in the generated output:
```go
c.SetIncludeDirectJumpTables(true, true, false)
```
These tables cost a little extra space (20 bytes for the digits and 52 bytes for upper-/lowercase each), but will make rendering of these often-used characters a lot faster. It shouldn't be too hard to implement this functionality in your font rendering code.

And generate C code:
```go
output := new(bytes.Buffer)
c.GenerateCStruct(output, "const static uint8_t fontData[] PROGMEM")
```
(of course you can also write to something else than a bytes.Buffer).

Don't forget to add error checking to these steps!


### Command-line tool
Either build or install the command-line tool:
- `cd cmd/bdf2array`
- `go build` (the executable will be placed in the local path) or `go install` (it will be added to you Go path)


#### Usage
Run `bdf2array -h` for an up-to-date overview of the available command line options.


#### Input font file
The input font file is mandatory. It can either be read from stdin (default setting, or when `-` is specified), or specified via the `-i` or `--input` option.


#### Input head file
This file is optional. If specified, the contents of this file will be written to the output file, before the generated code is added. You can use this feature for example to add extra #define, #include or other statements to the C file.


#### Output file
Specify the name of the output file with the `-o` or `--output` option. The default option (or `-`) writes the output to stdout.


#### Select codepoints
Specify which codepoints to include in the generated output, with the `-r` or `--range` option. This is mandatory.

Codepoints can be specified as individual numbers, lists (separated with a comma), ranges (from-to) or a combination. Numbers will be interpreted as decimal by default, hexadecimal when they are prefixed with `0x` or octal when they start with a zero. The option `-r` or `--range` can be specified more than once, to add additional codepoints.


#### Declaration
The generated code contains a declaration of the form:

```c
const uint8_t fontData[] PROGMEM = {
  /* Font data */
}
```

The part of the declaration before the `=`-sign can be changed with the `-d` or `--declaration` option. If the specified value contains the string `<structname>`, this string will be substituted with the value of the StructName option (see next section).

The default value is `const uint8_t <structname>[] PROGMEM`, which results in the code as shown in the example above.


#### StructName
Specify the name of the generated struct with the `-s` or `--structname` option. The default value is `fontData`.

This setting has no effect if a value is specified for the `-d` or `--declaration` option that does not contain the string `<structname>`.


#### Omit Direct Jump Tables
Optionally, specify the option flag `--no-djt` to omit all DJTs from the generated output.

Independent of this flag, empty DJTs will never be included in the generated output.
