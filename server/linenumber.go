package main

import (
	"bytes"
	"strconv"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type LineNumberRenderer struct{}

func NewLineNumberRenderer() *LineNumberRenderer {
	return &LineNumberRenderer{}
}

func (r *LineNumberRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindParagraph, r.renderParagraph)
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindListItem, r.renderListItem)
	reg.Register(ast.KindThematicBreak, r.renderThematicBreak)
}

// lineOfNode computes the 1-based source line number for an AST node.
// This function is a pure computation: source + node.Pos/node.Lines → int.
// It can be reused by any extension (GitHub Alert, Mermaid, …).
func lineOfNode(source []byte, n ast.Node) int {
	start := n.Pos()
	if start < 0 {
		if lines := n.Lines(); lines.Len() > 0 {
			start = lines.At(0).Start
		} else {
			return 0
		}
	}
	return bytes.Count(source[:start], []byte("\n")) + 1
}

func (r *LineNumberRenderer) renderParagraph(
	w util.BufWriter, source []byte, n ast.Node, entering bool,
) (ast.WalkStatus, error) {
	if entering {
		line := lineOfNode(source, n)
		w.WriteString(`<p data-source-line="` + strconv.Itoa(line) + `"`)
		if n.Attributes() != nil {
			html.RenderAttributes(w, n, html.GlobalAttributeFilter)
		}
		w.WriteByte('>')
	} else {
		w.WriteString("</p>\n")
	}
	return ast.WalkContinue, nil
}

func (r *LineNumberRenderer) renderHeading(
	w util.BufWriter, source []byte, n ast.Node, entering bool,
) (ast.WalkStatus, error) {
	h := n.(*ast.Heading)
	if entering {
		line := lineOfNode(source, n)
		w.WriteString(`<h` + strconv.Itoa(h.Level) + ` data-source-line="` + strconv.Itoa(line) + `"`)
		if n.Attributes() != nil {
			html.RenderAttributes(w, n, html.HeadingAttributeFilter)
		}
		w.WriteByte('>')
	} else {
		w.WriteString(`</h` + strconv.Itoa(h.Level) + ">\n")
	}
	return ast.WalkContinue, nil
}

func (r *LineNumberRenderer) renderBlockquote(
	w util.BufWriter, source []byte, n ast.Node, entering bool,
) (ast.WalkStatus, error) {
	if entering {
		line := lineOfNode(source, n)
		w.WriteString(`<blockquote data-source-line="` + strconv.Itoa(line) + `"`)
		if n.Attributes() != nil {
			html.RenderAttributes(w, n, html.GlobalAttributeFilter)
		}
		w.WriteString(">\n")
	} else {
		w.WriteString("</blockquote>\n")
	}
	return ast.WalkContinue, nil
}

func (r *LineNumberRenderer) renderListItem(
	w util.BufWriter, source []byte, n ast.Node, entering bool,
) (ast.WalkStatus, error) {
	if entering {
		line := lineOfNode(source, n)
		w.WriteString(`<li data-source-line="` + strconv.Itoa(line) + `"`)
		if n.Attributes() != nil {
			html.RenderAttributes(w, n, html.GlobalAttributeFilter)
		}
		w.WriteByte('>')
	} else {
		w.WriteString("</li>\n")
	}
	return ast.WalkContinue, nil
}

func (r *LineNumberRenderer) renderThematicBreak(
	w util.BufWriter, source []byte, n ast.Node, entering bool,
) (ast.WalkStatus, error) {
	if entering {
		line := lineOfNode(source, n)
		w.WriteString(`<hr data-source-line="` + strconv.Itoa(line) + `"`)
		if n.Attributes() != nil {
			html.RenderAttributes(w, n, html.GlobalAttributeFilter)
		}
		w.WriteString(">\n")
	}
	return ast.WalkSkipChildren, nil
}
