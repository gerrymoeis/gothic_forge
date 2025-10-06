package templates

import (
  "context"
  "io"
  templ "github.com/a-h/templ"
)

type DBPostItem struct {
  ID int64
  Title string
  Body string
  CreatedAt string
}

// fmtInt is a tiny helper used by this pure-Go template file to avoid importing strconv everywhere
func fmtInt(v int64) string {
  // Avoid adding fmt/strconv imports to repeated template files; keep minimal
  // Implement simple conversion
  if v == 0 { return "0" }
  neg := false
  if v < 0 { neg = true; v = -v }
  var buf [20]byte
  i := len(buf)
  for v > 0 {
    i--
    buf[i] = byte('0' + v%10)
    v /= 10
  }
  if neg { i--; buf[i] = '-' }
  return string(buf[i:])
}

func DBPostsList(items []DBPostItem) templ.Component {
  body := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
    _, _ = io.WriteString(w, "<section class=\"mx-auto max-w-6xl p-4\">")
    _, _ = io.WriteString(w, "<div class=\"flex justify-between items-center mb-4\"><h2 class=\"text-2xl font-bold\">Posts</h2><a class=\"btn btn-primary\" href=\"/db/posts/new\">New</a></div>")
    _, _ = io.WriteString(w, "<div class=\"card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10\"><div class=\"card-body\">")
    if len(items) == 0 {
      _, _ = io.WriteString(w, "<p class=\"opacity-80\">No posts yet.</p>")
    } else {
      _, _ = io.WriteString(w, "<ul class=\"menu\">")
      for _, it := range items {
        _, _ = io.WriteString(w, "<li><a href=\"/db/posts/" +  fmtInt(it.ID) + "/edit\">" + it.Title + "</a></li>")
      }
      _, _ = io.WriteString(w, "</ul>")
    }
    _, _ = io.WriteString(w, "</div></div></section>")
    return nil
  })
  return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error { return LayoutSEO(SEO{Title: "Posts", Description: "DB posts", Canonical: "/db/posts"}).Render(templ.WithChildren(ctx, body), w) })
}

func DBPostsForm(action string, item *DBPostItem, submit string) templ.Component {
  body := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
    _, _ = io.WriteString(w, "<section class=\"mx-auto max-w-xl p-4\"><div class=\"card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10\"><div class=\"card-body\">")
    _, _ = io.WriteString(w, "<h2 class=\"card-title\">Post</h2>")
    _, _ = io.WriteString(w, "<form method=\"post\" action=\"" + action + "\" class=\"grid gap-3\">")
    title := ""
    body := ""
    if item != nil { title = item.Title; body = item.Body }
    _, _ = io.WriteString(w, "<label class=\"form-control\"><span class=\"label-text\">Title</span><input class=\"input input-bordered\" name=\"title\" value=\"" + title + "\" required></label>")
    _, _ = io.WriteString(w, "<label class=\"form-control\"><span class=\"label-text\">Body</span><textarea class=\"textarea textarea-bordered\" name=\"body\">" + body + "</textarea></label>")
    _, _ = io.WriteString(w, "<button class=\"btn btn-primary\" type=\"submit\">" + submit + "</button>")
    _, _ = io.WriteString(w, "</form></div></div></section>")
    return nil
  })
  return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error { return LayoutSEO(SEO{Title: "Post", Description: "Post form", Canonical: "/db/posts/new"}).Render(templ.WithChildren(ctx, body), w) })
}
