package routes

import (
  "context"
  "net/http"
  "strconv"
  "time"

  "github.com/go-chi/chi/v5"
  "gothicforge3/app/templates"
  "gothicforge3/internal/auth"
  "gothicforge3/internal/db"
  "gothicforge3/internal/env"
  "github.com/jackc/pgx/v5/pgxpool"
)

func init() {
  RegisterRoute(func(r chi.Router) {
    // List
    r.Get("/db/posts", func(w http.ResponseWriter, req *http.Request) {
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      pool, ok := requireDB(req, w)
      if !ok { return }
      rows, err := pool.Query(req.Context(), `SELECT id, title, body, to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') FROM posts ORDER BY id DESC LIMIT 50`)
      if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
      defer rows.Close()
      list := make([]templates.DBPostItem, 0, 32)
      for rows.Next() {
        var it templates.DBPostItem
        if err := rows.Scan(&it.ID, &it.Title, &it.Body, &it.CreatedAt); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
        list = append(list, it)
      }
      _ = templates.DBPostsList(list).Render(req.Context(), w)
    })

    // New form
    r.Get("/db/posts/new", func(w http.ResponseWriter, req *http.Request) {
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      _ = templates.DBPostsForm("/db/posts", nil, "Create").Render(req.Context(), w)
    })

    // Create
    r.Post("/db/posts", func(w http.ResponseWriter, req *http.Request) {
      if !requireJWTGuard(req) { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
      pool, ok := requireDB(req, w)
      if !ok { return }
      _ = req.ParseForm()
      title := req.FormValue("title")
      body := req.FormValue("body")
      if title == "" { http.Redirect(w, req, "/db/posts/new", http.StatusSeeOther); return }
      if _, err := pool.Exec(req.Context(), `INSERT INTO posts (title, body) VALUES ($1, $2)`, title, body); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
      http.Redirect(w, req, "/db/posts", http.StatusSeeOther)
    })

    // Edit form
    r.Get("/db/posts/{id}/edit", func(w http.ResponseWriter, req *http.Request) {
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      pool, ok := requireDB(req, w)
      if !ok { return }
      id, _ := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
      row := pool.QueryRow(req.Context(), `SELECT id, title, body, to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') FROM posts WHERE id=$1`, id)
      var it templates.DBPostItem
      if err := row.Scan(&it.ID, &it.Title, &it.Body, &it.CreatedAt); err != nil { http.NotFound(w, req); return }
      _ = templates.DBPostsForm("/db/posts/"+strconv.FormatInt(id,10), &it, "Update").Render(req.Context(), w)
    })

    // Update
    r.Post("/db/posts/{id}", func(w http.ResponseWriter, req *http.Request) {
      if !requireJWTGuard(req) { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
      pool, ok := requireDB(req, w)
      if !ok { return }
      _ = req.ParseForm()
      id, _ := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
      title := req.FormValue("title"); body := req.FormValue("body")
      if _, err := pool.Exec(req.Context(), `UPDATE posts SET title=$1, body=$2, updated_at=now() WHERE id=$3`, title, body, id); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
      http.Redirect(w, req, "/db/posts", http.StatusSeeOther)
    })

    // Delete
    r.Post("/db/posts/{id}/delete", func(w http.ResponseWriter, req *http.Request) {
      if !requireJWTGuard(req) { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
      pool, ok := requireDB(req, w)
      if !ok { return }
      id, _ := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
      if _, err := pool.Exec(req.Context(), `DELETE FROM posts WHERE id=$1`, id); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
      http.Redirect(w, req, "/db/posts", http.StatusSeeOther)
    })

    RegisterURL("/db/posts")
  })
}

func requireJWTGuard(r *http.Request) bool { _, err := auth.ReadAndVerifyCookie(r, "gf_jwt"); return err == nil }

// requireDB ensures DATABASE_URL is configured and a connection is established.
// It responds with 503 when missing or 500 when connect fails.
func requireDB(req *http.Request, w http.ResponseWriter) (*pgxpool.Pool, bool) {
  if env.Get("DATABASE_URL", "") == "" {
    http.Error(w, "database not configured", http.StatusServiceUnavailable)
    return nil, false
  }
  ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
  defer cancel()
  if err := db.Connect(ctx); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return nil, false
  }
  return db.Pool(), true
}
