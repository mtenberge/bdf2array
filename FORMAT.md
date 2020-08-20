# Format of the exported font data
This library currently supports one encoding method, 'Byte-row encoding'.

For an other encoding method, and an extended graphics library, see https://github.com/olikraus/u8g2.

## Byte-row encoding
This encoding is optimized for monochrome displays (such as OLED displays), where the pixels are addressed in horizontal rows having a height of 8 pixels (i.e. each row is 1 byte high, LSB is on top), and where the microcontroller does not have video buffer memory.

It has the following properties:
- proportional width glyphs
- pixel data is stored in column-first format
- each pixel column starts on a new byte (easier decoding, without per-bit data processing and carry-over, at the cost of slightly higher storage size)
- per-codepoint variable vertical storage size (rounded up to the next multiple of 8 bit)
- direct jump tables for fast access to lowercase, uppercase and numeric characters
- Only 8-bit codepoints supported (e.g. 7-bit ASCII plus some additional characters)
- Missing/skipped character encodings allowed
- Max 60(w) x 32(h) character size
- Origin coordinate (x=0,y=0) is in top-left corner

### Font header

Byte nr  | Size        | Data
---------|-------------|--------------------------------------------------------------------------------------------
0        | 1 byte      | font bounding box width
1        | 1 byte      | font bounding box height
2        | 1 byte      | font bounding box top-left corner, X
3        | 1 byte      | font bounding box top-left corner, Y
4, bit 0 | 1 bit       | direct jump table present: digits 0-9
4, bit 1 | 1 bit       | direct jump table present: upper case A-Z
4, bit 2 | 1 bit       | direct jump table present: lower case a-z
5..      | 10x 2 bytes | optional: direct jump table 0-9.
..       | 26x 2 bytes | optional: direct jump table A-Z.
..       | 26x 2 bytes | optional: direct jump table a-z.
..       | 2 bytes     | jump offset to sequentially accessed codepoints. Only present if at least 1 DJT is present.

It is possible for the Bounding Box to be smaller than the font's theoretical bounding box, and its top-left coordinate doesn't necessarily have to be (0,0). This can occur for example when a subset of the codepoints from the font are selected to be included in the generated output.

Based on the Bounding Box, your rendering routine might decide to exclude (some of) the unused pixels when rendering, to save some space on the screen. This is not done by bdf2array yet: for flexibility, the choice is up to you. Actually, it wouldn't save any memory either, because this encoding only stores data for the smallest-fitting bounding box _per glyph_.

The Direct Jump Tables (DJT) contain 2 byte (little endian) offset values. These are measured from the first byte of the font data (ie the first codepoint will have offset 0). Non-existent codepoints are listed in the DJT with offset 0xFFFF.

Codepoints for which a DJT is present should be listed first in the codepoint data section, followed by the remaining codepoints. A jump offset to those remaining codepoints is provided (if at least one DJT is present). This allows for faster sequential scanning through the remaining codepoints (and a faster 'not found' conclusion when no glyph data is present for the requested codepoint).

### Per-codepoint data

Byte nr      | Size   | Data
-------------|--------|---------------------------------------------------------------------------------------------------------
0            | 1 byte | codepoint
1            | 1 byte | jump offset to next codepoint (calculated from byte 0)
2, bits 1..0 | 2 bits | number of row bytes (the possible values 1..4 are encoded as 0..3)
2, bits 6..2 | 5 bits | glyph y offset (from the top edge of the font's bounding box, cannot be negative)
3            | 1 byte | glyph x offset (from the left edge of the font's bounding box, cannot be negative)
4            | 1 byte | horizontal glyph pitch (=advance) (cannot be smaller than x offset + the number of pixel columns in the glyph data)
5..n         | var    | pixel data. The number of columns can be determined from the number of row bytes and the jump offset

To indicate the end of the data, the last codepoint is followed by a codepoint entry which has a zero jump offset field.
