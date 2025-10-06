package cmd

import (
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "time"

    "gothicforge3/internal/execx"
    "github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
    Use:   "add",
    Short: "Scaffold features in app/ (page, component, auth, oauth, db, module, resource)",
    Args:  cobra.MinimumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        banner()
        kind := strings.ToLower(args[0])
        var name string
        if kind == "page" || kind == "component" || kind == "oauth" || kind == "db" || kind == "module" || kind == "crud" || kind == "resource" || kind == "migration" || kind == "cruddb" {
            if len(args) < 2 {
                fmt.Println("Usage:")
                fmt.Println("  gforge add page <name>")
                fmt.Println("  gforge add component <name>")
                fmt.Println("  gforge add auth")
                fmt.Println("  gforge add oauth <provider>")
                fmt.Println("  gforge add db <name>")
                fmt.Println("  gforge add module <name>")
                fmt.Println("  gforge add crud <name>")
                fmt.Println("  gforge add resource <Name> <field:type> [field:type ...]")
                fmt.Println("  gforge add migration <name>")
                fmt.Println("  gforge add cruddb <Name> <field:type> [field:type ...]")
                return nil
            }
            name = args[1]
            if !isValidName(name) {
                return fmt.Errorf("invalid name: %s (use letters, numbers, dash, underscore)", name)
            }
        }
        switch kind {
        case "page":
            return scaffoldPage(name)
        case "component":
            return scaffoldComponent(name)
        case "auth":
            return scaffoldAuth()
        case "oauth":
            return scaffoldOAuth(name)
        case "db":
            return scaffoldDB(name)
        case "module":
            return scaffoldModule(name)
        case "crud":
            return scaffoldCRUD(name)
        case "resource":
            fields := []string{}
            if len(args) > 2 { fields = args[2:] }
            return scaffoldResource(name, fields)
        case "migration":
            return scaffoldMigration(name)
        case "cruddb":
            fields := []string{}
            if len(args) > 2 { fields = args[2:] }
            return scaffoldCRUDDB(name, fields)
        default:
            fmt.Println("Usage:")
            fmt.Println("  gforge add page <name>")
            fmt.Println("  gforge add component <name>")
            fmt.Println("  gforge add auth")
            fmt.Println("  gforge add oauth <provider>")
            fmt.Println("  gforge add db <name>")
            fmt.Println("  gforge add module <name>")
            fmt.Println("  gforge add crud <name>")
            fmt.Println("  gforge add resource <Name> <field:type> [field:type ...]")
            fmt.Println("  gforge add migration <name>")
            fmt.Println("  gforge add cruddb <Name> <field:type> [field:type ...]")
            return nil
        }
    },
}

// scaffoldOAuth creates placeholder OAuth routes for a provider.
func scaffoldOAuth(provider string) error {
    keb := kebabCase(provider)
    routePath := filepath.Join("app", "routes", fmt.Sprintf("oauth_%s.go", keb))
    routeSrc := fmt.Sprintf(`package routes

import (
    "net/http"
    "github.com/go-chi/chi/v5"
)

func init() {
    RegisterRoute(func(r chi.Router) {
        r.Get("/oauth/%[1]s/start", func(w http.ResponseWriter, req *http.Request) {
            w.Header().Set("Content-Type", "text/plain; charset=utf-8")
            w.WriteHeader(http.StatusNotImplemented)
            _, _ = w.Write([]byte("OAuth %[1]s start not implemented"))
        })
        r.Get("/oauth/%[1]s/callback", func(w http.ResponseWriter, req *http.Request) {
            w.Header().Set("Content-Type", "text/plain; charset=utf-8")
            w.WriteHeader(http.StatusNotImplemented)
            _, _ = w.Write([]byte("OAuth %[1]s callback not implemented"))
        })
        RegisterURL("/oauth/%[1]s/start")
    })
}
`, keb)
    if err := execx.WriteFileIfMissing(routePath, []byte(routeSrc), 0o644); err != nil { return err }
    fmt.Printf("Added OAuth placeholder: /oauth/%s/start, /oauth/%s/callback\n", keb, keb)
    fmt.Printf("  - %s\n", routePath)
    return nil
}

// dbFieldDesc describes a DB field for scaffolding.
type dbFieldDesc struct { Name, SQLType, GoName string }

// scaffoldCRUDDB generates a DB-backed CRUD feature under /db/<plural> using pgxpool and internal/db.
// Example: gforge add cruddb Post title:string body:text
func scaffoldCRUDDB(name string, fields []string) error {
    if len(fields) == 0 {
        return fmt.Errorf("cruddb requires at least one <field:type>")
    }
    keb := kebabCase(name)
    pas := pascalCase(name)
    plural := strings.ToLower(keb) + "s"
    table := plural

    fds := make([]dbFieldDesc, 0, len(fields))
    for _, f := range fields {
        parts := strings.SplitN(strings.TrimSpace(f), ":", 2)
        if parts[0] == "" { continue }
        nm := strings.ToLower(parts[0])
        tp := "text"
        if len(parts) == 2 {
            switch strings.ToLower(strings.TrimSpace(parts[1])) {
            case "string", "text": tp = "text"
            case "int", "integer": tp = "integer"
            case "bigint": tp = "bigint"
            case "bool", "boolean": tp = "boolean"
            case "float", "double", "doubleprecision": tp = "double precision"
            case "date": tp = "date"
            case "timestamp", "timestamptz": tp = "timestamptz"
            default: tp = "text"
            }
        }
        fds = append(fds, dbFieldDesc{Name: nm, SQLType: tp, GoName: pascalCase(nm)})
    }

    // 1) Migration
    cols := make([]string, 0, len(fds)+3)
    cols = append(cols, "  id bigserial PRIMARY KEY")
    for _, fd := range fds { cols = append(cols, fmt.Sprintf("  %s %s NOT NULL", fd.Name, fd.SQLType)) }
    cols = append(cols, "  created_at timestamptz DEFAULT now()")
    cols = append(cols, "  updated_at timestamptz DEFAULT now()")
    up := fmt.Sprintf("CREATE TABLE %s (\n%s\n);\n", table, strings.Join(cols, ",\n"))
    down := fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", table)
    mig := fmt.Sprintf("-- +goose Up\n%s\n-- +goose Down\n%s", up, down)
    mdir := filepath.Join("app", "db", "migrations")
    if err := os.MkdirAll(mdir, 0o755); err != nil { return err }
    ts := time.Now().UTC().Format("20060102150405")
    mfile := filepath.Join(mdir, fmt.Sprintf("%s_create_%s.sql", ts, table))
    if err := os.WriteFile(mfile, []byte(mig), 0o644); err != nil { return err }

    // 2) Template
    // Struct fields
    var structBuf strings.Builder
    for _, fd := range fds { structBuf.WriteString(fmt.Sprintf("  %s string\n", fd.GoName)) }
    // Form controls
    var formBuf strings.Builder
    for _, fd := range fds {
        label := fd.GoName
        if fd.SQLType == "text" && (fd.Name == "body" || fd.Name == "description" || fd.Name == "content") {
            formBuf.WriteString(fmt.Sprintf("        _, _ = io.WriteString(w, \"<label class=\\\"form-control\\\"><span class=\\\"label-text\\\">%s</span><textarea class=\\\"textarea textarea-bordered\\\" name=\\\"%s\\\">\" + item.%s + \"</textarea></label>\")\n", label, fd.Name, fd.GoName))
        } else if fd.SQLType == "text" {
            formBuf.WriteString(fmt.Sprintf("        _, _ = io.WriteString(w, \"<label class=\\\"form-control\\\"><span class=\\\"label-text\\\">%s</span><textarea class=\\\"textarea textarea-bordered\\\" name=\\\"%s\\\">\" + item.%s + \"</textarea></label>\")\n", label, fd.Name, fd.GoName))
        } else {
            formBuf.WriteString(fmt.Sprintf("        _, _ = io.WriteString(w, \"<label class=\\\"form-control\\\"><span class=\\\"label-text\\\">%s</span><input class=\\\"input input-bordered\\\" name=\\\"%s\\\" value=\\\"\" + item.%s + \"\\\" required></label>\")\n", label, fd.Name, fd.GoName))
        }
    }
    // List link text uses first field
    displayField := fds[0].GoName
    tmplPath := filepath.Join("app", "templates", fmt.Sprintf("db_%s.go", table))
    tmplSrc := fmt.Sprintf(`package templates

import (
  "context"
  "io"
  templ "github.com/a-h/templ"
)

type DB%[1]sItem struct {
  ID int64
%[2]s  CreatedAt string
}

// fmtInt helper
func fmtInt(v int64) string { if v==0 { return "0" }; neg:=v<0; if neg { v=-v }; var b [20]byte; i:=len(b); for v>0 { i--; b[i]=byte('0'+v%%10); v/=10 }; if neg { i--; b[i]='-'}; return string(b[i:]) }

func DB%[1]sList(items []DB%[1]sItem) templ.Component {
  body := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
    _, _ = io.WriteString(w, "<section class=\"mx-auto max-w-6xl p-4\">")
    _, _ = io.WriteString(w, "<div class=\"flex justify-between items-center mb-4\"><h2 class=\"text-2xl font-bold\">%[3]s</h2><a class=\"btn btn-primary\" href=\"/db/%[4]s/new\">New</a></div>")
    _, _ = io.WriteString(w, "<div class=\"card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10\"><div class=\"card-body\">")
    if len(items) == 0 {
      _, _ = io.WriteString(w, "<p class=\"opacity-80\">No items yet.</p>")
    } else {
      _, _ = io.WriteString(w, "<ul class=\"menu\">")
      for _, it := range items {
        _, _ = io.WriteString(w, "<li><a href=\"/db/%[4]s/" +  fmtInt(it.ID) + "/edit\">" + it.%[5]s + "</a></li>")
      }
      _, _ = io.WriteString(w, "</ul>")
    }
    _, _ = io.WriteString(w, "</div></div></section>")
    return nil
  })
  return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error { return LayoutSEO(SEO{Title: "%[3]s", Description: "%[3]s list", Canonical: "/db/%[4]s"}).Render(templ.WithChildren(ctx, body), w) })
}

func DB%[1]sForm(action string, item *DB%[1]sItem, submit string) templ.Component {
  if item == nil { item = &DB%[1]sItem{} }
  body := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
    _, _ = io.WriteString(w, "<section class=\"mx-auto max-w-xl p-4\"><div class=\"card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10\"><div class=\"card-body\">")
    _, _ = io.WriteString(w, "<h2 class=\"card-title\">%[3]s</h2>")
    _, _ = io.WriteString(w, "<form method=\"post\" action=\"" + action + "\" class=\"grid gap-3\">")
%[6]s    _, _ = io.WriteString(w, "<button class=\"btn btn-primary\" type=\"submit\">" + submit + "</button>")
    _, _ = io.WriteString(w, "</form></div></div></section>")
    return nil
  })
  return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error { return LayoutSEO(SEO{Title: "%[3]s", Description: "%[3]s form", Canonical: "/db/%[4]s/new"}).Render(templ.WithChildren(ctx, body), w) })
}
`, pas, structBuf.String(), pas, table, displayField, formBuf.String())
    if err := execx.WriteFileIfMissing(tmplPath, []byte(tmplSrc), 0o644); err != nil { return err }

    // 3) Routes
    // Build INSERT/UPDATE SQL pieces
    colNames := make([]string, 0, len(fds))
    valExprs := make([]string, 0, len(fds))
    setExprs := make([]string, 0, len(fds))
    for i, fd := range fds {
        colNames = append(colNames, fd.Name)
        p := fmt.Sprintf("$%d", i+1)
        if fd.SQLType == "text" {
            valExprs = append(valExprs, p)
            setExprs = append(setExprs, fmt.Sprintf("%s=%s", fd.Name, p))
        } else {
            valExprs = append(valExprs, fmt.Sprintf("CAST(%s AS %s)", p, fd.SQLType))
            setExprs = append(setExprs, fmt.Sprintf("%s=CAST(%s AS %s)", fd.Name, p, fd.SQLType))
        }
    }
    // List scan targets (ID + first field)
    var listSelect, listScan string
    if len(fds) > 0 {
        listSelect = fmt.Sprintf("id, (%s)::text, to_char(created_at, 'YYYY-MM-DD\"T\"HH24:MI:SS\"Z\"')", fds[0].Name)
        listScan = fmt.Sprintf("&it.ID, &it.%s, &it.CreatedAt", fds[0].GoName)
    } else {
        listSelect = "id, to_char(created_at, 'YYYY-MM-DD\"T\"HH24:MI:SS\"Z\"')"
        listScan = "&it.ID, &it.CreatedAt"
    }
    // Edit select and scan (all fields as text)
    selCols := make([]string, 0, len(fds))
    scanTargets := make([]string, 0, len(fds))
    for _, fd := range fds {
        selCols = append(selCols, fmt.Sprintf("(%s)::text", fd.Name))
        scanTargets = append(scanTargets, fmt.Sprintf("&it.%s", fd.GoName))
    }
    editSelect := fmt.Sprintf("id, %s, to_char(created_at, 'YYYY-MM-DD\"T\"HH24:MI:SS\"Z\"')", strings.Join(selCols, ", "))
    editScan := fmt.Sprintf("&it.ID, %s, &it.CreatedAt", strings.Join(scanTargets, ", "))

    routePath := filepath.Join("app", "routes", fmt.Sprintf("db_%s.go", table))
    routeSrc := fmt.Sprintf(`package routes

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
)

func init() {
  RegisterRoute(func(r chi.Router) {
    // List
    r.Get("/db/%[4]s", func(w http.ResponseWriter, req *http.Request) {
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      if env.Get("DATABASE_URL", "") == "" { http.Error(w, "database not configured", http.StatusServiceUnavailable); return }
      ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second); defer cancel()
      if err := db.Connect(ctx); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
      rows, err := db.Pool().Query(req.Context(), "SELECT %[6]s FROM %[3]s ORDER BY id DESC LIMIT 50")
      if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
      defer rows.Close()
      list := make([]templates.DB%[1]sItem, 0, 32)
      for rows.Next() {
        var it templates.DB%[1]sItem
        if err := rows.Scan(%[7]s); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
        list = append(list, it)
      }
      _ = templates.DB%[1]sList(list).Render(req.Context(), w)
    })

    // New form
    r.Get("/db/%[4]s/new", func(w http.ResponseWriter, req *http.Request) {
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      _ = templates.DB%[1]sForm("/db/%[4]s", nil, "Create").Render(req.Context(), w)
    })

    // Create
    r.Post("/db/%[4]s", func(w http.ResponseWriter, req *http.Request) {
      if _, err := auth.ReadAndVerifyCookie(req, "gf_jwt"); err != nil { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
      if env.Get("DATABASE_URL", "") == "" { http.Error(w, "database not configured", http.StatusServiceUnavailable); return }
      ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second); defer cancel()
      if err := db.Connect(ctx); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
      _ = req.ParseForm()
%[8]s      if _, err := db.Pool().Exec(req.Context(), "INSERT INTO %[3]s (%[9]s) VALUES (%[10]s)", %[11]s); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
      http.Redirect(w, req, "/db/%[4]s", http.StatusSeeOther)
    })

    // Edit form
    r.Get("/db/%[4]s/{id}/edit", func(w http.ResponseWriter, req *http.Request) {
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      if env.Get("DATABASE_URL", "") == "" { http.Error(w, "database not configured", http.StatusServiceUnavailable); return }
      ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second); defer cancel()
      if err := db.Connect(ctx); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
      id, _ := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
      row := db.Pool().QueryRow(req.Context(), "SELECT %[12]s FROM %[3]s WHERE id=$1", id)
      var it templates.DB%[1]sItem
      if err := row.Scan(%[13]s); err != nil { http.NotFound(w, req); return }
      _ = templates.DB%[1]sForm("/db/%[4]s/"+strconv.FormatInt(id,10), &it, "Update").Render(req.Context(), w)
    })

    // Update
    r.Post("/db/%[4]s/{id}", func(w http.ResponseWriter, req *http.Request) {
      if _, err := auth.ReadAndVerifyCookie(req, "gf_jwt"); err != nil { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
      if env.Get("DATABASE_URL", "") == "" { http.Error(w, "database not configured", http.StatusServiceUnavailable); return }
      ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second); defer cancel()
      if err := db.Connect(ctx); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
      _ = req.ParseForm()
      id, _ := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
%[14]s      if _, err := db.Pool().Exec(req.Context(), "UPDATE %[3]s SET %[15]s, updated_at=now() WHERE id=$%[16]d", %[17]s, id); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
      http.Redirect(w, req, "/db/%[4]s", http.StatusSeeOther)
    })

    // Delete
    r.Post("/db/%[4]s/{id}/delete", func(w http.ResponseWriter, req *http.Request) {
      if _, err := auth.ReadAndVerifyCookie(req, "gf_jwt"); err != nil { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
      if env.Get("DATABASE_URL", "") == "" { http.Error(w, "database not configured", http.StatusServiceUnavailable); return }
      ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second); defer cancel()
      if err := db.Connect(ctx); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
      id, _ := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
      if _, err := db.Pool().Exec(req.Context(), "DELETE FROM %[3]s WHERE id=$1", id); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
      http.Redirect(w, req, "/db/%[4]s", http.StatusSeeOther)
    })

    RegisterURL("/db/%[4]s")
  })
}
`, pas, keb, table, plural, displayField,
        listSelect, listScan,
        buildFormRead(fds),
        strings.Join(colNames, ", "), strings.Join(valExprs, ", "), strings.Join(formArgList(fds), ", "),
        editSelect, editScan,
        buildFormRead(fds), strings.Join(setExprs, ", "), len(fds)+1, strings.Join(formArgList(fds), ", "))
    if err := execx.WriteFileIfMissing(routePath, []byte(routeSrc), 0o644); err != nil { return err }

    fmt.Printf("Added DB CRUD: /db/%s (migration + routes + templates)\n", table)
    fmt.Printf("  - %s\n", mfile)
    fmt.Printf("  - %s\n", tmplPath)
    fmt.Printf("  - %s\n", routePath)
    return nil
}

func buildFormRead(fds []dbFieldDesc) string {
    if len(fds) == 0 { return "" }
    var b strings.Builder
    for _, fd := range fds {
        b.WriteString(fmt.Sprintf("      %s := req.FormValue(\"%s\")\n", fd.Name, fd.Name))
    }
    return b.String()
}

func formArgList(fds []dbFieldDesc) []string {
    out := make([]string, 0, len(fds))
    for _, fd := range fds { out = append(out, fd.Name) }
    return out
}

// scaffoldMigration creates a timestamped goose SQL migration file.
// Example: gforge add migration create_posts
func scaffoldMigration(name string) error {
    keb := kebabCase(name)
    dir := filepath.Join("app", "db", "migrations")
    if err := os.MkdirAll(dir, 0o755); err != nil { return err }
    ts := time.Now().UTC().Format("20060102150405")
    file := filepath.Join(dir, fmt.Sprintf("%s_%s.sql", ts, keb))
    content := "-- +goose Up\n-- Write your UP migration here\n\n-- +goose Down\n-- Write your DOWN migration here\n"
    if err := os.WriteFile(file, []byte(content), 0o644); err != nil { return err }
    fmt.Printf("Added migration: %s\n", filepath.Base(file))
    fmt.Printf("  - %s\n", file)
    return nil
}

// scaffoldResource creates a page and a timestamped SQL migration for a basic resource.
// Example: gforge add resource Post title:string body:text
func scaffoldResource(name string, fields []string) error {
    // Page and route
    if err := scaffoldPage(name); err != nil { return err }

    // Migration skeleton for Phase 2 (goose-style markers)
    keb := kebabCase(name)
    table := strings.ToLower(keb) + "s" // naive pluralization
    cols := make([]string, 0, len(fields))
    for _, f := range fields {
        f = strings.TrimSpace(f)
        if f == "" { continue }
        parts := strings.SplitN(f, ":", 2)
        col := strings.ToLower(parts[0])
        typ := "text"
        if len(parts) == 2 {
            switch strings.ToLower(strings.TrimSpace(parts[1])) {
            case "string", "text": typ = "text"
            case "int", "integer": typ = "integer"
            case "bigint": typ = "bigint"
            case "bool", "boolean": typ = "boolean"
            case "float", "double", "doubleprecision": typ = "double precision"
            case "date": typ = "date"
            case "timestamp", "timestamptz": typ = "timestamptz"
            default: typ = "text"
            }
        }
        cols = append(cols, fmt.Sprintf("  %s %s NOT NULL", col, typ))
    }
    // Core columns
    prelude := []string{
        "  id bigserial PRIMARY KEY",
        "  created_at timestamptz DEFAULT now()",
        "  updated_at timestamptz DEFAULT now()",
    }
    body := strings.Join(append(prelude, cols...), ",\n")
    up := fmt.Sprintf("CREATE TABLE %s (\n%s\n);\n", table, body)
    down := fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", table)
    sql := fmt.Sprintf("-- +goose Up\n%s\n-- +goose Down\n%s", up, down)

    // Write file under app/db/migrations
    dir := filepath.Join("app", "db", "migrations")
    if err := os.MkdirAll(dir, 0o755); err != nil { return err }
    ts := time.Now().UTC().Format("20060102150405")
    file := filepath.Join(dir, fmt.Sprintf("%s_create_%s.sql", ts, table))
    if err := os.WriteFile(file, []byte(sql), 0o644); err != nil { return err }
    fmt.Printf("Added resource: /%s (page) + migration %s\n", keb, filepath.Base(file))
    fmt.Printf("  - %s\n", file)
    return nil
}

// scaffoldDB creates a starter SQL schema file under app/db.
func scaffoldDB(name string) error {
    keb := kebabCase(name)
    sqlPath := filepath.Join("app", "db", fmt.Sprintf("%s.sql", keb))
    sqlSrc := fmt.Sprintf(`-- SQL starter for %s
-- Edit this file and manage migrations with your preferred tool.

-- example table
-- create table %s (
--   id serial primary key,
--   created_at timestamptz default now(),
--   name text not null
-- );
`, keb, keb)
    if err := execx.WriteFileIfMissing(sqlPath, []byte(sqlSrc), 0o644); err != nil { return err }
    fmt.Printf("Added DB schema starter: %s\n", sqlPath)
    return nil
}

// scaffoldAuth creates minimal session-backed login/logout routes and a login page.
func scaffoldAuth() error {
    // Routes
    routePath := filepath.Join("app", "routes", "auth.go")
    routeSrc := `package routes

import (
    "net/http"
    "strings"
    "github.com/go-chi/chi/v5"
    "gothicforge3/app/templates"
    "gothicforge3/internal/server"
)

func init() {
    RegisterRoute(func(r chi.Router) {
        r.Get("/login", func(w http.ResponseWriter, req *http.Request) {
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            _ = templates.AuthLogin().Render(req.Context(), w)
        })
        r.Post("/login", func(w http.ResponseWriter, req *http.Request) {
            _ = req.ParseForm()
            user := strings.TrimSpace(req.FormValue("username"))
            if user == "" {
                http.Redirect(w, req, "/login?err=1", http.StatusSeeOther)
                return
            }
            server.Sessions().Put(req.Context(), "user", user)
            http.Redirect(w, req, "/", http.StatusSeeOther)
        })
        r.Get("/logout", func(w http.ResponseWriter, req *http.Request) {
            server.Sessions().Remove(req.Context(), "user")
            http.Redirect(w, req, "/", http.StatusSeeOther)
        })
        RegisterURL("/login")
    })
}
`
    if err := execx.WriteFileIfMissing(routePath, []byte(routeSrc), 0o644); err != nil { return err }

    // Template
    tmplPath := filepath.Join("app", "templates", "auth_login.go")
    tmplSrc := `package templates

import (
    "context"
    "io"
    templ "github.com/a-h/templ"
)

func AuthLogin() templ.Component {
    body := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        _, _ = io.WriteString(w, "<section class=\"mx-auto max-w-xl p-4\"><div class=\"card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10\"><div class=\"card-body\">")
        _, _ = io.WriteString(w, "<h2 class=\"card-title\">Sign in</h2>")
        _, _ = io.WriteString(w, "<form method=\"post\" action=\"/login\" class=\"grid gap-3\">")
        _, _ = io.WriteString(w, "<label class=\"form-control\"><span class=\"label-text\">Username</span><input type=\"text\" name=\"username\" class=\"input input-bordered\" required></label>")
        _, _ = io.WriteString(w, "<button class=\"btn btn-primary\" type=\"submit\">Continue</button>")
        _, _ = io.WriteString(w, "</form>")
        _, _ = io.WriteString(w, "</div></div></section>")
        return nil
    })
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error { return LayoutSEO(SEO{Title: "Sign in", Description: "Demo auth form", Canonical: "/login"}).Render(templ.WithChildren(ctx, body), w) })
}
`
    if err := execx.WriteFileIfMissing(tmplPath, []byte(tmplSrc), 0o644); err != nil { return err }

    fmt.Println("Added auth routes: /login, /logout")
    fmt.Printf("  - %s\n", routePath)
    fmt.Printf("  - %s\n", tmplPath)
    return nil
}

func init() { rootCmd.AddCommand(addCmd) }

func isValidName(s string) bool {
    re := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
    return re.MatchString(s)
}

func pascalCase(s string) string {
    parts := regexp.MustCompile(`[-_\s]+`).Split(s, -1)
    for i, p := range parts {
        if p == "" { continue }
        parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
    }
    return strings.Join(parts, "")
}

func kebabCase(s string) string {
    s = strings.TrimSpace(s)
    s = strings.ReplaceAll(s, "_", "-")
    s = strings.ToLower(s)
    return s
}

func scaffoldPage(name string) error {
    keb := kebabCase(name)
    pas := pascalCase(name)
    // 1) Template component (pure Go, no templ codegen required)
    tmplPath := filepath.Join("app", "templates", fmt.Sprintf("page_%s.go", keb))
    tmplSrc := fmt.Sprintf(`package templates

import (
    "context"
    "io"
    templ "github.com/a-h/templ"
)

func Page%[1]s() templ.Component {
    body := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        _, _ = io.WriteString(w, "<section class=\"mx-auto max-w-6xl p-4\">")
        _, _ = io.WriteString(w, "<div class=\"card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10\">")
        _, _ = io.WriteString(w, "<div class=\"card-body\">")
        _, _ = io.WriteString(w, "<h2 class=\"card-title\">%[1]s</h2>")
        _, _ = io.WriteString(w, "<p class=\"opacity-80\">Scaffolded page. Edit at app/templates/page_%[2]s.go</p>")
        _, _ = io.WriteString(w, "</div></div></section>")
        return nil
    })
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error { return LayoutSEO(SEO{Title: "%[1]s", Description: "%[1]s page", Canonical: "/%[2]s"}).Render(templ.WithChildren(ctx, body), w) })
}
`, pas, keb)
    if err := execx.WriteFileIfMissing(tmplPath, []byte(tmplSrc), 0o644); err != nil { return err }

    // 2) Route registrar that mounts GET /<keb>
    routePath := filepath.Join("app", "routes", fmt.Sprintf("page_%s.go", keb))
    routeSrc := fmt.Sprintf(`package routes

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "gothicforge3/app/templates"
)

func init() {
    RegisterRoute(func(r chi.Router) {
        r.Get("/%[1]s", func(w http.ResponseWriter, req *http.Request) {
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            _ = templates.Page%[2]s().Render(req.Context(), w)
        })
        RegisterURL("/%[1]s")
    })
}
`, keb, pas)
    if err := execx.WriteFileIfMissing(routePath, []byte(routeSrc), 0o644); err != nil { return err }

    fmt.Printf("Added page: /%s\n", keb)
    fmt.Printf("  - %s\n", tmplPath)
    fmt.Printf("  - %s\n", routePath)
    return nil
}

func scaffoldComponent(name string) error {
    keb := kebabCase(name)
    pas := pascalCase(name)
    compPath := filepath.Join("app", "templates", fmt.Sprintf("component_%s.go", keb))
    compSrc := fmt.Sprintf(`package templates

import (
    "context"
    "io"
    templ "github.com/a-h/templ"
)

func Component%[1]s() templ.Component {
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        _, _ = io.WriteString(w, "<div class=\"alert alert-info\">Component %[1]s</div>")
        return nil
    })
}
`, pas)
    if err := execx.WriteFileIfMissing(compPath, []byte(compSrc), 0o644); err != nil { return err }
    fmt.Printf("Added component: %s\n", compPath)
    return nil
}

// scaffoldModule bundles a page and db schema under the same name.
func scaffoldModule(name string) error {
    if err := scaffoldPage(name); err != nil { return err }
    if err := scaffoldDB(name); err != nil { return err }
    fmt.Printf("Added module: %s (page + db)\n", name)
    return nil
}

// scaffoldCRUD creates a memory-backed CRUD feature with JWT-protected mutating actions.
func scaffoldCRUD(name string) error {
    keb := kebabCase(name)
    pas := pascalCase(name)

    // 1) Templates
    tmplPath := filepath.Join("app", "templates", fmt.Sprintf("crud_%s.go", keb))
    tmplSrc := fmt.Sprintf(`package templates

import (
    "context"
    "io"
    templ "github.com/a-h/templ"
)

type %[1]sItem struct {
    ID int
    Name string
    Description string
    CreatedAt string
}

func Crud%[1]sList(items []%[1]sItem) templ.Component {
    body := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        _, _ = io.WriteString(w, "<section class=\"mx-auto max-w-6xl p-4\">")
        _, _ = io.WriteString(w, "<div class=\"flex justify-between items-center mb-4\"><h2 class=\"text-2xl font-bold\">%[1]s</h2><a class=\"btn btn-primary\" href=\"/%[2]s/new\">New</a></div>")
        _, _ = io.WriteString(w, "<div class=\"card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10\"><div class=\"card-body\">")
        _, _ = io.WriteString(w, "<ul class=\"menu\">")
        for _, it := range items {
            _, _ = io.WriteString(w, "<li><a href=\"/%[2]s/" + it.Name + "\">" + it.Name + "</a></li>")
        }
        _, _ = io.WriteString(w, "</ul></div></div></section>")
        return nil
    })
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error { return LayoutSEO(SEO{Title: "%[1]s", Description: "%[1]s list", Canonical: "/%[2]s"}).Render(templ.WithChildren(ctx, body), w) })
}

func Crud%[1]sForm(action string, item *%[1]sItem, submit string) templ.Component {
    body := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        _, _ = io.WriteString(w, "<section class=\"mx-auto max-w-xl p-4\"><div class=\"card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10\"><div class=\"card-body\">")
        _, _ = io.WriteString(w, "<h2 class=\"card-title\">%[1]s</h2>")
        _, _ = io.WriteString(w, "<form method=\"post\" action=\"" + action + "\" class=\"grid gap-3\">")
        name := ""
        desc := ""
        if item != nil { name = item.Name; desc = item.Description }
        _, _ = io.WriteString(w, "<label class=\"form-control\"><span class=\"label-text\">Name</span><input class=\"input input-bordered\" name=\"name\" value=\"" + name + "\" required></label>")
        _, _ = io.WriteString(w, "<label class=\"form-control\"><span class=\"label-text\">Description</span><textarea class=\"textarea textarea-bordered\" name=\"description\">" + desc + "</textarea></label>")
        _, _ = io.WriteString(w, "<button class=\"btn btn-primary\" type=\"submit\">" + submit + "</button>")
        _, _ = io.WriteString(w, "</form></div></div></section>")
        return nil
    })
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error { return LayoutSEO(SEO{Title: "%[1]s", Description: "%[1]s form", Canonical: "/%[2]s/new"}).Render(templ.WithChildren(ctx, body), w) })
}
`, pas, keb)
    if err := execx.WriteFileIfMissing(tmplPath, []byte(tmplSrc), 0o644); err != nil { return err }

    // 2) Routes (memory-backed, JWT-protected mutations)
    routePath := filepath.Join("app", "routes", fmt.Sprintf("crud_%s.go", keb))
    routeSrc := fmt.Sprintf(`package routes

import (
    "net/http"
    "sort"
    "strconv"
    "sync"
    "time"
    "github.com/go-chi/chi/v5"
    "gothicforge3/app/templates"
    "gothicforge3/internal/auth"
)

var (%[2]sMu sync.RWMutex; %[2]sStore = map[int]templates.%[1]sItem{}; %[2]sID int)

func init() {
    RegisterRoute(func(r chi.Router) {
        r.Get("/%[3]s", func(w http.ResponseWriter, req *http.Request) {
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            %[2]sMu.RLock();
            items := make([]templates.%[1]sItem, 0, len(%[2]sStore))
            for _, it := range %[2]sStore { items = append(items, it) }
            %[2]sMu.RUnlock()
            // stable order by Name
            sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
            _ = templates.Crud%[1]sList(items).Render(req.Context(), w)
        })
        r.Get("/%[3]s/new", func(w http.ResponseWriter, req *http.Request) {
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            _ = templates.Crud%[1]sForm("/%[3]s", nil, "Create").Render(req.Context(), w)
        })
        r.Post("/%[3]s", func(w http.ResponseWriter, req *http.Request) {
            if !requireJWT(req) { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
            _ = req.ParseForm()
            name := req.FormValue("name"); desc := req.FormValue("description")
            if name == "" { http.Redirect(w, req, "/%[3]s/new", http.StatusSeeOther); return }
            %[2]sMu.Lock(); %[2]sID++; id := %[2]sID
            %[2]sStore[id] = templates.%[1]sItem{ID: id, Name: name, Description: desc, CreatedAt: time.Now().Format(time.RFC3339)}
            %[2]sMu.Unlock()
            http.Redirect(w, req, "/%[3]s", http.StatusSeeOther)
        })
        r.Get("/%[3]s/{id}/edit", func(w http.ResponseWriter, req *http.Request) {
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            id, _ := strconv.Atoi(chi.URLParam(req, "id"))
            %[2]sMu.RLock(); it, ok := %[2]sStore[id]; %[2]sMu.RUnlock(); if !ok { http.NotFound(w, req); return }
            _ = templates.Crud%[1]sForm("/%[3]s/"+strconv.Itoa(id), &it, "Update").Render(req.Context(), w)
        })
        r.Post("/%[3]s/{id}", func(w http.ResponseWriter, req *http.Request) {
            if !requireJWT(req) { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
            _ = req.ParseForm(); id, _ := strconv.Atoi(chi.URLParam(req, "id"))
            name := req.FormValue("name"); desc := req.FormValue("description")
            %[2]sMu.Lock(); if it, ok := %[2]sStore[id]; ok { it.Name = name; it.Description = desc; %[2]sStore[id] = it }; %[2]sMu.Unlock()
            http.Redirect(w, req, "/%[3]s", http.StatusSeeOther)
        })
        r.Post("/%[3]s/{id}/delete", func(w http.ResponseWriter, req *http.Request) {
            if !requireJWT(req) { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
            id, _ := strconv.Atoi(chi.URLParam(req, "id"))
            %[2]sMu.Lock(); delete(%[2]sStore, id); %[2]sMu.Unlock()
            http.Redirect(w, req, "/%[3]s", http.StatusSeeOther)
        })
        RegisterURL("/%[3]s")
    })
}

func requireJWT(r *http.Request) bool { _, err := auth.ReadAndVerifyCookie(r, "gf_jwt"); return err == nil }
`, pas, keb, keb)
    if err := execx.WriteFileIfMissing(routePath, []byte(routeSrc), 0o644); err != nil { return err }

    fmt.Printf("Added CRUD: /%s (memory-backed; POST/PUT/DELETE require JWT)\n", keb)
    fmt.Printf("  - %s\n", tmplPath)
    fmt.Printf("  - %s\n", routePath)
    return nil
}
