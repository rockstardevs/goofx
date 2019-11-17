package goofx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/golang/glog"
)

// Cleaner cleans the given data to return valid XML.
type Cleaner interface {
	CleanupXML(data []byte) (*bytes.Buffer, error)
}

type cleaner struct{}

var cleanerSingleton *cleaner
var initCleaner sync.Once

// GetCleaner returns the singleton instance of cleaner.
func GetCleaner() Cleaner {
	initCleaner.Do(func() {
		cleanerSingleton = &cleaner{}
	})
	return cleanerSingleton
}

// CleanupXML returns cleaned XML from the given data.
func (c cleaner) CleanupXML(data []byte) (*bytes.Buffer, error) {
	var (
		xmlIndex    int                               // Index for start of XML like data.
		tagStack    = make([]*xml.StartElement, 1000) // A stack to keep parsed tags.
		lastTagIdx  = -1                              // Index for the last tag on the stack.
		lastData    string                            // Holds the last parsed char data.
		lastElement *xml.StartElement                 // Last parsed element start tag.
		cleanXML    bytes.Buffer                      // Buffer to hold cleaned XML.

	)
	// Detect the start of XML like data.
	if xmlIndex = bytes.Index(data, []byte("<OFX>")); xmlIndex == -1 {
		return nil, fmt.Errorf("error - invalid file, OFX tag not found")
	}

	// Start a xml decoder on the context of source data that is XML like.
	reader := bytes.NewReader(data[xmlIndex:])
	decoder := xml.NewDecoder(reader)

	printTagStack := func() []string {
		stack := make([]string, 0)
		for c := 0; lastTagIdx+1 > c; c++ {
			stack = append(stack, tagStack[c].Name.Local)
		}
		return stack
	}

	// Read parsed XML tokens from the XML decoder into token and re-assemble them into another
	// buffer, while adding any missing starting or closing tags and trimming spaces/newlines.
	for {
		token, err := decoder.RawToken()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch t := token.(type) {
		case xml.CharData:
			lastData = escapeString(strings.TrimSpace(string([]byte(t))))
			glog.V(3).Infof("case chardata (%s) %#v", lastData, t)
		case xml.StartElement:
			glog.V(3).Infof("case start element %s", t.Name.Local)
			// If last data exists, it takes highest precedence. This is a start tag and last data
			// exists implies that the previous end tag is missing.
			if lastData != "" {
				glog.V(3).Infof("StartTag: previous tag needs to be closed: %s %v", lastData, lastElement)
				// If last data exists but no last element, the current tag being a start element
				// implies the data is missing both start and end tags.
				if lastElement == nil {
					return nil, fmt.Errorf("error: charData(%s) missing start and end tags", lastData)
				}
				writeElement(lastElement, lastData, &cleanXML)
				lastData = ""
				lastElement = nil
			}
			// If this tag is an aggregate, flush it and push it on the stack for dequeue later.
			// If this tag is an element, update lastElement as it can't have nested tags.
			if IsAggregate(t.Name.Local) {
				glog.V(3).Infof("StartTag: %s is aggregate, pushing to stack", t.Name.Local)
				lastTagIdx++
				tagStack[lastTagIdx] = &t
				writeStartTag(&t, &cleanXML)
			} else {
				glog.V(3).Infof("StartTag: %s is NOT aggregate, updating lastElement", t.Name.Local)
				lastElement = &t
			}
			glog.V(3).Infof("Stack: %#v", printTagStack())
		case xml.EndElement:
			glog.V(3).Infof("case end element %s", t.Name.Local)
			isAggregate := IsAggregate(t.Name.Local)
			// If last data exists, it takes highest precedence. This is an end tag and last data
			// exists implies this must be the corresponding end tag if this is an element.
			// If this is an aggregate, the previous element end tag is missing.
			if lastData != "" {
				glog.V(3).Infof("EndTag: previous tag needs to be closed: %s %v", lastData, lastElement)
				// If this is an element tag but not the same as lastElement, that is an error.
				if lastElement != nil && t.Name != lastElement.Name && !isAggregate {
					// There is a last element as well this is a data (non aggregate) element.
					// We can not determine which of the two is missing a closing tag.
					return nil, fmt.Errorf("error: charData(%s) has ambigious closing tags", lastData)
				}
				// If this is an aggregate tag and lastElement isn't set, that is an error.
				if lastElement == nil && isAggregate {
					return nil, fmt.Errorf("error: charData(%s) missing start and end tags", lastData)
				}
				if lastElement != nil {
					// Implies this tag is aggregate or same as lastElement.
					writeElement(lastElement, lastData, &cleanXML)
				} else {
					// Implies this tag is not aggregate.
					writeElementFromName(t.Name, lastData, &cleanXML)
				}
				lastData = ""
				lastElement = nil
			}

			if isAggregate {
				glog.V(3).Infof("EndTag: %s is aggregate, popping from stack", t.Name.Local)
				glog.V(3).Infof("Stack: %#v", printTagStack())
				// Close every open tag till the current closing tag is matched.
				for lastTagIdx > -1 {
					lastTag := tagStack[lastTagIdx].Name
					writeEndTag(lastTag, &cleanXML)
					lastTagIdx--
					if lastTag.Local == t.Name.Local {
						break
					}
				}
				glog.V(3).Infof("Stack: %#v", printTagStack())
			}
		}
	}
	return &cleanXML, nil
}
