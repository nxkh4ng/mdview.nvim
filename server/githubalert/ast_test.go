package githubalert

import "testing"

func TestAlertNodeKind(t *testing.T) {
	n := &Alert{}
	if n.Kind() != AlertKind {
		t.Fatalf("expected AlertKind, got %v", n.Kind())
	}
}

func TestAlertTitleNodeKind(t *testing.T) {
	n := &AlertTitle{}
	if n.Kind() != AlertTitleKind {
		t.Fatalf("expected AlertTitleKind, got %v", n.Kind())
	}
}
