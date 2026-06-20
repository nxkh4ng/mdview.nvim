package main

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func TestMarkdownToHTML_SimpleParagraph(t *testing.T) {
	html, err := markdownToHTML([]byte("hello"), "")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(html, []byte("<p")) {
		t.Errorf("expected <p> tag, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte("hello")) {
		t.Errorf("expected 'hello' in output, got:\n%s", html)
	}
}

func TestMarkdownToHTML_Heading(t *testing.T) {
	html, err := markdownToHTML([]byte("# Title"), "")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(html, []byte("<h1")) {
		t.Errorf("expected <h1> tag, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte("Title")) {
		t.Errorf("expected 'Title' in output, got:\n%s", html)
	}
}

func TestMarkdownToHTML_Table(t *testing.T) {
	input := []byte("| A | B |\n|---|---|\n| 1 | 2 |")
	html, err := markdownToHTML(input, "")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(html, []byte("<table")) {
		t.Errorf("expected <table> tag, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte("<th")) {
		t.Errorf("expected <th> tag, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte("<td")) {
		t.Errorf("expected <td> tag, got:\n%s", html)
	}
}

func TestMarkdownToHTML_List(t *testing.T) {
	input := []byte("- item 1\n- item 2")
	html, err := markdownToHTML(input, "")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(html, []byte("<ul")) {
		t.Errorf("expected <ul> tag, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte("<li")) {
		t.Errorf("expected <li> tag, got:\n%s", html)
	}
}

func TestMarkdownToHTML_CodeBlock(t *testing.T) {
	input := []byte("```go\nfmt.Println(\"hello\")\n```")
	html, err := markdownToHTML(input, "")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(html, []byte("<pre")) {
		t.Errorf("expected <pre> tag, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte("chroma")) {
		t.Errorf("expected chroma class, got:\n%s", html)
	}
}

func TestDataLineInjection_Simple(t *testing.T) {
	input := []byte("hello\n\nworld")
	html, err := markdownToHTML(input, "")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(html, []byte(`data-source-line="1"`)) {
		t.Errorf("expected data-source-line=1, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte(`data-source-line="3"`)) {
		t.Errorf("expected data-source-line=3 (empty line=2), got:\n%s", html)
	}
}

func TestDataLineInjection_Table(t *testing.T) {
	input := []byte("| A | B |\n|---|---|\n| 1 | 2 |")
	html, err := markdownToHTML(input, "")
	if err != nil {
		t.Fatal(err)
	}

	// Table and cells should have data-source-line
	if !bytes.Contains(html, []byte(`data-source-line="1"`)) {
		t.Errorf("expected data-source-line=1 on table/header, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte(`data-source-line="3"`)) {
		t.Errorf("expected data-source-line=3 on data row, got:\n%s", html)
	}
}

func TestDataLineInjection_DefinitionList(t *testing.T) {
	input := []byte("term\n: definition")
	html, err := markdownToHTML(input, "")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(html, []byte(`<dt`)) {
		t.Errorf("expected <dt>, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte(`data-source-line="1"`)) {
		t.Errorf("expected data-source-line=1 on dt, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte(`data-source-line="2"`)) {
		t.Errorf("expected data-source-line=2 on dd, got:\n%s", html)
	}
}

func TestDataLineInjection_HTMLBlock(t *testing.T) {
	// Raw HTML block: kbd should NOT have data-source-line
	// (it's raw HTML passthrough, not a goldmark AST node)
	input := []byte("before\n\n<kbd>Ctrl</kbd>\n\nafter")
	html, err := markdownToHTML(input, "")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(html, []byte(`data-source-line="1"`)) {
		t.Errorf("expected data-source-line=1 on first paragraph, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte(`data-source-line="5"`)) {
		t.Errorf("expected data-source-line=5 on last paragraph, got:\n%s", html)
	}
}

func TestDataLineInjection_EmptyInput(t *testing.T) {
	html, err := markdownToHTML([]byte(""), "")
	if err != nil {
		t.Fatal(err)
	}

	if len(html) != 0 {
		t.Errorf("expected empty output for empty input, got:\n%s", html)
	}
}

func TestAssetRewriter_LocalImage(t *testing.T) {
	html, err := markdownToHTML([]byte("![](image.png)"), "/home/user/docs")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(html), "/local/home/user/docs/image.png") {
		t.Errorf("expected rewritten image path, got:\n%s", html)
	}
}

func TestAssetRewriter_LocalLink(t *testing.T) {
	html, err := markdownToHTML([]byte("[doc](doc.pdf)"), "/home/user/docs")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(html), "/local/home/user/docs/doc.pdf") {
		t.Errorf("expected rewritten link path, got:\n%s", html)
	}
}

func TestAssetRewriter_AbsolutePathSkipped(t *testing.T) {
	html, err := markdownToHTML([]byte("![](/absolute/image.png)"), "/home/user/docs")
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(html), "/local") {
		t.Errorf("absolute path should NOT be rewritten, got:\n%s", html)
	}
}

func TestAssetRewriter_ExternalURLSkipped(t *testing.T) {
	html, err := markdownToHTML([]byte("![img](https://example.com/foo.png)"), "/home/user/docs")
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(html), "/local") {
		t.Errorf("external URL should NOT be rewritten, got:\n%s", html)
	}
}

func TestAssetRewriter_EmptyBaseDir(t *testing.T) {
	html, err := markdownToHTML([]byte("![](image.png)"), "")
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(html), "/local") {
		t.Errorf("empty baseDir should not rewrite, got:\n%s", html)
	}
}

func TestMarkdownToHTML_Strikethrough(t *testing.T) {
	html, err := markdownToHTML([]byte("~~striked~~"), "")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(html, []byte("<del")) {
		t.Errorf("expected <del> tag for strikethrough, got:\n%s", html)
	}
}

func TestMarkdownToHTML_TaskList(t *testing.T) {
	input := []byte("- [ ] todo\n- [x] done")
	html, err := markdownToHTML(input, "")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(html, []byte(`type="checkbox"`)) {
		t.Errorf("expected checkbox input, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte(`checked`)) {
		t.Errorf("expected checked attribute, got:\n%s", html)
	}
}

func TestMarkdownToHTML_DefinitionList(t *testing.T) {
	input := []byte("term\n: definition")
	html, err := markdownToHTML(input, "")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(html, []byte("<dl")) {
		t.Errorf("expected <dl> tag, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte("<dt")) {
		t.Errorf("expected <dt> tag, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte("<dd")) {
		t.Errorf("expected <dd> tag, got:\n%s", html)
	}
}

func TestMarkdownToHTML_Linkify(t *testing.T) {
	html, err := markdownToHTML([]byte("visit https://example.com"), "")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(html, []byte(`<a href="https://example.com"`)) {
		t.Errorf("expected auto-linked URL, got:\n%s", html)
	}
}

func TestMarkdownToHTML_WindowsBackslash(t *testing.T) {
	html, err := markdownToHTML([]byte("![img](image.png)"), "C:\\Users\\docs")
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(html), `\\`) {
		t.Errorf("backslash should be normalized, got:\n%s", html)
	}
	if !strings.Contains(string(html), "/local/C:/Users/docs/image.png") {
		t.Errorf("expected normalized Windows path, got:\n%s", html)
	}
}

// --- Benchmark helpers ---

// benchSource generates markdown with n lines of content.
func benchSource(n int) []byte {
	var buf bytes.Buffer
	buf.WriteString("# Title\n\n")
	for i := range n / 3 {
		// Mix of paragraphs, lists, and code blocks
		buf.WriteString("Paragraph ")
		buf.WriteString(strconv.Itoa(i))
		buf.WriteString(" with **bold** and *italic*.\n\n")

		buf.WriteString("- list item ")
		buf.WriteString(strconv.Itoa(i))
		buf.WriteString("\n\n")

		buf.WriteString("```go\n")
		buf.WriteString("func foo() {\n")
		buf.WriteString("    fmt.Println(\"hello\")\n")
		buf.WriteString("}\n")
		buf.WriteString("```\n\n")
	}
	return buf.Bytes()
}

var benchTestMD = func() []byte {
	b, _ := os.ReadFile("../test.md")
	if b == nil {
		return benchSource(597)
	}
	return b
}()

// --- Benchmark: full pipeline ---

func BenchmarkMarkdownToHTML_Short(b *testing.B) {
	src := benchSource(50)

	for b.Loop() {
		markdownToHTML(src, "")
	}
}

func BenchmarkMarkdownToHTML_Medium(b *testing.B) {
	src := benchSource(500)

	for b.Loop() {
		markdownToHTML(src, "")
	}
}

func BenchmarkMarkdownToHTML_Large(b *testing.B) {
	src := benchSource(5000)

	for b.Loop() {
		markdownToHTML(src, "")
	}
}

func BenchmarkMarkdownToHTML_TestFile(b *testing.B) {
	src := benchTestMD

	for b.Loop() {
		markdownToHTML(src, "")
	}
}

// --- Benchmark: micro-benchmarks ---

func BenchmarkLineOfNode(b *testing.B) {
	src := []byte("# Title\n\nParagraph with **bold**.\n\n- item 1\n- item 2\n")
	md := getMarkdown("")
	reader := text.NewReader(src)
	doc := md.Parser().Parse(reader, parser.WithContext(parser.NewContext()))

	var nodes []ast.Node
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Type() == ast.TypeBlock {
			nodes = append(nodes, n)
		}
		return ast.WalkContinue, nil
	})

	for b.Loop() {
		for _, n := range nodes {
			lineOfNode(src, n)
		}
	}
}

func BenchmarkStampLineNumbers(b *testing.B) {
	src := benchTestMD
	md := getMarkdown("")
	reader := text.NewReader(src)
	doc := md.Parser().Parse(reader, parser.WithContext(parser.NewContext()))

	for b.Loop() {
		stampLineNumbers(doc, src)
	}
}

func TestAssetRewriter_MailtoLink(t *testing.T) {
	html, err := markdownToHTML([]byte("[email](mailto:test@test.com)"), "/home/user/docs")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(html), "/local") {
		t.Errorf("mailto: link should NOT be rewritten, got:\n%s", html)
	}
}

func TestAlertNoteHTML(t *testing.T) {
	input := []byte("> [!NOTE]\n> Highlights information that users should take into account, even when skimming.\n")
	html, err := markdownToHTML(input, "")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(html, []byte(`class="markdown-alert markdown-alert-note"`)) {
		t.Errorf("expected alert div with note class, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte(`class="markdown-alert-title"`)) {
		t.Errorf("expected alert title class, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte(`data-source-line="1"`)) {
		t.Errorf("expected data-source-line=1 on alert, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte("<svg")) {
		t.Errorf("expected SVG icon in output, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte("Note")) {
		t.Errorf("expected title text 'Note' in output, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte("Highlights information")) {
		t.Errorf("expected body content, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte("</div>")) {
		t.Errorf("expected closing </div> tag, got:\n%s", html)
	}
}
func TestAlertWarningHTML(t *testing.T) {
	input := []byte("> [!WARNING]\n> Proceed with caution.\n")
	html, err := markdownToHTML(input, "")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(html, []byte(`markdown-alert-warning`)) {
		t.Errorf("expected warning class, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte("Warning")) {
		t.Errorf("expected title text 'Warning', got:\n%s", html)
	}
}
func TestAlertRegularBlockquoteFallback(t *testing.T) {
	input := []byte("> This is a regular blockquote.\n")
	html, err := markdownToHTML(input, "")
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Contains(html, []byte("markdown-alert")) {
		t.Errorf("should NOT render as alert, got:\n%s", html)
	}
	if !bytes.Contains(html, []byte("<blockquote")) {
		t.Errorf("expected <blockquote> for regular quote, got:\n%s", html)
	}
}
func TestAlertExtraTextFallback(t *testing.T) {
	input := []byte("> [!NOTE] Extra text here\n")
	html, err := markdownToHTML(input, "")
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Contains(html, []byte("markdown-alert")) {
		t.Errorf("should NOT render as alert when extra text follows, got:\n%s", html)
	}
}
