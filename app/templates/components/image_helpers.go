package templates

import (
	"strconv"
	"strings"
)

// SrcsetWithParam builds a simple srcset string by appending a query parameter (e.g., ?w=WIDTH)
// to the provided src for each width, producing items like: "src?w=320 320w, src?w=640 640w".
func SrcsetWithParam(src, param string, widths []int) string {
	if src == "" || len(widths) == 0 {
		return ""
	}
	var b strings.Builder
	for i, w := range widths {
		if i > 0 {
			b.WriteString(", ")
		}
		sep := "?"
		if strings.Contains(src, "?") {
			sep = "&"
		}
		b.WriteString(src)
		b.WriteString(sep)
		b.WriteString(param)
		b.WriteString("=")
		b.WriteString(strconv.Itoa(w))
		b.WriteString(" ")
		b.WriteString(strconv.Itoa(w))
		b.WriteString("w")
	}
	return b.String()
}

// SrcsetCF builds a Cloudflare Image Resizing srcset using the /cdn-cgi/image endpoint.
// Example item: "/cdn-cgi/image/format=auto,quality=85,width=320/path/to/img.jpg 320w"
// Works with absolute or relative src values.
func SrcsetCF(src string, widths []int) string {
	if src == "" || len(widths) == 0 {
		return ""
	}
	var b strings.Builder
	cleaned := strings.TrimPrefix(src, "/")
	for i, w := range widths {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("/cdn-cgi/image/format=auto,quality=85,width=")
		b.WriteString(strconv.Itoa(w))
		b.WriteString("/")
		b.WriteString(cleaned)
		b.WriteString(" ")
		b.WriteString(strconv.Itoa(w))
		b.WriteString("w")
	}
	return b.String()
}
