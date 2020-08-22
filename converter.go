package bdf2array

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"os"
	"sort"

	"github.com/zachomedia/go-bdf"
)

const maxBytesPerLine = 16
const maxFontWidth = 60
const maxFontHeight = 32

// Converter is the main struct for this library.
type Converter struct {
	Font             *bdf.Font      // BDF Font data
	Glyphs           map[int]*Glyph // The keys of this map are codepoint values. Only the requested codepoints are added.
	Encoding         []byte         // The binary encoding of the selected glyphs
	Annotations      map[int]string // Annotations. Key is the byte address (index in the encoded binary data)
	withDJTNumeric   bool           // Include a Direct Jump Table for numeric values
	withDJTUppercase bool           // Include a Direct Jump Table for uppercase values
	withDJTLowercase bool           // Include a Direct Jump Table for lowercase values
	fontLoaded       bool           // The font has been loaded
}

// NewConverter creates and returns a new Converter
func NewConverter() *Converter {
	c := &Converter{
		Glyphs: make(map[int]*Glyph),
	}
	return c
}

// LoadFontFromFile loads the specified file, which must contain uncompressed BDF font data.
func (c *Converter) LoadFontFromFile(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		err = f.Close()
	}()

	return c.LoadFont(f)
}

// LoadFont loads the font from the specified io.Reader.
func (c *Converter) LoadFont(input io.Reader) error {
	fontdata, err := ioutil.ReadAll(input)
	if err != nil {
		return err
	}
	if len(fontdata) == 0 {
		return fmt.Errorf("empty input font data")
	}

	c.Font, err = bdf.Parse(fontdata)
	if err != nil {
		return err
	}
	c.fontLoaded = true
	return nil
}

// IncludeCodepoint adds a codepoint to the set of the to-be-encoded glyphs.
func (c *Converter) IncludeCodepoint(codepoint int) (err error) {
	if !c.fontLoaded {
		return fmt.Errorf("no font loaded")
	}
	bdfChar, ok := c.Font.Encoding[rune(codepoint)]
	if !ok {
		return fmt.Errorf("skipping codepoint 0x%02X (not present in font)", codepoint)
	}

	g := &Glyph{
		Font:      c.Font,
		Codepoint: codepoint,
		Character: bdfChar,
		Alpha:     bdfChar.Alpha,
		Advance:   bdfChar.Advance[0],
	}
	g.scanImageRegion()

	err = g.encodeGlyph()
	if err != nil {
		return
	}

	c.Glyphs[codepoint] = g
	return nil
}

// SetIncludeDirectJumpTables sets the options to include or exclude the Direct Jump Tables
func (c *Converter) SetIncludeDirectJumpTables(
	includeDirectJumpTableNumeric bool, // include a DJT for numeric characters (digits)
	includeDirectJumpTableUppercase bool, // include a DJT for uppercase letters
	includeDirectJumpTableLowercase bool, // include a DJT for lowercase letters
) {
	c.withDJTNumeric = includeDirectJumpTableNumeric
	c.withDJTUppercase = includeDirectJumpTableUppercase
	c.withDJTLowercase = includeDirectJumpTableLowercase
}

// GenerateEncoding generates the encoding of the font, with the added codepoints included.
// Empty DJTs are always omitted.
// Annotations are indexed by byte offset in the end result.
// Empty annotation strings indicate, that that byte should start on a new line.
func (c *Converter) GenerateEncoding() (err error) {
	headerAnnotations := make(map[int]string)
	glyphAnnotations := make(map[int]string)
	fontHeader := new(bytes.Buffer)
	glyphData := new(bytes.Buffer)

	remainingGlyphs := make(map[int]*Glyph, len(c.Glyphs))
	for key, value := range c.Glyphs {
		remainingGlyphs[key] = value
	}

	// Omit requested DJTs if there are no characters available for them:
	includeDirectJumpTableNumeric := c.withDJTNumeric && c.codepointPresent('0', '9')
	includeDirectJumpTableUppercase := c.withDJTUppercase && c.codepointPresent('A', 'Z')
	includeDirectJumpTableLowercase := c.withDJTLowercase && c.codepointPresent('a', 'z')

	// Generate the header:
	err = c.generateFixedFontHeader(fontHeader, headerAnnotations, includeDirectJumpTableNumeric, includeDirectJumpTableUppercase, includeDirectJumpTableLowercase)
	if err != nil {
		return
	}
	if includeDirectJumpTableNumeric {
		err = c.generateDJT('0', '9', fontHeader, glyphData, remainingGlyphs, headerAnnotations, glyphAnnotations)
		if err != nil {
			return
		}
	}
	if includeDirectJumpTableUppercase {
		err = c.generateDJT('A', 'Z', fontHeader, glyphData, remainingGlyphs, headerAnnotations, glyphAnnotations)
		if err != nil {
			return
		}
	}
	if includeDirectJumpTableLowercase {
		err = c.generateDJT('a', 'z', fontHeader, glyphData, remainingGlyphs, headerAnnotations, glyphAnnotations)
		if err != nil {
			return
		}
	}
	if includeDirectJumpTableNumeric || includeDirectJumpTableUppercase || includeDirectJumpTableLowercase {
		// add jump address to the sequentially accessed glyphs:
		c.addAnnotation(headerAnnotations, fontHeader.Len(), "Jump address to sequentially addressed characters")
		jumpOffset := glyphData.Len()
		err = fontHeader.WriteByte(byte(jumpOffset)) // LSB
		if err != nil {
			return
		}
		err = fontHeader.WriteByte(byte(jumpOffset >> 8)) // nolint:gomnd
		if err != nil {
			return
		}
	}

	// add the remaining glyph data:
	err = c.encodeRemainingGlyphs(glyphData, remainingGlyphs, glyphAnnotations)
	if err != nil {
		return
	}

	// merge buffers and annotations:
	result, annotations, err := c.mergeHeaderAndGlyphData(fontHeader, glyphData, headerAnnotations, glyphAnnotations)
	if err != nil {
		return
	}

	c.Encoding = result.Bytes()
	c.Annotations = annotations
	return nil
}

// GenerateCStruct writes the encoded data as C Struct into 'output'.
// In 'declaration', specify everything that should appear before the '=' sign.
func (c *Converter) GenerateCStruct(output io.Writer, declaration string) (err error) {
	if len(c.Encoding) == 0 {
		err = c.GenerateEncoding()
		if err != nil {
			return
		}
	}

	_, err = output.Write([]byte(fmt.Sprintf(`// Font: %s
// Size: %d
// Ascent: %d
// Descent: %d
// Capheight: %d
// Characters: %d
// Bytes: %d
// Generated by bdf2array
`,
		c.Font.Name,
		c.Font.Size,
		c.Font.Ascent,
		c.Font.Descent,
		c.Font.CapHeight,
		len(c.Glyphs),
		len(c.Encoding),
	)))
	if err != nil {
		return
	}

	_, err = output.Write([]byte(declaration + " = {\n"))
	if err != nil {
		return
	}

	numOnLine := 0 // number of bytes already written to the current line
	separator := ""
	for address, encodedValue := range c.Encoding {
		err = c.printAnnotation(output, address, &numOnLine, &separator)
		if err != nil {
			return
		}

		_, err = output.Write([]byte(fmt.Sprintf("%s0x%02X", separator, encodedValue)))
		if err != nil {
			return
		}
		numOnLine++
		if numOnLine >= maxBytesPerLine {
			separator = ",\n"
			numOnLine = 0
		} else {
			separator = ","
		}
	}

	_, err = output.Write([]byte("\n};\n"))
	if err != nil {
		return
	}

	return nil
}

func (c *Converter) printAnnotation(output io.Writer, address int, numOnLine *int, separator *string) (err error) {
	text, ok := c.Annotations[address]
	if !ok {
		return
	}
	if *numOnLine > 0 {
		*separator += "\n"
	}
	_, err = output.Write([]byte(*separator + "// " + text + ":\n"))
	if err != nil {
		return
	}
	*separator = ""
	*numOnLine = 0
	return
}

// codepointPresent returns true if at least 1 codepoint between 'from' and 'to' (including) is present
func (c *Converter) codepointPresent(from, to int) bool {
	for codepoint := from; codepoint < to; codepoint++ {
		if _, ok := c.Glyphs[codepoint]; ok {
			return true
		}
	}
	return false
}

// generateFixedFontHeader generates the fixed part of the font header
func (c *Converter) generateFixedFontHeader(
	fontHeader *bytes.Buffer,
	headerAnnotations map[int]string,
	djtNumeric bool,
	djtUppercase bool,
	djtLowercase bool,
) (err error) {
	boundingBox := c.findBoundingBox()

	width := boundingBox.Dx()
	if width > maxFontWidth {
		return fmt.Errorf("bounding box width %d is too large", width)
	}
	c.addAnnotation(headerAnnotations, fontHeader.Len(), "bounding box width")
	err = fontHeader.WriteByte(byte(width)) // byte 0: bounding box width
	if err != nil {
		return
	}

	height := boundingBox.Dy()
	if height > maxFontHeight {
		return fmt.Errorf("bounding box height %d is too large", height)
	}
	c.addAnnotation(headerAnnotations, fontHeader.Len(), "bounding box height")
	err = fontHeader.WriteByte(byte(height)) // byte 1: bounding box height
	if err != nil {
		return
	}

	x := boundingBox.Min.X
	if x < 0 || x > maxFontWidth {
		return fmt.Errorf("bounding box X position %d out of range", x)
	}
	c.addAnnotation(headerAnnotations, fontHeader.Len(), "bounding box top-left X")
	err = fontHeader.WriteByte(byte(x)) // byte 2: bounding box top left X
	if err != nil {
		return
	}

	y := boundingBox.Min.Y
	if y < 0 || y > maxFontHeight {
		return fmt.Errorf("bounding box Y position %d out of range", y)
	}
	c.addAnnotation(headerAnnotations, fontHeader.Len(), "bounding box top-left Y")
	err = fontHeader.WriteByte(byte(y)) // byte 3: bounding box top left Y
	if err != nil {
		return
	}

	// byte 4: enabled DJTs:
	c.addAnnotation(headerAnnotations, fontHeader.Len(), "enabled Direct Jump Tables")
	var enabledDJTs byte
	if djtNumeric {
		enabledDJTs |= 0x01
	}
	if djtUppercase {
		enabledDJTs |= 0x02
	}
	if djtLowercase {
		enabledDJTs |= 0x04
	}
	err = fontHeader.WriteByte(enabledDJTs)
	if err != nil {
		return
	}
	return nil
}

// findBoundingBox returns the smallest bounding box that fits all selected glyphs
func (c *Converter) findBoundingBox() (result image.Rectangle) {
	for _, glyph := range c.Glyphs {
		result = result.Union(glyph.BoundingBox)
	}
	return
}

// generateDJT generates the glyphs in the specified range, and fills in their offsets in the DJT.
// This function adds data to fontHeader and glyphData, and removes the processed glyphs from remainingGlyphs.
func (c *Converter) generateDJT(
	from, to int,
	fontHeader *bytes.Buffer,
	glyphData *bytes.Buffer,
	remainingGlyphs map[int]*Glyph,
	headerAnnotations map[int]string,
	glyphAnnotations map[int]string,
) error {
	c.addAnnotation(headerAnnotations, fontHeader.Len(), fmt.Sprintf("Direct Jump Table 0x%02X-0x%02X ('%c'-'%c')", from, to, from, to))
	for codepoint := from; codepoint <= to; codepoint++ {
		jumpOffset := 0xFFFF
		if glyph, ok := remainingGlyphs[codepoint]; ok {
			jumpOffset = glyphData.Len()
			// add glyph data:
			c.addAnnotation(glyphAnnotations, glyphData.Len(), fmt.Sprintf("glyph data for code point 0x%02X ('%c')", codepoint, codepoint))
			encoding, err := glyph.getEncoding()
			if err != nil {
				return err
			}
			_, err = glyphData.Write(encoding)
			if err != nil {
				return err
			}
			delete(remainingGlyphs, codepoint)
		}
		// fill in the address in the DJT:
		err := fontHeader.WriteByte(byte(jumpOffset)) // LSB
		if err != nil {
			return err
		}
		err = fontHeader.WriteByte(byte(jumpOffset >> 8)) // nolint:gomnd
		if err != nil {
			return err
		}
	}
	return nil
}

// addAnnotation adds an annotation text.
// The map key is the byte address
// Multiple annotations for the same position are merged.
func (c *Converter) addAnnotation(annotations map[int]string, address int, text string) {
	if existing, ok := annotations[address]; ok {
		// key already exists
		if existing != "" {
			text = existing + " / " + text
		}
	}
	annotations[address] = text
}

func (c *Converter) encodeRemainingGlyphs(glyphData *bytes.Buffer, remainingGlyphs map[int]*Glyph, glyphAnnotations map[int]string) (err error) {
	// sort the remaining glyphs by codepoint:
	keys := make([]int, len(remainingGlyphs))
	i := 0
	for key := range remainingGlyphs {
		keys[i] = key
		i++
	}
	sort.Ints(keys)

	for _, key := range keys {
		glyph := remainingGlyphs[key]
		c.addAnnotation(glyphAnnotations, glyphData.Len(), fmt.Sprintf("glyph data for code point 0x%02X ('%c')", key, key))
		encoding, err := glyph.getEncoding()
		if err != nil {
			return err
		}
		_, err = glyphData.Write(encoding)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Converter) mergeHeaderAndGlyphData(
	fontHeader *bytes.Buffer,
	glyphData *bytes.Buffer,
	headerAnnotations map[int]string,
	glyphAnnotations map[int]string,
) (result *bytes.Buffer, annotations map[int]string, err error) {
	result = new(bytes.Buffer)

	// add the terminating zero glyph:
	c.addAnnotation(glyphAnnotations, glyphData.Len(), "glyph data terminator")
	_, err = glyphData.Write([]byte{0x00, 0x00})
	if err != nil {
		return
	}

	// merge the annotations:
	annotations = headerAnnotations // we can re-use these unaltered
	glyphOffset := fontHeader.Len()
	for key, value := range glyphAnnotations {
		annotations[key+glyphOffset] = value
	}

	// merge the data:
	_, err = result.Write(fontHeader.Bytes())
	if err != nil {
		return
	}
	_, err = result.Write(glyphData.Bytes())
	return
}
