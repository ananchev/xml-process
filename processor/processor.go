package processor

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"slices"
	"strconv"
)

type Parts struct {
	Parts []Part `xml:"Part"`
}

type Part struct {
	Level                             int        `xml:"Level"`
	Sequence                          string     `xml:"Sequence"`
	Id                                string     `xml:"ID"`
	Revision                          int        `xml:"Revision"`
	Name                              string     `xml:"Name"`
	Quantity                          string     `xml:"Quantity"`
	Unit                              string     `xml:"Unit"`
	Type                              string     `xml:"Type"`
	ReleaseStatus                     string     `xml:"ReleaseStatus"`
	SBRL                              string     `xml:"SBRL"`
	TCL                               string     `xml:"TCL"`
	PCN                               string     `xml:"PCN"`
	Manufacturer                      string     `xml:"Manufacturer"`
	MPN                               string     `xml:"MPN"`
	SAPBE01ProcurementType2           string     `xml:"SAPBE01ProcurementType2"`
	SAPBE01SpecialProcurement2        string     `xml:"SAPBE01SpecialProcurement2"`
	SAPBE01MaterialProvisionIndicator string     `xml:"SAPBE01MaterialProvisionIndicator"`
	Documents                         []Document `xml:"Document"`
}

type Document struct {
	DocumentlD    string   `xml:"DocumentlD"`
	DocumentRev   string   `xml:"DocumentRev"`
	DocumentName  string   `xml:"DocumentName"`
	DocumentLinks []string `xml:"DocumentLink"`
	DOC_URL_TMP   []string `xml:"DOC_URL_TMP"`
	DOC_REL_TMP   string   `xml:"DOC_REL_TMP"`
}

var XMLRewrite bool

func TransformXML(logfile string, input string) {

	InitLogger(logfile)

	XMLRewrite = false

	xmlFile, err := os.Open(input)

	LogInfo("Running transformation for file '{f}'", "f", input)
	if err != nil {
		LogError("Error opening XML file: {e}", "e", err)
		return
	}
	defer xmlFile.Close()

	var buf bytes.Buffer

	utf_reader := NewValidUTF8Reader(xmlFile)

	decoder := xml.NewDecoder(utf_reader)
	decoder.CharsetReader = identReader
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "    ")

	for {
		// Read tokens from the XML document in a stream.
		t, err := decoder.Token()

		// If we are at the end of the file, we are done
		if err == io.EOF {
			LogInfo("Reached the end the file")
			break
		} else if err != nil {
			LogError("Error decoding token: {e}", "e", err)
			break
		} else if t == nil {
			break
		}

		// inspect the token
		switch se := t.(type) {

		// start of an element and the complete token in t
		case xml.StartElement:
			switch se.Name.Local {

			// Found an item, so we process it
			case "Part":
				var p Part
				// We decode the documents elements into our data model...
				if err = decoder.DecodeElement(&p, &se); err != nil {
					LogError("Error decoding item: {e}", "e", err)
					return
				}

				// ...and use it for whatever we want to
				LogInfo("Processing part '{p}/{r}'", "p", p.Id, "r", p.Revision)
				for i, document := range p.Documents {
					r, dl := document.hasMoreThanOneDataset()
					if r {
						LogInfo("Extracted all document link elements")

						LogInfo("Deleting the Document element....")
						p.Documents = slices.Delete(p.Documents, i, i+1)

						for docLink, docURL := range dl {
							e1 := Document{
								DocumentlD:    document.DocumentlD,
								DocumentRev:   document.DocumentRev,
								DocumentName:  document.DocumentName,
								DocumentLinks: []string{docLink},
								DOC_URL_TMP:   []string{docURL},
								DOC_REL_TMP:   document.DOC_REL_TMP,
							}
							p.Documents = append(p.Documents, e1)
							LogInfo("Appended new document element for document link {l}", "l", docLink)
						}
					}
				}
				if err = encoder.EncodeElement(p, se); err != nil {
					LogInfo("Error encoding the modified part element for '{p}/{r}': {e}", "p", p.Id, "r", p.Revision, "e", err)
					return
				}
				continue
			}
		}
		if err := encoder.EncodeToken(xml.CopyToken(t)); err != nil {
			LogInfo("Error encoding the complete XML token: {e}", "e", err)
			return
		}

	}

	// must call flush, otherwise some elements will be missing
	if err := encoder.Flush(); err != nil {
		LogInfo("Error while flushing the buffered XML to the underlying writer: {e}", "e", err)
		return
	}

	if XMLRewrite {
		LogInfo("XML file '{t}' will be rewritten to reflect the transformed document link elements", "t", input)
	} else {
		LogInfo("No modifications are required for XML file '{t}'.", "t", input)
		return
	}

	// with os.Create if the file already exists, it is truncated
	f, err := os.Create(input)
	if err != nil {
		LogInfo("Error trunkating file '{t}': {e}", "t", input, "e", err)
		return
	}

	// write the new XML
	w := bufio.NewWriter(f)
	n4, err := w.WriteString(buf.String())
	if err != nil {
		LogInfo("Error writing to '{t}': {e}", "t", input, "e", err)
		return
	}
	LogInfo("Wrote {d} bytes into '{t}'", "d", strconv.Itoa(n4), "t", input)

}

func (d Document) hasMoreThanOneDataset() (res bool, datasets map[string]string) {

	if len(d.DocumentLinks) <= 1 {
		LogInfo("Document '{d}/{r}' does not have multiple DocumentLink elements", "d", d.DocumentlD, "r", d.DocumentRev)
		return false, nil
	}

	ret := make(map[string]string)
	LogInfo("Found '{n}' documentLink elements for '{d}/{r}'.", "n", strconv.Itoa(len(d.DocumentLinks)), "d", d.DocumentlD, "r", d.DocumentRev)
	XMLRewrite = true
	for i := range d.DocumentLinks {
		LogInfo("...storing document link '{l1}' with doc_url_tmp '{l2}'", "l1", d.DocumentLinks[i], "l2", d.DOC_URL_TMP[i])
		ret[d.DocumentLinks[i]] = d.DOC_URL_TMP[i]
	}
	return true, ret
}

func identReader(encoding string, input io.Reader) (io.Reader, error) {
	return input, nil
}
