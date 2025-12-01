package templates

import (
  "context"
  "fmt"
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
    if err := writeHTML(w, "<section class=\"mx-auto max-w-6xl p-4\">"); err != nil {
      return fmt.Errorf("DBPostsList template (section start): %w", err)
    }
    if err := writeHTML(w, "<div class=\"flex justify-between items-center mb-4\"><h2 class=\"text-2xl font-bold\">Posts</h2><a class=\"btn btn-primary\" href=\"/db/posts/new\">New</a></div>"); err != nil {
      return fmt.Errorf("DBPostsList template (header): %w", err)
    }
    if err := writeHTML(w, "<div class=\"card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10\"><div class=\"card-body\">"); err != nil {
      return fmt.Errorf("DBPostsList template (card start): %w", err)
    }
    if len(items) == 0 {
      if err := writeHTML(w, "<p class=\"opacity-80\">No posts yet.</p>"); err != nil {
        return fmt.Errorf("DBPostsList template (empty message): %w", err)
      }
    } else {
      if err := writeHTML(w, "<ul class=\"menu\">"); err != nil {
        return fmt.Errorf("DBPostsList template (list start): %w", err)
      }
      for _, it := range items {
        if err := writeHTML(w, "<li><a href=\"/db/posts/" +  fmtInt(it.ID) + "/edit\">" + it.Title + "</a></li>"); err != nil {
          return fmt.Errorf("DBPostsList template (list item): %w", err)
        }
      }
      if err := writeHTML(w, "</ul>"); err != nil {
        return fmt.Errorf("DBPostsList template (list end): %w", err)
      }
    }
    if err := writeHTML(w, "</div></div></section>"); err != nil {
      return fmt.Errorf("DBPostsList template (section end): %w", err)
    }
    return nil
  })
  return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error { return LayoutSEO(SEO{Title: "Posts", Description: "DB posts", Canonical: "/db/posts"}).Render(templ.WithChildren(ctx, body), w) })
}

func DBPostsForm(action string, item *DBPostItem, submit string) templ.Component {
  body := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
    if err := writeHTML(w, "<section class=\"mx-auto max-w-xl p-4\"><div class=\"card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10\"><div class=\"card-body\">"); err != nil {
      return fmt.Errorf("DBPostsForm template (section start): %w", err)
    }
    if err := writeHTML(w, "<h2 class=\"card-title\">Post</h2>"); err != nil {
      return fmt.Errorf("DBPostsForm template (title): %w", err)
    }
    if err := writeHTML(w, "<form method=\"post\" action=\"" + action + "\" class=\"grid gap-3\">"); err != nil {
      return fmt.Errorf("DBPostsForm template (form start): %w", err)
    }
    title := ""
    bodyContent := ""
    if item != nil { title = item.Title; bodyContent = item.Body }
    if err := writeHTML(w, "<label class=\"form-control\"><span class=\"label-text\">Title</span><input class=\"input input-bordered\" name=\"title\" value=\"" + title + "\" required></label>"); err != nil {
      return fmt.Errorf("DBPostsForm template (title input): %w", err)
    }
    if err := writeHTML(w, "<label class=\"form-control\"><span class=\"label-text\">Body</span><textarea class=\"textarea textarea-bordered\" name=\"body\">" + bodyContent + "</textarea></label>"); err != nil {
      return fmt.Errorf("DBPostsForm template (body textarea): %w", err)
    }
    if err := writeHTML(w, "<button class=\"btn btn-primary\" type=\"submit\">" + submit + "</button>"); err != nil {
      return fmt.Errorf("DBPostsForm template (submit button): %w", err)
    }
    if err := writeHTML(w, "</form></div></div></section>"); err != nil {
      return fmt.Errorf("DBPostsForm template (section end): %w", err)
    }
    return nil
  })
  return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error { return LayoutSEO(SEO{Title: "Post", Description: "Post form", Canonical: "/db/posts/new"}).Render(templ.WithChildren(ctx, body), w) })
}
