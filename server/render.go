package main

import (
	"bytes"

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
		highlighting.NewHighlighting(
			highlighting.WithStyle("monokai"),
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
