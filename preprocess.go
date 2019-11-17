package goofx

import "regexp"

// preprocessOFXData applies one-off transforms to fix bad data.
// This should not be required as the library matures.
func preprocessOFXData(content []byte) []byte {
	from := regexp.MustCompile(`(</CURDEF>\s+)(<BANKID>)`)
	content = from.ReplaceAll(content, []byte("$1<BANKACCTFROM>$2"))
	return content
}
