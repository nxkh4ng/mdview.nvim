package metablock

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type Extension struct{}

func (e *Extension) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(newMetaParser(), 0),
		),
	)
	md.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewMetaBlockRenderer(), 0),
		),
	)
}
