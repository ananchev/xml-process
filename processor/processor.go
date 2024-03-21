package processor

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/ananchev/processxml/logger"
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

func TransformXML(input string) {

	// Stand-in for other io.Readers like a file
	xmlFile, err := os.Open(input)

	logger.LogInfo("Running transformation for file '{f}'", "f", input)
	if err != nil {
		logger.LogError("Error opening XML file: {e}", "e", err)
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
			logger.LogInfo("Reached the end the file")
			break
		} else if err != nil {
			logger.LogError("Error decoding token: {e}", "e", err)
			break
		} else if t == nil {
			break
		}

		// Here, we inspect the token
		switch se := t.(type) {

		// We have the start of an element. However, we have the complete token in t
		case xml.StartElement:
			switch se.Name.Local {

			// Found an item, so we process it
			case "Part":
				var p Part
				// We decode the documents elements into our data model...
				if err = decoder.DecodeElement(&p, &se); err != nil {
					logger.LogError("Error decoding item: {e}", "e", err)
					return
				}

				// // And use it for whatever we want to
				// processxml.logger.LogInfo.Printf("Document Name: '%s' with Id: %s", d.DocumentName, d.DocumentlD)
				logger.LogInfo("Processing part '{p}/{r}'", "p", p.Id, "r", p.Revision)
				for i, document := range p.Documents {
					r, dl := document.hasMoreThanOneDataset()
					if r {
						logger.LogInfo("Extracted all document link elements")

						logger.LogInfo("Deleting the Document element....")
						p.Documents = slices.Delete(p.Documents, i, i+1)

						for docLink, docURL := range dl {
							//processxml.logger.LogInfo.Println(element)
							e1 := Document{
								DocumentlD:    document.DocumentlD,
								DocumentRev:   document.DocumentRev,
								DocumentName:  document.DocumentName,
								DocumentLinks: []string{docLink},
								DOC_URL_TMP:   []string{docURL},
								DOC_REL_TMP:   document.DOC_REL_TMP,
							}
							p.Documents = append(p.Documents, e1)
							logger.LogInfo("Appended new document element for document link {l}", "l", docLink)
						}
					}
				}
				if err = encoder.EncodeElement(p, se); err != nil {
					logger.LogInfo("Error encoding the modified part element for '{p}/{r}': {e}", "p", p.Id, "r", p.Revision, "e", err)
					return
				}
				continue
			}
		}
		if err := encoder.EncodeToken(xml.CopyToken(t)); err != nil {
			logger.LogInfo("Error encoding the complete XML token: {e}", "e", err)
			return
		}

	}

	// must call flush, otherwise some elements will be missing
	if err := encoder.Flush(); err != nil {
		logger.LogInfo("Error while flushing the buffered XML to the underlying writer: {e}", "e", err)
		return
	}

	// create a temporary XML file
	file_no_ext := strings.TrimSuffix(input, filepath.Ext(input))

	tmp_file := file_no_ext + "_tmp.xml"
	f, err := os.Create(tmp_file)
	if err != nil {
		logger.LogInfo("Error creating temporary XML file '{t}': {e}", "t", tmp_file, "e", err)
		return
	}

	// write the new XML
	w := bufio.NewWriter(f)
	n4, err := w.WriteString(buf.String())
	if err != nil {
		logger.LogInfo("Error writing to temporary XML file '{t}': {e}", "t", tmp_file, "e", err)
		return
	}
	logger.LogInfo("Wrote {d} bytes into temporary XML file '{t}'", "d", strconv.Itoa(n4), "t", tmp_file)

	// remove the original file
	r := os.Remove(input)
	if r != nil {
		logger.LogInfo("Error removing the orginal XML file '{i}': {e}", "i", input, "e", err)
		return
	}
	logger.LogInfo("Removed the orginal XML file '{i}'.", "i", input)

	// rename the temp file to the original
	n := os.Rename(tmp_file, input)
	if n != nil {
		logger.LogInfo("Error renaming the temporary XML file '{i}': {e}", "i", tmp_file, "e", err)
		return
	}
	logger.LogInfo("Renamed the temporary XML file '{i}' to '{i1}'.", "i", tmp_file, "i1", input)
	logger.LogInfo("Finished processing '{i1}'.", "i1", input)

}

func (d Document) hasMoreThanOneDataset() (res bool, datasets map[string]string) {

	if len(d.DocumentLinks) <= 1 {
		logger.LogInfo("Document '{d}/{r}' does not have multiple DocumentLink elements", "d", d.DocumentlD, "r", d.DocumentRev)
		return false, nil
	}

	ret := make(map[string]string)
	logger.LogInfo("Found '{n}' documentLink elements for '{d}/{r}'.", "n", strconv.Itoa(len(d.DocumentLinks)), "d", d.DocumentlD, "r", d.DocumentRev)
	for i := range d.DocumentLinks {
		logger.LogInfo("...storing document link '{l1}' with doc_url_tmp '{l2}'", "l1", d.DocumentLinks[i], "l2", d.DOC_URL_TMP[i])
		ret[d.DocumentLinks[i]] = d.DOC_URL_TMP[i]
	}
	return true, ret
}

func identReader(encoding string, input io.Reader) (io.Reader, error) {
	return input, nil
}
