# bdf2array
Converts BDF font data to a C array.

The output format is especially suited for small 1 bit-per-pixel displays attached to small microcontrollers, and can be rendered quickly, without a lot of CPU overhead (however, this is a subjective matter, others may think of this differently).

See `FORMAT.md` for a description of the supported encoding format.

## Usage
Currently, this module is only a library. A command-line tool will probably be added later.

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
