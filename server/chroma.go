package main

import (
	"bytes"
	"net/http"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

var chromaCSS = sync.OnceValue(func() string {
	var buf bytes.Buffer

	buf.WriteString("@media (prefers-color-scheme: light) {\n")
	_ = html.New(html.WithClasses(true)).WriteCSS(&buf, styles.Get("github"))
	buf.WriteString("}\n")

	buf.WriteString("@media (prefers-color-scheme: dark) {\n")
	_ = html.New(html.WithClasses(true)).WriteCSS(&buf, styles.Get("github-dark"))
	buf.WriteString("}\n")

	s := buf.String()
	s = strings.ReplaceAll(s, ".chroma.light", ".chroma")
	s = strings.ReplaceAll(s, ".chroma.dark", ".chroma")

	return s
})

func handleChromaCSS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Write([]byte(chromaCSS()))
	}
}
