// Package metablock implements a goldmark extension for parsing and rendering
// YAML front matter as a vertical metadata table.
package metablock

import (
	"github.com/goccy/go-yaml"
	"github.com/yuin/goldmark/ast"
)

var MetaBlockKind = ast.NewNodeKind("MetaBlock")

type MetaBlock struct {
	ast.BaseBlock
	Meta    yaml.MapSlice
	KeyLine map[string]int
}

func NewMetaBlock() *MetaBlock {
	return &MetaBlock{}
}

func (m *MetaBlock) Kind() ast.NodeKind {
	return MetaBlockKind
}

func (m *MetaBlock) Dump(source []byte, level int) {
	ast.DumpHelper(m, source, level, nil, nil)
}
