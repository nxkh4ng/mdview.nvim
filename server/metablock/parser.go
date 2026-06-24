package metablock

import (
	"bytes"

	"github.com/goccy/go-yaml"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type metaParser struct{}

func newMetaParser() parser.BlockParser {
	return &metaParser{}
}

func isSeparator(line []byte) bool {
	trimmed := util.TrimRightSpace(util.TrimLeftSpace(line))
	for i := range trimmed {
		if trimmed[i] != '-' {
			return false
		}
	}
	return true
}

func (p *metaParser) Trigger() []byte {
	return []byte{'-'}
}

func (p *metaParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	linenum, _ := reader.Position()
	if linenum != 0 {
		return nil, parser.NoChildren
	}

	line, _ := reader.PeekLine()
	if isSeparator(line) {
		return NewMetaBlock(), parser.NoChildren
	}

	return nil, parser.NoChildren
}

func (p *metaParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	if isSeparator(line) && !util.IsBlank(line) {
		reader.Advance(segment.Len())
		return parser.Close
	}
	node.Lines().Append(segment)
	return parser.Continue | parser.NoChildren
}

func (p *metaParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	metaBlock := node.(*MetaBlock)

	lines := node.Lines()
	var buf bytes.Buffer
	for i := 0; i < lines.Len(); i++ {
		segment := lines.At(i)
		buf.Write(segment.Value(reader.Source()))
	}

	meta := yaml.MapSlice{}
	if err := yaml.UnmarshalWithOptions(buf.Bytes(), &meta, yaml.UseOrderedMap()); err != nil {
		metaBlock.SetAttributeString("yaml-error", err.Error())
		return
	}
	metaBlock.Meta = meta

	source := reader.Source()
	linesOf := map[string]int{}
	for i := 0; i < lines.Len(); i++ {
		segment := lines.At(i)
		lineBytes := segment.Value(source)

		lineNum := bytes.Count(source[:segment.Start], []byte("\n")) + 1
		trimLine := bytes.TrimSpace(lineBytes)
		if idx := bytes.IndexByte(trimLine, ':'); idx > 0 {
			indent := len(lineBytes) - len(bytes.TrimLeft(lineBytes, " \t"))
			if indent == 0 {
				key := string(bytes.TrimSpace(trimLine[:idx]))
				linesOf[key] = lineNum
			}
		}
	}
	metaBlock.KeyLine = linesOf
}

func (p *metaParser) CanInterruptParagraph() bool {
	return false
}

func (p *metaParser) CanAcceptIndentedLine() bool {
	return false
}
