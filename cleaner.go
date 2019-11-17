package goofx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/golang/glog"
)

// Cleaner cleans the given data to return valid XML.
type Cleaner interface {
	// Init initializes cleaner with the given data
	Init([]byte) error
	// Cleanup processes an initialized cleaner and returns cleaned data.
	CleanupXML() (*bytes.Buffer, error)
}

type cleaner struct {
	decoder     *xml.Decoder
	tagStack    []*xml.StartElement // A stack to keep parsed tags.
	lastTagIdx  int                 // Index for the last tag on the stack.
	lastData    string              // Holds the last parsed char data.
	lastElement *xml.StartElement   // Last parsed element start tag.
	cleanXML    bytes.Buffer        // Buffer to hold cleaned XML.
}

// NewCleaner returns an instance of cleaner.
func NewCleaner() Cleaner {
	return &cleaner{
		tagStack:   make([]*xml.StartElement, 1000),
		lastTagIdx: -1,
		lastData:   "",
	}
}

// Init initializes this cleaner with the given data.
func (c *cleaner) Init(data []byte) error {
	// Detect the start of XML like data.
	xmlIndex := bytes.Index(data, []byte("<OFX>"))
	if xmlIndex == -1 {
		return fmt.Errorf("error - invalid file, OFX tag not found")
	}

	// Start a xml decoder on the context of source data that is XML like.
	reader := bytes.NewReader(data[xmlIndex:])
	c.decoder = xml.NewDecoder(reader)

	return nil
}

// printTagStack returns the tag stack for debugging.
func (c *cleaner) printTagStack() []string {
	stack := make([]string, 0)
	for i := 0; c.lastTagIdx+1 > i; i++ {
		stack = append(stack, c.tagStack[i].Name.Local)
	}
	return stack
}

func (c *cleaner) processStartElement(t xml.StartElement) error {
	glog.V(3).Infof("case start element %s", t.Name.Local)
	// If last data exists, it takes highest precedence. This is a start tag and last data
	// exists implies that the previous end tag is missing.
	if c.lastData != "" {
		glog.V(3).Infof("StartTag: previous tag needs to be closed: %s %v", c.lastData, c.lastElement)
		// If last data exists but no last element, the current tag being a start element
		// implies the data is missing both start and end tags.
		if c.lastElement == nil {
			return fmt.Errorf("error: charData(%s) missing start and end tags", c.lastData)
		}
		writeElement(c.lastElement, c.lastData, &c.cleanXML)
		c.lastData = ""
		c.lastElement = nil
	}
	// If this tag is an aggregate, flush it and push it on the stack for dequeue later.
	// If this tag is an element, update lastElement as it can't have nested tags.
	if IsAggregate(t.Name.Local) {
		glog.V(3).Infof("StartTag: %s is aggregate, pushing to stack", t.Name.Local)
		c.lastTagIdx++
		c.tagStack[c.lastTagIdx] = &t
		writeStartTag(&t, &c.cleanXML)
	} else {
		glog.V(3).Infof("StartTag: %s is NOT aggregate, updating lastElement", t.Name.Local)
		c.lastElement = &t
	}
	glog.V(3).Infof("Stack: %#v", c.printTagStack())

	return nil
}

func (c *cleaner) processEndElement(t xml.EndElement) error {
	glog.V(3).Infof("case end element %s", t.Name.Local)
	isAggregate := IsAggregate(t.Name.Local)
	// If last data exists, it takes highest precedence. This is an end tag and last data
	// exists implies this must be the corresponding end tag if this is an element.
	// If this is an aggregate, the previous element end tag is missing.
	if c.lastData != "" {
		glog.V(3).Infof("EndTag: previous tag needs to be closed: %s %v", c.lastData, c.lastElement)
		// If this is an element tag but not the same as lastElement, that is an error.
		if c.lastElement != nil && t.Name != c.lastElement.Name && !isAggregate {
			// There is a last element as well this is a data (non aggregate) element.
			// We can not determine which of the two is missing a closing tag.
			return fmt.Errorf("error: charData(%s) has ambigious closing tags", c.lastData)
		}
		// If this is an aggregate tag and lastElement isn't set, that is an error.
		if c.lastElement == nil && isAggregate {
			return fmt.Errorf("error: charData(%s) missing start and end tags", c.lastData)
		}
		if c.lastElement != nil {
			// Implies this tag is aggregate or same as lastElement.
			writeElement(c.lastElement, c.lastData, &c.cleanXML)
		} else {
			// Implies this tag is not aggregate.
			writeElementFromName(t.Name, c.lastData, &c.cleanXML)
		}
		c.lastData = ""
		c.lastElement = nil
	}

	if isAggregate {
		glog.V(3).Infof("EndTag: %s is aggregate, popping from stack", t.Name.Local)
		glog.V(3).Infof("Stack: %#v", c.printTagStack())
		// Close every open tag till the current closing tag is matched.
		for c.lastTagIdx > -1 {
			lastTag := c.tagStack[c.lastTagIdx].Name
			writeEndTag(lastTag, &c.cleanXML)
			c.lastTagIdx--
			if lastTag.Local == t.Name.Local {
				break
			}
		}
		glog.V(3).Infof("Stack: %#v", c.printTagStack())
	}

	return nil
}

// CleanupXML returns cleaned XML from the given data.
func (c *cleaner) CleanupXML() (*bytes.Buffer, error) {
	// Read parsed XML tokens from the XML decoder into token and re-assemble them into another
	// buffer, while adding any missing starting or closing tags and trimming spaces/newlines.
	for {
		token, err := c.decoder.RawToken()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch t := token.(type) {
		case xml.CharData:
			c.lastData = EscapeString(strings.TrimSpace(string([]byte(t))))
			glog.V(3).Infof("case chardata (%s) %#v", c.lastData, t)
		case xml.StartElement:
			if err := c.processStartElement(t); err != nil {
				return nil, err
			}
		case xml.EndElement:
			if err := c.processEndElement(t); err != nil {
				return nil, err
			}
		}
	}
	return &c.cleanXML, nil
}
