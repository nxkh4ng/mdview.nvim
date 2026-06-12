package main

import (
	"bytes"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
)

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
		extension.TaskList,
		extension.DefinitionList,
		extension.Footnote,
		highlighting.NewHighlighting(
			// highlighting.WithStyle("github"),
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(true),
			),
		),
	),
)

func markdownToHTML(source []byte) ([]byte, error) {
	var buf bytes.Buffer
	err := md.Convert(source, &buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
