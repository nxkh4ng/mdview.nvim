package main

import (
	"bytes"
	"strconv"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type LineInjector struct{}

func (l *LineInjector) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if n.Type() == ast.TypeInline {
			return ast.WalkContinue, nil
		}

		if n.Kind() == ast.KindList {
			return ast.WalkContinue, nil
		}

		start := n.Pos()
		if start < 0 {
			if lines := n.Lines(); lines.Len() > 0 {
				start = lines.At(0).Start
			} else {
				return ast.WalkContinue, nil
			}
		}

		line := bytes.Count(source[:start], []byte("\n")) + 1

		n.SetAttribute([]byte("data-source-line"), []byte(strconv.Itoa(line)))
		return ast.WalkContinue, nil
	})
}
