package chapters

import (
	"bytes"
	"encoding/binary"
	"errors"
	"unicode/utf16"
)

// Chapter is one podcast chapter.
type Chapter struct {
	Index   int    `json:"index"`
	Title   string `json:"title"`
	StartMs uint32 `json:"startMs"`
	EndMs   uint32 `json:"endMs"`
}

// synchsafe decodes a 4-byte synchsafe integer (7 bits per byte).
func synchsafe(b []byte) uint32 {
	return uint32(b[0])<<21 | uint32(b[1])<<14 | uint32(b[2])<<7 | uint32(b[3])
}

// ParseID3Chapters extracts CHAP frames from the leading bytes of an MP3.
// Supports ID3v2.3 (plain frame sizes) and v2.4 (synchsafe frame sizes).
func ParseID3Chapters(data []byte) ([]Chapter, error) {
	if len(data) < 10 || !bytes.Equal(data[0:3], []byte("ID3")) {
		return nil, errors.New("no ID3v2 tag")
	}
	major := data[3] // version: 3 or 4
	tagSize := synchsafe(data[6:10])
	end := 10 + int(tagSize)
	if end > len(data) || end < 10 {
		end = len(data) // tag may extend past our fetched window; parse what we have
	}
	body := data[10:end]

	var chapters []Chapter
	i := 0
	for i+10 <= len(body) {
		id := string(body[i : i+4])
		if id == "\x00\x00\x00\x00" {
			break // padding
		}
		var size int
		if major >= 4 {
			size = int(synchsafe(body[i+4 : i+8]))
		} else {
			size = int(binary.BigEndian.Uint32(body[i+4 : i+8]))
		}
		i += 10 // skip frame header
		if size <= 0 || i+size > len(body) {
			break
		}
		if id == "CHAP" {
			if ch, ok := parseCHAP(body[i:i+size], major); ok {
				chapters = append(chapters, ch)
			}
		}
		i += size
	}
	for idx := range chapters {
		chapters[idx].Index = idx
	}
	return chapters, nil
}

// parseCHAP parses a CHAP frame body:
//
//	element-id (null-terminated) | startMs(4) | endMs(4) | startOffset(4) | endOffset(4) | sub-frames...
func parseCHAP(b []byte, major byte) (Chapter, bool) {
	nul := bytes.IndexByte(b, 0)
	if nul < 0 || nul+1+16 > len(b) {
		return Chapter{}, false
	}
	p := nul + 1
	ch := Chapter{
		StartMs: binary.BigEndian.Uint32(b[p : p+4]),
		EndMs:   binary.BigEndian.Uint32(b[p+4 : p+8]),
	}
	p += 16 // skip start/end ms + start/end byte offsets
	// Remaining bytes are embedded sub-frames; find TIT2 for the title.
	sub := b[p:]
	j := 0
	for j+10 <= len(sub) {
		sid := string(sub[j : j+4])
		var ssize int
		if major >= 4 {
			ssize = int(synchsafe(sub[j+4 : j+8]))
		} else {
			ssize = int(binary.BigEndian.Uint32(sub[j+4 : j+8]))
		}
		j += 10
		if ssize <= 0 || j+ssize > len(sub) {
			break
		}
		if sid == "TIT2" {
			ch.Title = decodeText(sub[j : j+ssize])
		}
		j += ssize
	}
	return ch, true
}

// decodeText decodes an ID3 text frame body (first byte is the encoding).
func decodeText(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	enc := b[0]
	payload := b[1:]
	switch enc {
	case 1: // UTF-16 with BOM
		return decodeUTF16(payload, binary.LittleEndian)
	case 2: // UTF-16BE without BOM
		return decodeUTF16(payload, binary.BigEndian)
	default: // 0 = ISO-8859-1, 3 = UTF-8
		return string(bytes.Trim(payload, "\x00"))
	}
}

// decodeUTF16 decodes a UTF-16 byte payload, honoring a leading BOM if present.
func decodeUTF16(payload []byte, order binary.ByteOrder) string {
	if len(payload) >= 2 {
		switch {
		case payload[0] == 0xFF && payload[1] == 0xFE:
			order, payload = binary.LittleEndian, payload[2:]
		case payload[0] == 0xFE && payload[1] == 0xFF:
			order, payload = binary.BigEndian, payload[2:]
		}
	}
	u16 := make([]uint16, 0, len(payload)/2)
	for k := 0; k+1 < len(payload); k += 2 {
		u16 = append(u16, order.Uint16(payload[k:k+2]))
	}
	return string(bytes.Trim([]byte(string(utf16.Decode(u16))), "\x00"))
}
