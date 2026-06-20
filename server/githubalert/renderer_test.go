package githubalert

import (
	"strings"
	"testing"

	"github.com/yuin/goldmark"
)

func render(t *testing.T, source string) string {
	t.Helper()
	md := goldmark.New(
		goldmark.WithExtensions(
			&Extension{},
		),
	)
	var buf strings.Builder
	if err := md.Convert([]byte(source), &buf); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestRenderNote(t *testing.T) {
	html := render(t, "> [!NOTE]\n> content\n")
	if !strings.Contains(html, `class="markdown-alert markdown-alert-note"`) {
		t.Errorf("expected markdown-alert-note class, got:\n%s", html)
	}
	if !strings.Contains(html, `class="markdown-alert-title"`) {
		t.Errorf("expected markdown-alert-title element, got:\n%s", html)
	}
	if !strings.Contains(html, "Note") {
		t.Errorf("expected 'Note' title text, got:\n%s", html)
	}
	if !strings.Contains(html, "content") {
		t.Errorf("expected 'content' body text, got:\n%s", html)
	}
	if !strings.Contains(html, "<svg") {
		t.Errorf("expected SVG icon, got:\n%s", html)
	}
}

func TestRenderTip(t *testing.T) {
	html := render(t, "> [!TIP]\n> helpful\n")
	if !strings.Contains(html, `class="markdown-alert markdown-alert-tip"`) {
		t.Errorf("expected markdown-alert-tip class, got:\n%s", html)
	}
	if !strings.Contains(html, "Tip") {
		t.Errorf("expected 'Tip' title text, got:\n%s", html)
	}
}

func TestRenderImportant(t *testing.T) {
	html := render(t, "> [!IMPORTANT]\n> important info\n")
	if !strings.Contains(html, `class="markdown-alert markdown-alert-important"`) {
		t.Errorf("expected markdown-alert-important class, got:\n%s", html)
	}
	if !strings.Contains(html, "Important") {
		t.Errorf("expected 'Important' title text, got:\n%s", html)
	}
}

func TestRenderWarning(t *testing.T) {
	html := render(t, "> [!WARNING]\n> be careful\n")
	if !strings.Contains(html, `class="markdown-alert markdown-alert-warning"`) {
		t.Errorf("expected markdown-alert-warning class, got:\n%s", html)
	}
	if !strings.Contains(html, "Warning") {
		t.Errorf("expected 'Warning' title text, got:\n%s", html)
	}
}

func TestRenderCaution(t *testing.T) {
	html := render(t, "> [!CAUTION]\n> danger\n")
	if !strings.Contains(html, `class="markdown-alert markdown-alert-caution"`) {
		t.Errorf("expected markdown-alert-caution class, got:\n%s", html)
	}
	if !strings.Contains(html, "Caution") {
		t.Errorf("expected 'Caution' title text, got:\n%s", html)
	}
}

func TestRenderCaseInsensitive(t *testing.T) {
	html := render(t, "> [!Note]\n> content\n")
	if !strings.Contains(html, `markdown-alert-note`) {
		t.Errorf("expected lowercase kind in class, got:\n%s", html)
	}
}

func TestRenderBodyMarkdown(t *testing.T) {
	html := render(t, "> [!NOTE]\n> **bold** and `code`\n")
	if !strings.Contains(html, "<strong>bold</strong>") {
		t.Errorf("expected bold rendering, got:\n%s", html)
	}
	if !strings.Contains(html, "<code>code</code>") {
		t.Errorf("expected code rendering, got:\n%s", html)
	}
}

func TestRenderBodyList(t *testing.T) {
	html := render(t, "> [!TIP]\n> - item 1\n> - item 2\n")
	if !strings.Contains(html, "<ul") {
		t.Errorf("expected list inside alert, got:\n%s", html)
	}
	if !strings.Contains(html, "item 1") {
		t.Errorf("expected list item, got:\n%s", html)
	}
	if !strings.Contains(html, "item 2") {
		t.Errorf("expected second list item, got:\n%s", html)
	}
}

func TestRenderBodyCodeBlock(t *testing.T) {
	html := render(t, "> [!NOTE]\n> ```go\n> fmt.Println(\"hello\")\n> ```\n")
	if !strings.Contains(html, "<pre") {
		t.Errorf("expected code block inside alert, got:\n%s", html)
	}
	if !strings.Contains(html, "hello") {
		t.Errorf("expected code content, got:\n%s", html)
	}
}

func TestRenderEmptyAlert(t *testing.T) {
	html := render(t, "> [!NOTE]\n")
	if !strings.Contains(html, `class="markdown-alert markdown-alert-note"`) {
		t.Errorf("expected alert with no body, got:\n%s", html)
	}
	if !strings.Contains(html, "Note") {
		t.Errorf("expected title, got:\n%s", html)
	}
}

func TestRenderFallbackBlockquote(t *testing.T) {
	// Invalid kind should render as blockquote, not alert
	html := render(t, "> [!INVALID]\n> content\n")
	if strings.Contains(html, "markdown-alert") {
		t.Errorf("expected NO markdown-alert for invalid kind, got:\n%s", html)
	}
	if !strings.Contains(html, "<blockquote") {
		t.Errorf("expected blockquote for invalid kind, got:\n%s", html)
	}
}

func TestRenderFallbackExtraText(t *testing.T) {
	html := render(t, "> [!NOTE] extra\n> content\n")
	if strings.Contains(html, "markdown-alert") {
		t.Errorf("expected NO alert when extra text after bracket, got:\n%s", html)
	}
}

func TestRenderMultipleAlerts(t *testing.T) {
	html := render(t, "> [!NOTE]\n> first\n\n> [!WARNING]\n> second\n")
	if !strings.Contains(html, "markdown-alert-note") {
		t.Errorf("expected first alert, got:\n%s", html)
	}
	if !strings.Contains(html, "markdown-alert-warning") {
		t.Errorf("expected second alert, got:\n%s", html)
	}
}

func TestRenderClosesDiv(t *testing.T) {
	html := render(t, "> [!NOTE]\n> content\n")
	if !strings.Contains(html, "</div>") {
		t.Errorf("expected closing </div> tag, got:\n%s", html)
	}
	if !strings.Contains(html, "<div") {
		t.Errorf("expected opening <div> tag, got:\n%s", html)
	}
}

func TestRenderClosingTag(t *testing.T) {
	html := render(t, "> [!NOTE]\n> content\n")
	if !strings.Contains(html, "</div>") {
		t.Errorf("expected closing </div> tag, got:\n%s", html)
	}
}
