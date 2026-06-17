package main

import (
	"bytes"
	"strings"
	"testing"
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
