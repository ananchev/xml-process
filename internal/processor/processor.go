package processor

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
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

func tranform() {

	// Stand-in for other io.Readers like a file
	xmlFile, err := os.Open("input.xml")
	if err != nil {
		fmt.Println("Error opening XML file:", err)
		return
	}
	defer xmlFile.Close()

	var buf bytes.Buffer

	decoder := xml.NewDecoder(xmlFile)
	decoder.CharsetReader = identReader
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "    ")

	// encoder := xml.NewEncoder(&buf)

	for {
		// Read tokens from the XML document in a stream.
		t, err := decoder.Token()

		// If we are at the end of the file, we are done
		if err == io.EOF {
			// log.Println("The end")
			break
		} else if err != nil {
			log.Fatalf("Error decoding token: %s", err)
		} else if t == nil {
			break
		}

		// Here, we inspect the token
		switch se := t.(type) {

		// We have the start of an element.
		// However, we have the complete token in t
		case xml.StartElement:
			switch se.Name.Local {

			// Found an item, so we process it
			case "Part":
				var p Part
				// We decode the documents elements into our data model...
				if err = decoder.DecodeElement(&p, &se); err != nil {
					log.Fatalf("Error decoding item: %s", err)
				}

				// // And use it for whatever we want to
				// log.Printf("Document Name: '%s' with Id: %s", d.DocumentName, d.DocumentlD)

				for i, document := range p.Documents {
					r, dl := document.hasMoreThanOneDataset()
					if r {

						p.Documents = slices.Delete(p.Documents, i, i+1)

						for docLink, docURL := range dl {
							//log.Println(element)

							e1 := Document{
								DocumentlD:    document.DocumentlD,
								DocumentRev:   document.DocumentRev,
								DocumentName:  document.DocumentName,
								DocumentLinks: []string{docLink},
								DOC_URL_TMP:   []string{docURL},
								DOC_REL_TMP:   document.DOC_REL_TMP,
							}
							p.Documents = append(p.Documents, e1)
						}
					}
				}
				if err = encoder.EncodeElement(p, se); err != nil {
					log.Fatal(err)
				}
				continue
			}
		}
		if err := encoder.EncodeToken(xml.CopyToken(t)); err != nil {
			log.Fatal(err)
		}

	}

	// must call flush, otherwise some elements will be missing
	if err := encoder.Flush(); err != nil {
		log.Fatal(err)
	}

	//fmt.Println(buf.String())

	f, err := os.Create("output.xml")
	if err != nil {
		fmt.Println("Error creating XML file:", err)
		return
	}

	w := bufio.NewWriter(f)
	n4, err := w.WriteString(buf.String())
	if err != nil {
		fmt.Println("Error writing XML file:", err)
		return
	}
	log.Printf("wrote %d bytes\n", n4)

}

func (d Document) hasMoreThanOneDataset() (res bool, datasets map[string]string) {

	if len(d.DocumentLinks) <= 1 {
		return false, nil
	}

	ret := make(map[string]string)
	for i := range d.DocumentLinks {
		ret[d.DocumentLinks[i]] = d.DOC_URL_TMP[i]
	}

	return true, ret
}

func identReader(encoding string, input io.Reader) (io.Reader, error) {
	return input, nil
}
