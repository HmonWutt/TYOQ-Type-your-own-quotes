package epub

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

// buildEPUB packs entries into an in-memory EPUB. The mimetype entry
// is written first and stored uncompressed, as the EPUB spec requires.
func buildEPUB(t *testing.T, entries map[string]string) *zip.Reader {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	mt, err := w.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mt.Write([]byte("application/epub+zip")); err != nil {
		t.Fatal(err)
	}

	for name, body := range entries {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// fixture returns a small but awkward EPUB: namespaced container
// tags, a non-lexical spine order, script/style noise, and a page
// with an image only.
func fixture(t *testing.T) *zip.Reader {
	return buildEPUB(t, map[string]string{
		"META-INF/container.xml": `<?xml version="1.0"?>
<ocf:container xmlns:ocf="urn:oasis:names:tc:opendocument:xmlns:container">
  <ocf:rootfiles>
    <ocf:rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </ocf:rootfiles>
</ocf:container>`,

		"OEBPS/content.opf": `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <manifest>
    <item id="cover" href="cover.xhtml" media-type="application/xhtml+xml"/>
    <item id="ch2" href="text/ch2.xhtml" media-type="application/xhtml+xml"/>
    <item id="ch1" href="text/ch1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="ch1"/>
    <itemref idref="ch2"/>
    <itemref idref="cover"/>
  </spine>
</package>`,

		"OEBPS/text/ch1.xhtml": `<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml"><head><title>One</title>
<style>p { color: red; }</style>
<script>var ignoreMe = "noise";</script></head>
<body>
<div class="wrapper"><div class="inner">
<p>First   paragraph with   extra spaces.</p>
<p>Text with <b>bold</b> and <i>italics</i>.</p>
</div></div>
</body></html>`,

		"OEBPS/text/ch2.xhtml": `<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml"><body>
<h1>Chapter Two</h1>
<p>Second chapter &amp; ampersand decoded.</p>
</body></html>`,

		"OEBPS/cover.xhtml": `<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml"><body>
<div><img src="cover.png" alt="cover"/></div>
</body></html>`,
	})
}

func TestExtractText(t *testing.T) {
	text, err := ExtractTextReader(fixture(t))
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"First paragraph with extra spaces.", // whitespace collapsed
		"Text with bold and italics.",        // inline tags flattened
		"Chapter Two",
		"Second chapter & ampersand decoded.", // entity decoded
	}
	for _, w := range want {
		if !strings.Contains(text, w) {
			t.Errorf("text missing %q\ngot:\n%s", w, text)
		}
	}
	if !strings.Contains(text, "&amp;") && strings.Contains(text, "&ampamp;") {
		t.Error("entity not decoded correctly")
	}

	// reading order: chapter one content before chapter two
	if i, j := strings.Index(text, "First"), strings.Index(text, "Second chapter"); i > j {
		t.Error("chapters not in spine order")
	}

	// noise must be stripped
	for _, noise := range []string{"ignoreMe", "color: red"} {
		if strings.Contains(text, noise) {
			t.Errorf("text contains %q", noise)
		}
	}

	// empty pages (image-only cover) must not add blank paragraphs
	if strings.Contains(text, "\n\n\n") {
		t.Error("unexpected run of blank lines between chapters")
	}
}

func TestOPFPath(t *testing.T) {
	r := fixture(t)
	p, err := opfPath(r)
	if err != nil {
		t.Fatal(err)
	}
	if p != "OEBPS/content.opf" {
		t.Errorf("got %q, want OEBPS/content.opf", p)
	}
}

func TestSpineFiles(t *testing.T) {
	r := fixture(t)
	files, err := spineFiles(r, "OEBPS/content.opf")
	if err != nil {
		t.Fatal(err)
	}
	want := "OEBPS/text/ch1.xhtml, OEBPS/text/ch2.xhtml, OEBPS/cover.xhtml"
	if got := strings.Join(files, ", "); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMissingContainer(t *testing.T) {
	r := buildEPUB(t, map[string]string{"OEBPS/content.opf": "<package/>"})
	if _, err := ExtractTextReader(r); err == nil {
		t.Error("expected error for missing container.xml")
	}
}

func TestLeafBlockNoDuplicates(t *testing.T) {
	r := buildEPUB(t, map[string]string{
		"META-INF/container.xml": `<container><rootfile full-path="a.opf"/></container>`,
		"a.opf": `<package><manifest>
      <item id="c" href="c.xhtml"/>
    </manifest><spine><itemref idref="c"/></spine></package>`,
		"c.xhtml": `<html><body><blockquote><p>quoted</p></blockquote></body></html>`,
	})
	text, err := ExtractTextReader(r)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(text, "quoted"); got != 1 {
		t.Errorf("quoted text appeared %d times, want 1\ngot:\n%s", got, text)
	}
}

func TestURLEscapedHref(t *testing.T) {
	r := buildEPUB(t, map[string]string{
		"META-INF/container.xml": `<container><rootfile full-path="a.opf"/></container>`,
		"a.opf": `<package><manifest>
      <item id="c" href="my%20chapter.xhtml"/>
    </manifest><spine><itemref idref="c"/></spine></package>`,
		"my chapter.xhtml": `<html><body><p>escaped href</p></body></html>`,
	})
	text, err := ExtractTextReader(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "escaped href") {
		t.Errorf("URL-escaped href not resolved; got:\n%s", text)
	}
}
