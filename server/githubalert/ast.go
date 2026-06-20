// Package githubalert implements GitHub Alert Style rendering for goldmark.
// It parses blockquote lines prefixed with "> [!TYPE]" and renders them
// as styled alert callouts with SVG icons and title text.
package githubalert

import "github.com/yuin/goldmark/ast"

var AlertKind = ast.NewNodeKind("GitHubAlert")

type Alert struct {
	ast.BaseBlock
}

func NewAlert() *Alert {
	return &Alert{}
}

func (n *Alert) Kind() ast.NodeKind {
	return AlertKind
}

func (n *Alert) Dump(source []byte, level int) {
	kind := ""
	if v, ok := n.AttributeString("kind"); ok {
		kind = v.(string)
	}
	ast.DumpHelper(n, source, level, map[string]string{"kind": kind}, nil)
}

var AlertTitleKind = ast.NewNodeKind("GitHubAlertTitle")

type AlertTitle struct {
	ast.BaseBlock
}

func NewAlertTitle() *AlertTitle {
	return &AlertTitle{}
}

func (n *AlertTitle) Kind() ast.NodeKind {
	return AlertTitleKind
}

func (n *AlertTitle) Dump(source []byte, level int) {
	kind := ""
	if v, ok := n.AttributeString("kind"); ok {
		kind = v.(string)
	}
	ast.DumpHelper(n, source, level, map[string]string{"kind": kind}, nil)
}
