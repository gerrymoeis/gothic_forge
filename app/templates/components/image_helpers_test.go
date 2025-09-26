package templates

import (
	"testing"
)

func TestSrcsetWithParam_NoQuery(t *testing.T) {
	src := "/img.jpg"
	widths := []int{320, 640}
	got := SrcsetWithParam(src, "w", widths)
	expected := "/img.jpg?w=320 320w, /img.jpg?w=640 640w"
	if got != expected {
		t.Fatalf("unexpected srcset: got %q, want %q", got, expected)
	}
}

func TestSrcsetWithParam_WithQuery(t *testing.T) {
	src := "/img.jpg?q=80"
	widths := []int{320, 640}
	got := SrcsetWithParam(src, "w", widths)
	expected := "/img.jpg?q=80&w=320 320w, /img.jpg?q=80&w=640 640w"
	if got != expected {
		t.Fatalf("unexpected srcset with query: got %q, want %q", got, expected)
	}
}

func TestSrcsetCF(t *testing.T) {
	src := "/media/a.jpg"
	widths := []int{320, 640}
	got := SrcsetCF(src, widths)
	expected := "/cdn-cgi/image/format=auto,quality=85,width=320/media/a.jpg 320w, /cdn-cgi/image/format=auto,quality=85,width=640/media/a.jpg 640w"
	if got != expected {
		t.Fatalf("unexpected cf srcset: got %q, want %q", got, expected)
	}
}
