package bdf2array

import (
	"bytes"
	"fmt"
	"image"

	bdf "github.com/zachomedia/go-bdf"
)

const maxNumBytesPerColumn = 4
const glyphHeaderSize = 5

// Glyph contains data for a single glyph
type Glyph struct {
	Font           *bdf.Font
	Codepoint      int
	Character      *bdf.Character
	Alpha          *image.Alpha
	EncodedSize    int
	JumpOffset     int
	Advance        int
	BytesPerColumn int
	BoundingBox    image.Rectangle
	AlphaOffset    image.Point // offset between the font bounding box coordinates and the coordinate system of the Alpha image
	TopLeftOffset  image.Point
	EncodedData    []byte
}

func (g *Glyph) scanImageRegion() {
	g.BoundingBox = image.Rectangle{}
	bounds := g.Alpha.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if g.Alpha.AlphaAt(x, y).A != 0 {
				// active pixel, make sure it falls inside the bounding box:
				g.BoundingBox = g.BoundingBox.Union(image.Rectangle{
					Min: image.Pt(x, y),
					Max: image.Pt(x+1, y+1),
				})
			}
		}
	}

	// determine the offset between the coordinate system of the original Alpha image and the font bounding box:
	baseline := image.Pt(0, g.Font.Ascent)
	bottomLeft := baseline.Add(image.Pt(g.Character.LowerPoint[0], -g.Character.LowerPoint[1]))
	g.AlphaOffset = image.Pt(
		bottomLeft.X-g.Alpha.Bounds().Min.X,
		bottomLeft.Y-g.Alpha.Bounds().Max.Y,
	)
	if g.AlphaOffset.X < 0 || g.AlphaOffset.Y < 0 {
		fmt.Printf("Warning: Negative AlphaOffset found: %s, this should not happen\n", g.AlphaOffset)
	}
}

// EncodeGlyph creates the encoded data for the specified glyph.
func (g *Glyph) encodeGlyph() (err error) {
	if g.BoundingBox.Empty() {
		// empty bitmap (e.g. a space)
		g.EncodedData = make([]byte, 0) // empty data
		g.JumpOffset = glyphHeaderSize
		return // done
	}

	// calculate the offset of the top-left corner of the resulting bitmap data:
	g.TopLeftOffset = g.AlphaOffset.Add(g.BoundingBox.Min)

	// check the Advance value:
	if g.Advance < (g.TopLeftOffset.X + g.BoundingBox.Dx()) {
		fmt.Printf("minimum 'advance' value of character %d increased by %d pixels, to prevent horizontal overlap of glyphs\n", g.Codepoint, g.TopLeftOffset.X+g.BoundingBox.Dx()-g.Advance)
		g.Advance = g.TopLeftOffset.X + g.BoundingBox.Dx()
	}

	dy := g.BoundingBox.Dy()
	g.BytesPerColumn = ((dy - 1) / 8) + 1
	if g.BytesPerColumn > maxNumBytesPerColumn {
		return fmt.Errorf("character %d too tall (%d exceeds the maximum of 32)", g.Codepoint, dy)
	}

	bitmap := &bytes.Buffer{}

	for x := g.BoundingBox.Min.X; x < g.BoundingBox.Max.X; x++ {
		y := g.BoundingBox.Min.Y
		for row := 0; row < g.BytesPerColumn; row++ {
			var val uint8
			for bit := 0; bit < 8; bit++ {
				if g.Alpha.AlphaAt(x, y).A != 0 { // reading outside the image's bounding box always returns 0, so no special measures are needed here
					val |= 1 << bit
				}
				y++
			}
			err = bitmap.WriteByte(val)
			if err != nil {
				return
			}
		}
	}

	g.EncodedData = bitmap.Bytes()
	g.JumpOffset = glyphHeaderSize + len(g.EncodedData)

	return nil
}

func (g *Glyph) getEncoding() (result []byte, err error) {
	result = make([]byte, g.JumpOffset)

	if g.Codepoint > 255 || g.JumpOffset > 255 || g.TopLeftOffset.Y > 31 || g.TopLeftOffset.Y > 255 || g.Advance > 255 {
		err = fmt.Errorf("range exceeded on codepoint %d", g.Codepoint)
		return
	}
	// fill in the header:
	result[0] = byte(g.Codepoint)
	result[1] = byte(g.JumpOffset)
	result[2] = byte(g.TopLeftOffset.Y)<<2 | (byte(g.BytesPerColumn-1) & 0x03) // nolint:gomnd
	result[3] = byte(g.TopLeftOffset.X)
	result[4] = byte(g.Advance)
	// and copy the data:
	copy(result[5:], g.EncodedData)
	return
}
