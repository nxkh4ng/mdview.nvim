package githubalert

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var alertRegex = regexp.MustCompile(`^\[!([a-zA-Z]+)\]\s*$`)

var validKinds = map[string]struct{}{
	"note":      {},
	"tip":       {},
	"important": {},
	"warning":   {},
	"caution":   {},
}

type alertParser struct{}

func newAlertParser() parser.BlockParser {
	return &alertParser{}
}

func (p *alertParser) Trigger() []byte {
	return []byte{'>'}
}

func (p *alertParser) process(reader text.Reader) (bool, int) {
	line, _ := reader.PeekLine()
	indent, pos := util.IndentWidth(line, reader.LineOffset())
	if indent > 3 || pos >= len(line) || line[pos] != '>' {
		return false, 0
	}

	advanceBy := 1
	if pos+advanceBy >= len(line) || line[pos+advanceBy] == '\n' {
		return true, advanceBy
	}
	if line[pos+advanceBy] == ' ' || line[pos+advanceBy] == '\t' {
		advanceBy++
	}
	if line[pos+advanceBy-1] == '\t' {
		reader.SetPadding(2)
	}
	return true, advanceBy
}

func (p *alertParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	ok, advanceBy := p.process(reader)
	if !ok {
		return nil, parser.NoChildren
	}

	line, _ := reader.PeekLine()
	if len(line) <= advanceBy {
		return nil, parser.NoChildren
	}

	content := line[advanceBy:]
	match := alertRegex.FindSubmatch(content)
	if match == nil {
		return nil, parser.NoChildren
	}

	kind := strings.ToLower(string(match[1]))
	if _, ok := validKinds[kind]; !ok {
		return nil, parser.NoChildren
	}

	alert := NewAlert()
	alert.SetAttributeString("kind", kind)

	if idx := bytes.IndexByte(line, ']'); idx >= 0 {
		reader.Advance(idx)
	}

	return alert, parser.HasChildren
}

func (p *alertParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	ok, advanceBy := p.process(reader)
	if !ok {
		return parser.Close
	}

	reader.Advance(advanceBy)
	return parser.Continue | parser.HasChildren
}

func (p *alertParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (p *alertParser) CanInterruptParagraph() bool {
	return true
}

func (p *alertParser) CanAcceptIndentedLine() bool {
	return false
}

type alertTitleParser struct{}

func newAlertTitleParser() parser.BlockParser {
	return &alertTitleParser{}
}

func (p *alertTitleParser) Trigger() []byte {
	return []byte{']'}
}

func (p *alertTitleParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	if parent.ChildCount() != 0 || parent.Kind() != AlertKind {
		return nil, parser.NoChildren
	}

	reader.Advance(1)

	kind := ""
	if v, ok := parent.AttributeString("kind"); ok {
		kind = v.(string)
	}

	title := NewAlertTitle()
	title.SetAttributeString("kind", kind)

	return title, parser.NoChildren
}

func (p *alertTitleParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

func (p *alertTitleParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (p *alertTitleParser) CanInterruptParagraph() bool {
	return false
}

func (p *alertTitleParser) CanAcceptIndentedLine() bool {
	return true
}
