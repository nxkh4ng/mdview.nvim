package githubalert

import (
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// Helper: parse markdown và trả về document AST
func parse(t *testing.T, source string) ast.Node {
	t.Helper()
	md := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithBlockParsers(
				util.Prioritized(newAlertParser(), 799),
				util.Prioritized(newAlertTitleParser(), 0),
			),
		),
	)
	reader := text.NewReader([]byte(source))
	doc := md.Parser().Parse(reader, parser.WithContext(parser.NewContext()))
	return doc
}

// Helper: tìm Alert node đầu tiên trong AST
func findAlert(doc ast.Node) *Alert {
	var found *Alert
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if alert, ok := n.(*Alert); ok {
			found = alert
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return found
}

// 5 loại alert
func TestParserDetectsNote(t *testing.T) {
	doc := parse(t, "> [!NOTE]\n> content\n")
	alert := findAlert(doc)
	if alert == nil {
		t.Fatal("expected Alert node")
	}
	kind, _ := alert.AttributeString("kind")
	if kind != "note" {
		t.Errorf("expected kind=note, got %v", kind)
	}
}

func TestParserDetectsTip(t *testing.T) {
	doc := parse(t, "> [!TIP]\n> content\n")
	alert := findAlert(doc)
	if alert == nil {
		t.Fatal("expected Alert node")
	}
	kind, _ := alert.AttributeString("kind")
	if kind != "tip" {
		t.Errorf("expected kind=tip, got %v", kind)
	}
}

func TestParserDetectsImportant(t *testing.T) {
	doc := parse(t, "> [!IMPORTANT]\n> content\n")
	alert := findAlert(doc)
	if alert == nil {
		t.Fatal("expected Alert node")
	}
	kind, _ := alert.AttributeString("kind")
	if kind != "important" {
		t.Errorf("expected kind=important, got %v", kind)
	}
}

func TestParserDetectsWarning(t *testing.T) {
	doc := parse(t, "> [!WARNING]\n> content\n")
	alert := findAlert(doc)
	if alert == nil {
		t.Fatal("expected Alert node")
	}
	kind, _ := alert.AttributeString("kind")
	if kind != "warning" {
		t.Errorf("expected kind=warning, got %v", kind)
	}
}

func TestParserDetectsCaution(t *testing.T) {
	doc := parse(t, "> [!CAUTION]\n> content\n")
	alert := findAlert(doc)
	if alert == nil {
		t.Fatal("expected Alert node")
	}
	kind, _ := alert.AttributeString("kind")
	if kind != "caution" {
		t.Errorf("expected kind=caution, got %v", kind)
	}
}

// Case insensitive
func TestParserCaseInsensitive(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"> [!NOTE]\n> a\n"},
		{"> [!note]\n> a\n"},
		{"> [!Note]\n> a\n"},
		{"> [!NoTe]\n> a\n"},
	}
	for _, tt := range tests {
		doc := parse(t, tt.input)
		alert := findAlert(doc)
		if alert == nil {
			t.Errorf("expected Alert for input: %s", tt.input)
		}
	}
}

// Cấu trúc AST chính xác
func TestParserASTStructure(t *testing.T) {
	doc := parse(t, "> [!TIP]\n> hello\n")
	alert := findAlert(doc)
	if alert == nil {
		t.Fatal("expected Alert node")
	}

	// Alert phải có 2 children: AlertTitle + Paragraph
	if alert.ChildCount() != 2 {
		t.Fatalf("expected 2 children, got %d", alert.ChildCount())
	}

	// Child 0: AlertTitle
	title, ok := alert.FirstChild().(*AlertTitle)
	if !ok {
		t.Fatal("first child should be AlertTitle")
	}
	kind, _ := title.AttributeString("kind")
	if kind != "tip" {
		t.Errorf("expected kind=tip on title, got %v", kind)
	}

	// Child 1: Paragraph (hello)
	para := alert.FirstChild().NextSibling()
	if para == nil || para.Kind() != ast.KindParagraph {
		t.Error("second child should be Paragraph")
	}
}

// Fallback: invalid kind → blockquote thường
func TestParserFallbackInvalidKind(t *testing.T) {
	doc := parse(t, "> [!INVALID]\n> content\n")
	alert := findAlert(doc)
	if alert != nil {
		t.Error("expected NO Alert for [!INVALID]")
	}
}

// Fallback: extra text sau bracket
func TestParserFallbackExtraText(t *testing.T) {
	doc := parse(t, "> [!NOTE] extra text\n> content\n")
	alert := findAlert(doc)
	if alert != nil {
		t.Error("expected NO Alert for '> [!NOTE] extra text'")
	}
}

// Fallback: folding symbols
func TestParserFallbackFoldingPlus(t *testing.T) {
	doc := parse(t, "> [!NOTE]+\n> content\n")
	alert := findAlert(doc)
	if alert != nil {
		t.Error("expected NO Alert for '> [!NOTE]+'")
	}
}

func TestParserFallbackFoldingMinus(t *testing.T) {
	doc := parse(t, "> [!WARNING]-\n> content\n")
	alert := findAlert(doc)
	if alert != nil {
		t.Error("expected NO Alert for '> [!WARNING]-'")
	}
}

// Blockquote thường vẫn hoạt động
func TestParserRegularBlockquote(t *testing.T) {
	doc := parse(t, "> regular blockquote\n")
	alert := findAlert(doc)
	if alert != nil {
		t.Error("expected NO Alert for regular blockquote")
	}
}

// Nội dung markdown bên trong alert
func TestParserContentWithBold(t *testing.T) {
	doc := parse(t, "> [!NOTE]\n> **bold** content\n")
	alert := findAlert(doc)
	if alert == nil {
		t.Fatal("expected Alert node")
	}
	// Kiểm tra paragraph có Emphasis bên trong
	para := alert.LastChild()
	if para == nil || para.Kind() != ast.KindParagraph {
		t.Fatal("expected Paragraph")
	}
	hasEmphasis := false
	ast.Walk(para, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if n.Kind() == ast.KindEmphasis && entering {
			hasEmphasis = true
		}
		return ast.WalkContinue, nil
	})
	if !hasEmphasis {
		t.Error("expected Emphasis (bold) inside paragraph")
	}
}

// Nhiều alert liên tiếp
func TestParserMultipleAlerts(t *testing.T) {
	input := "> [!NOTE]\n> first\n\n> [!WARNING]\n> second\n"
	doc := parse(t, input)

	count := 0
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if _, ok := n.(*Alert); ok && entering {
			count++
		}
		return ast.WalkContinue, nil
	})
	if count != 2 {
		t.Errorf("expected 2 Alerts, got %d", count)
	}
}

// Empty alert (chỉ 1 dòng)
func TestParserEmptyAlert(t *testing.T) {
	doc := parse(t, "> [!NOTE]\n")
	alert := findAlert(doc)
	if alert == nil {
		t.Fatal("expected Alert node")
	}
	if alert.ChildCount() != 1 {
		t.Fatalf("expected 1 child (AlertTitle only), got %d", alert.ChildCount())
	}
}
