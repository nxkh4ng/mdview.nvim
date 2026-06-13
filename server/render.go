package main

import (
	"bytes"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
)

type noPreWrapper struct{}

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
		extension.TaskList,
		extension.DefinitionList,

		// Syntax highlighting
		highlighting.NewHighlighting(
			highlighting.WithStyle("github"),
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(true),
				chromahtml.WithPreWrapper(noPreWrapper{}),
			),
			highlighting.WithWrapperRenderer(codeBlockWrapper),
		),
	),

	// Inject `data-source-line` into DOM
	goldmark.WithParserOptions(
		parser.WithASTTransformers(
			util.Prioritized(&LineInjector{}, 100),
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

func (noPreWrapper) Start(code bool, styleAttr string) string {
	return ""
}

func (noPreWrapper) End(code bool) string {
	return ""
}

func codeBlockWrapper(w util.BufWriter, c highlighting.CodeBlockContext, entering bool) {
	if entering {
		line := "0"
		if attrs := c.Attributes(); attrs != nil {
			if val, ok := attrs.GetString("data-source-line"); ok {
				if b, ok := val.([]byte); ok {
					line = string(b)
				}
			}
		}

		w.WriteString(`<pre class="chroma" data-source-line="` + line + `">`)
		w.WriteString(`<code`)
		if !c.Highlighted() {
			if lang, ok := c.Language(); ok {
				w.WriteString(` class="language-`)
				w.Write(lang)
				w.WriteString(`"`)
			}
		}
		w.WriteString(`>`)
	} else {
		w.WriteString(`</code></pre>`)
	}
}
