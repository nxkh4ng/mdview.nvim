package main

import (
	"bytes"
	"strings"
	"sync"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
)

type noPreWrapper struct{}

var baseMd = sync.OnceValue(func() goldmark.Markdown {
	return newGoldmark("")
})
var mdCache sync.Map

func newGoldmark(baseDir string) goldmark.Markdown {
	baseDir = strings.ReplaceAll(baseDir, "\\", "/")
	transformers := []util.PrioritizedValue{
		util.Prioritized(&LineInjector{}, 100),
	}

	if baseDir != "" {
		transformers = append(
			transformers,
			util.Prioritized(&AssetRewriter{BaseDir: baseDir}, 50),
		)
	}

	return goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
			extension.DefinitionList,

			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
					chromahtml.WithPreWrapper(noPreWrapper{}),
				),
				highlighting.WithWrapperRenderer(codeBlockWrapper),
			),
		),

		goldmark.WithParserOptions(
			parser.WithASTTransformers(transformers...),
		),
	)
}

func markdownToHTML(source []byte, baseDir string) ([]byte, error) {
	var buf bytes.Buffer

	md := getMarkdown(baseDir)

	err := md.Convert(source, &buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func getMarkdown(baseDir string) goldmark.Markdown {
	if baseDir == "" {
		return baseMd()
	}

	if cached, ok := mdCache.Load(baseDir); ok {
		return cached.(goldmark.Markdown)
	}

	md := newGoldmark(baseDir)
	mdCache.Store(baseDir, md)
	return md
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
