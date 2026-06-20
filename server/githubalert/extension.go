package githubalert

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type Extension struct{}

func (e *Extension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(newAlertParser(), 799),
			util.Prioritized(newAlertTitleParser(), 0),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewAlertRenderer(), 0),
			util.Prioritized(NewAlertTitleRenderer(), 0),
		),
	)
}
