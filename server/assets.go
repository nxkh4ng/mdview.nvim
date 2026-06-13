package main

import (
	"path"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type AssetRewriter struct {
	BaseDir string
}

func (r *AssetRewriter) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	if r.BaseDir == "" {
		return
	}

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n.Kind() {
		case ast.KindImage:
			img := n.(*ast.Image)
			dest := string(img.Destination)

			if needsRewrite(dest) {
				img.Destination = []byte("/local" + path.Join(r.BaseDir, dest))
			}
		case ast.KindLink:
			link := n.(*ast.Link)
			dest := string(link.Destination)

			if needsRewrite(dest) {
				link.Destination = []byte("/local" + path.Join(r.BaseDir, dest))
			}
		}

		return ast.WalkContinue, nil
	})
}

func needsRewrite(dest string) bool {
	return !strings.HasPrefix(dest, "http://") &&
		!strings.HasPrefix(dest, "https://") &&
		!strings.HasPrefix(dest, "data:") &&
		!strings.HasPrefix(dest, "#") &&
		!strings.HasPrefix(dest, "/")
}
