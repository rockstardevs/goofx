package goofx

import (
	"bytes"
	"encoding/xml"
	"unicode/utf8"

	"github.com/golang/glog"
)

//revive:disable:exported

var (
	// XML Escape sequences.
	// from https://golang.org/src/encoding/xml/xml.go:1840
	escQuot = []byte("&#34;") // shorter than "&quot;"
	escApos = []byte("&#39;") // shorter than "&apos;"
	escAmp  = []byte("&amp;")
	escLt   = []byte("&lt;")
	escGt   = []byte("&gt;")
	escTab  = []byte("&#x9;")
	escNl   = []byte("&#xA;")
	escCr   = []byte("&#xD;")
	escFffd = []byte("\uFFFD") // Unicode replacement character
)

// Decide whether the given rune is in the XML Character Range, per
// the Char production of http://www.xml.com/axml/testaxml.htm,
// Section 2.2 Characters.
// Lifted from https://golang.org/src/encoding/xml/xml.go:1102
func isInCharacterRange(r rune) (inrange bool) {
	return r == 0x09 ||
		r == 0x0A ||
		r == 0x0D ||
		r >= 0x20 && r <= 0xDF77 ||
		r >= 0xE000 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0x10FFFF
}

// escapeString returns properly escaped XML equivalent of the plain text data s.
// based on https://golang.org/src/encoding/xml/xml.go:1907
func escapeString(s string) string {
	var (
		result bytes.Buffer
		esc    []byte
		last   = 0
	)
	for i := 0; i < len(s); {
		r, width := utf8.DecodeRuneInString(s[i:])
		i += width
		switch r {
		case '"':
			esc = escQuot
		case '\'':
			esc = escApos
		case '&':
			esc = escAmp
		case '<':
			esc = escLt
		case '>':
			esc = escGt
		case '\t':
			esc = escTab
		case '\n':
			esc = escNl
		case '\r':
			esc = escCr
		default:
			if !isInCharacterRange(r) || (r == 0xFFFD && width == 1) {
				esc = escFffd
				break
			}
			continue
		}
		result.WriteString(s[last : i-width])
		result.Write(esc)
		last = i
	}
	result.WriteString(s[last:])
	return result.String()
}

// writeStartTag writes the given start element to the given buffer.
// based on https://golang.org/src/encoding/xml/marshal.go:678
func writeStartTag(e *xml.StartElement, buff *bytes.Buffer) {
	glog.V(3).Infof("pushed: %s", e.Name.Local)
	buff.WriteByte('<')
	buff.WriteString(e.Name.Local)
	// Namespace
	if e.Name.Space != "" {
		buff.WriteString(` xmlns="`)
		buff.WriteString(escapeString(e.Name.Space))
		buff.WriteByte('"')
	}
	// Attributes
	for _, attr := range e.Attr {
		name := attr.Name
		if name.Local == "" {
			continue
		}
		buff.WriteByte(' ')
		buff.WriteString(name.Local)
		buff.WriteString(`="`)
		buff.WriteString(escapeString(attr.Value))
		buff.WriteByte('"')
	}
	buff.WriteByte('>')
}

// writeEndTag writes the closing tag for the given end element to the given buffer.
// based on https://golang.org/src/encoding/xml/marshal.go:717
func writeEndTag(name xml.Name, buff *bytes.Buffer) {
	glog.V(3).Infof("popped: %s", name.Local)
	buff.Write([]byte("</"))
	buff.WriteString(name.Local)
	buff.WriteByte('>')
}

// writeElement writes the starting and closing tags and data for the given elements
// to the given buffer.
func writeElement(startTag *xml.StartElement, data string, buff *bytes.Buffer) {
	writeStartTag(startTag, buff)
	buff.WriteString(data)
	writeEndTag(startTag.Name, buff)
}

// writeElementFromName writes the starting and closing tags and data for the given elements
// to the given buffer.
func writeElementFromName(name xml.Name, data string, buff *bytes.Buffer) {
	startTag := xml.StartElement{Name: name}
	writeStartTag(&startTag, buff)
	buff.WriteString(data)
	writeEndTag(name, buff)
}
