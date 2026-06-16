package main

import (
	"testing"
)

// Benchmark render WITHOUT baseDir (use cache baseMd)
func BenchmarkRenderNoBaseDir(b *testing.B) {
	source := []byte("# Hello\n\nThis is a **test**.\n\n- item 1\n- item 2\n")
	for b.Loop() {
		markdownToHTML(source, "")
	}
}

// Benchmark render WITH baseDir (create new goldmark each time)
func BenchmarkRenderWithBaseDir(b *testing.B) {
	source := []byte("# Hello\n\nThis is a **test**.\n\n- item 1\n- item 2\n")
	for b.Loop() {
		markdownToHTML(source, "/home/user/project")
	}
}
