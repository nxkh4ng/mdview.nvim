package main

import (
	"bytes"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

func chromaCSS() string {
	var buf bytes.Buffer

	buf.WriteString("@media (prefers-color-scheme: light) {\n")
	_ = html.New(html.WithClasses(true)).WriteCSS(&buf, styles.Get("github"))
	buf.WriteString("}\n")

	buf.WriteString("@media (prefers-color-scheme: dark) {\n")
	_ = html.New(html.WithClasses(true)).WriteCSS(&buf, styles.Get("xcode-dark"))
	buf.WriteString("}\n")

	return buf.String()
}
