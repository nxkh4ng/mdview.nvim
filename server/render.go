package main

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	"mdview.nvim/githubalert"
	"mdview.nvim/metablock"
)

type noPreWrapper struct{}

type result struct {
	html []byte
	err  error
}

var (
	baseMd = sync.OnceValue(func() goldmark.Markdown {
		return newGoldmark("")
	})
	mdCache      sync.Map
	renderMu     sync.Mutex
	renderCancel context.CancelFunc
)

func newGoldmark(baseDir string) goldmark.Markdown {
	baseDir = strings.ReplaceAll(baseDir, "\\", "/")
	transformers := []util.PrioritizedValue{}

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
			emoji.Emoji,

			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
					chromahtml.WithPreWrapper(noPreWrapper{}),
				),
				highlighting.WithWrapperRenderer(codeBlockWrapper),
			),
			&githubalert.Extension{},
			&metablock.Extension{},
		),

		goldmark.WithParserOptions(
			parser.WithASTTransformers(transformers...),
		),

		goldmark.WithRendererOptions(
			renderer.WithNodeRenderers(
				util.Prioritized(NewLineNumberRenderer(), 0),
			),
		),
	)
}

func markdownToHTML(source []byte, baseDir string) ([]byte, error) {
	var buf bytes.Buffer

	md := getMarkdown(baseDir)

	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader, parser.WithContext(parser.NewContext()))

	stampLineNumbers(doc, source)

	err := md.Renderer().Render(&buf, source, doc)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func markdownToHTMLWithTimeout(source []byte, baseDir string, timeout time.Duration) ([]byte, error) {
	renderMu.Lock()
	if renderCancel != nil {
		renderCancel()
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	renderCancel = cancel
	renderMu.Unlock()

	ch := make(chan result, 1)
	go func() {
		html, err := markdownToHTML(source, baseDir)
		select {
		case ch <- result{html, err}:
		case <-ctx.Done():
		}
	}()

	select {
	case r := <-ch:
		return r.html, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// stampLineNumbers sets data-source-line on block nodes that are
// NOT handled by LineNumberRenderer at render time.
func stampLineNumbers(doc ast.Node, source []byte) {
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		// Skip inlines, document root, and list containers
		if n.Type() == ast.TypeInline {
			return ast.WalkContinue, nil
		}
		if n.Kind() == ast.KindDocument || n.Kind() == ast.KindList {
			return ast.WalkContinue, nil
		}

		// Skip nodes that LineNumberRenderer handles (they get it at render time)
		switch n.Kind() {
		case ast.KindParagraph, ast.KindHeading, ast.KindBlockquote,
			ast.KindListItem, ast.KindThematicBreak:
			return ast.WalkContinue, nil
		}

		line := lineOfNode(source, n)
		if line > 0 {
			n.SetAttribute([]byte("data-source-line"), []byte(strconv.Itoa(line)))
		}
		return ast.WalkContinue, nil
	})
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
