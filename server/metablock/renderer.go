package metablock

import (
	"fmt"
	"html"
	"strconv"

	"github.com/goccy/go-yaml"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	goldmark_html "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type metaBlockRenderer struct{}

func NewMetaBlockRenderer() renderer.NodeRenderer {
	return &metaBlockRenderer{}
}

func (r *metaBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(MetaBlockKind, r.render)
}

func (r *metaBlockRenderer) render(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	metaBlock := n.(*MetaBlock)
	if v, founded := n.AttributeString("yaml-error"); founded {
		if msg, ok := v.(string); ok {
			w.WriteString(`<pre class="yaml-error" data-source-line="1">`)
			w.WriteString(html.EscapeString(msg))
			w.WriteString("</pre>\n")
			return ast.WalkSkipChildren, nil
		}
	}
	if metaBlock.Meta == nil {
		return ast.WalkSkipChildren, nil
	}

	w.WriteString("<table ")
	if n.Attributes() != nil {
		goldmark_html.RenderAttributes(w, n, goldmark_html.GlobalAttributeFilter)
	}
	w.WriteString("><tbody>")

	for _, item := range metaBlock.Meta {
		key := fmt.Sprintf("%v", item.Key)
		line := metaBlock.KeyLine[key]

		if line > 0 {
			w.WriteString(`<tr data-source-line="` + strconv.Itoa(line) + `"><th>`)
		} else {
			w.WriteString("<tr><th>")
		}

		w.WriteString(html.EscapeString(key))
		w.WriteString("</th><td>")
		renderValue(w, item.Value)
		w.WriteString("</td></tr>")
	}

	w.WriteString("</tbody></table>\n")
	return ast.WalkSkipChildren, nil
}

func renderValue(w util.BufWriter, v any) {
	switch val := v.(type) {
	case nil:
	case string:
		w.WriteString(html.EscapeString(val))
	case int:
		w.WriteString(strconv.Itoa(val))
	case float64:
		w.WriteString(strconv.FormatFloat(val, 'f', -1, 64))
	case bool:
		w.WriteString(strconv.FormatBool(val))
	case []any:
		renderArray(w, val)
	case yaml.MapSlice:
		renderMapHorizontalSlice(w, val)
	default:
		w.WriteString(html.EscapeString(fmt.Sprintf("%v", val)))
	}
}

func renderArray(w util.BufWriter, arr []any) {
	for i, item := range arr {
		if i > 0 {
			w.WriteString(", ")
		}
		renderValue(w, item)
	}
}

func renderMapHorizontalSlice(w util.BufWriter, m yaml.MapSlice) {
	w.WriteString("<table><thead><tr>")
	for _, item := range m {
		w.WriteString("<th>")
		w.WriteString(html.EscapeString(fmt.Sprintf("%v", item.Key)))
		w.WriteString("</th>")
	}
	w.WriteString("</tr></thead><tbody><tr>")
	for _, item := range m {
		w.WriteString("<td>")
		renderValue(w, item.Value)
		w.WriteString("</td>")
	}
	w.WriteString("</tr></tbody></table>")
}
