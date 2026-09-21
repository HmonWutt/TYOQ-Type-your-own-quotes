package epub

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"io"
	"net/url"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func ExtractText(file string) (string, error) {
	r, err := zip.OpenReader(file)
	if err != nil {
		return "", err
	}
	defer r.Close()
	return extract(&r.Reader)
}

func ExtractTextReader(r *zip.Reader) (string, error) {
	return extract(r)
}

func extract(r *zip.Reader) (string, error) {
	var pages []string
	opf, err := opfPath(r)
	if err != nil {
		return "", err
	}
	files, err := spineFiles(r, opf)
	if err != nil {
		return "", err
	}

	for _, name := range files {
		f, err := r.Open(name)
		if err != nil {
			continue // missing chapter: degrade gracefully
		}
		if t := htmlText(f); t != "" {
			pages = append(pages, t)
		}
		f.Close()
	}
	// return pages, nil // allow user to select chapters to type
	return strings.Join(pages, "\n\n"), nil
}

// opfPath locates the OPF manifest via META-INF/container.xml
func opfPath(zr *zip.Reader) (string, error) {
	f, err := zr.Open("META-INF/container.xml")
	if err != nil {
		return "", err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				return "", errors.New("epub: no rootfile in container.xml")
			}
			return "", err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "rootfile" {
			continue
		}
		for _, a := range se.Attr {
			if a.Name.Local == "full-path" {
				return a.Value, nil
			}
		}
	}
}

// spineFiles returns the content documents listed in the OPF spine,
// in reading order. The manifest maps item ids to file paths; the
// spine orders those ids.
func spineFiles(zr *zip.Reader, opfPath string) ([]string, error) {
	f, err := zr.Open(opfPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	manifest := map[string]string{}
	var order []string
	inManifest := false
	inSpine := false
	base := path.Dir(opfPath)

	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "manifest":
				inManifest = true
			case "spine":
				inSpine = true
			case "item":
				if !inManifest {
					continue
				}
				var id, href string
				for _, a := range t.Attr {
					switch a.Name.Local {
					case "id":
						id = a.Value
					case "href":
						href = a.Value
					}
				}
				// hrefs are relative to the OPF and may be URL-escaped
				if unescaped, err := url.PathUnescape(href); err == nil {
					href = unescaped
				}
				manifest[id] = path.Join(base, href)
			case "itemref":
				if !inSpine {
					continue
				}
				for _, a := range t.Attr {
					if a.Name.Local == "idref" {
						if name, ok := manifest[a.Value]; ok {
							order = append(order, name)
						}
					}
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "manifest":
				inManifest = false
			case "spine":
				inSpine = false
			}
		}
	}
	return order, nil
}

const blockTags = "p, h1, h2, h3, h4, h5, h6, li, blockquote, pre, div, td, dd, dt, figcaption, section, article"

// Only leaf block elements are collected, so text nested under wrapper divs is not duplicated.
func htmlText(r io.Reader) string {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return ""
	}
	// script and style subtrees are removed before collection.
	doc.Find("script, style").Remove()

	var lines []string
	doc.Find(blockTags).FilterFunction(func(_ int, s *goquery.Selection) bool {
		return s.Find(blockTags).Length() == 0
	}).Each(func(_ int, s *goquery.Selection) {
		if t := strings.Join(strings.Fields(s.Text()), " "); t != "" {
			lines = append(lines, t)
		}
	})
	return strings.Join(lines, "\n")
}
