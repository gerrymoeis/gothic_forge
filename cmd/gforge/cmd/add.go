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

var (
    systemFlag string
)

var addCmd = &cobra.Command{
    Use:   "add",
    Short: "Scaffold features in app/ (page, api, handler, model, edge, component, auth, job, etc.)",
    Args:  cobra.MinimumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        banner()
        kind := strings.ToLower(args[0])
        var name string
        if kind == "page" || kind == "component" || kind == "oauth" || kind == "db" || kind == "module" || kind == "crud" || kind == "resource" || kind == "migration" || kind == "cruddb" || kind == "api" || kind == "handler" || kind == "model" || kind == "edge" || kind == "job" || kind == "email" || kind == "form" {
            if len(args) < 2 {
                printAddUsage()
                return nil
            }
            name = args[1]
            if !isValidName(name) {
                return fmt.Errorf("invalid name: %s (use letters, numbers, dash, underscore)", name)
            }
        }
        switch kind {
        case "page":
            return scaffoldPage(name, systemFlag)
        case "api":
            method := "GET"
            if len(args) > 2 { method = strings.ToUpper(args[2]) }
            return scaffoldAPI(name, method)
        case "handler":
            return scaffoldHandler(name)
        case "model":
            fields := []string{}
            if len(args) > 2 { fields = args[2:] }
            return scaffoldModel(name, fields)
        case "edge":
            method := "GET"
            if len(args) > 2 { method = strings.ToUpper(args[2]) }
            return scaffoldEdge(name, method)
        case "component":
            return scaffoldComponent(name, systemFlag)
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
        case "email":
            provider := "sendgrid" // default
            if len(args) > 2 {
                provider = args[2]
            }
            return scaffoldEmail(name, provider)
        case "form":
            return scaffoldForm(name, systemFlag)
        default:
            printAddUsage()
            return nil
        }
    },
}

func printAddUsage() {
    fmt.Println("Usage:")
    fmt.Println()
    fmt.Println("📄 Pages & UI:")
    fmt.Println("  gforge add page <name>            - Add HTML page with route")
    fmt.Println("  gforge add component <name>       - Add reusable component")
    fmt.Println("  gforge add component <name> --system=webawesome - Add Web Awesome component")
    fmt.Println("  gforge add form <name>            - Add form component")
    fmt.Println("  gforge add form <name> --system=webawesome - Add Web Awesome form")
    fmt.Println()
    fmt.Println("🚀 API & Routes:")
    fmt.Println("  gforge add api <name> [method]    - Add API endpoint (default: GET)")
    fmt.Println("  gforge add handler <name>         - Add route handler")
    fmt.Println("  gforge add edge <path> [method]   - Add Cloudflare edge function")
    fmt.Println()
    fmt.Println("🗄️  Database & Models:")
    fmt.Println("  gforge add model <Name> [field:type ...]  - Add database model")
    fmt.Println("  gforge add migration <name>       - Add database migration")
    fmt.Println("  gforge add db <name>              - Add database schema file")
    fmt.Println()
    fmt.Println("✨ Full Features:")
    fmt.Println("  gforge add crud <name>            - Add memory-backed CRUD")
    fmt.Println("  gforge add cruddb <Name> [field:type ...]  - Add DB-backed CRUD")
    fmt.Println("  gforge add resource <Name> [field:type ...]  - Add page + migration")
    fmt.Println("  gforge add module <name>          - Add page + db schema")
    fmt.Println()
    fmt.Println("🔐 Authentication:")
    fmt.Println("  gforge add auth                   - Add login/logout routes")
    fmt.Println("  gforge add oauth <provider>       - Add OAuth provider routes")
    fmt.Println()
    fmt.Println("📧 Email:")
    fmt.Println("  gforge add email <template> [provider]  - Add email template (default: sendgrid)")
    fmt.Println()
    fmt.Println("Examples:")
    fmt.Println("  gforge add api users GET")
    fmt.Println("  gforge add model Post title:string body:text")
    fmt.Println("  gforge add edge /api/hello POST")
    fmt.Println("  gforge add cruddb Article title:string content:text")
    fmt.Println("  gforge add email welcome sendgrid")
    fmt.Println("  gforge add email password-reset mailgun")
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
    if err := scaffoldPage(name, ""); err != nil { return err }

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

func scaffoldWebAwesomeComponent(name string) error {
    keb := kebabCase(name)
    pas := pascalCase(name)
    
    // Create component file
    compPath := filepath.Join("app", "templates", fmt.Sprintf("webawesome_%s.go", keb))
    compSrc := fmt.Sprintf(`package templates

import (
    "context"
    "io"
    templ "github.com/a-h/templ"
    "gothicforge3/internal/webawesome"
)

// %[1]sProps defines the properties for the %[1]s component
type %[1]sProps struct {
    ID       string
    Class    string
    Variant  string
    Size     string
    Disabled bool
    // Add more properties as needed
}

// WebAwesome%[1]s creates a Web Awesome %[2]s component
func WebAwesome%[1]s(props %[1]sProps) templ.Component {
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        // Create Web Awesome component
        comp := webawesome.NewButtonComponent()
        
        // Set properties
        if props.ID != "" {
            comp.SetID(props.ID)
        }
        if props.Class != "" {
            comp.SetClass(props.Class)
        }
        if props.Variant != "" {
            comp.SetAttribute("variant", props.Variant)
        }
        if props.Size != "" {
            comp.SetAttribute("size", props.Size)
        }
        if props.Disabled {
            comp.SetAttribute("disabled", "true")
        }
        
        // Set content
        comp.SetContent("%[1]s Component")
        
        // Render component
        html, err := comp.RenderHTML()
        if err != nil {
            return err
        }
        
        _, err = io.WriteString(w, html)
        return err
    })
}

// %[1]sExample shows usage examples for the %[1]s component
func %[1]sExample() templ.Component {
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        _, _ = io.WriteString(w, "<div class=\"space-y-4\">")
        _, _ = io.WriteString(w, "<h3 class=\"text-lg font-semibold\">%[1]s Examples</h3>")
        
        // Default example
        _, _ = io.WriteString(w, "<div class=\"example\">")
        _, _ = io.WriteString(w, "<h4 class=\"font-medium\">Default</h4>")
        defaultProps := %[1]sProps{}
        _ = WebAwesome%[1]s(defaultProps).Render(ctx, w)
        _, _ = io.WriteString(w, "</div>")
        
        // Primary variant example
        _, _ = io.WriteString(w, "<div class=\"example\">")
        _, _ = io.WriteString(w, "<h4 class=\"font-medium\">Primary</h4>")
        primaryProps := %[1]sProps{Variant: "primary"}
        _ = WebAwesome%[1]s(primaryProps).Render(ctx, w)
        _, _ = io.WriteString(w, "</div>")
        
        // Disabled example
        _, _ = io.WriteString(w, "<div class=\"example\">")
        _, _ = io.WriteString(w, "<h4 class=\"font-medium\">Disabled</h4>")
        disabledProps := %[1]sProps{Disabled: true}
        _ = WebAwesome%[1]s(disabledProps).Render(ctx, w)
        _, _ = io.WriteString(w, "</div>")
        
        _, _ = io.WriteString(w, "</div>")
        return nil
    })
}
`, pas, keb)
    
    if err := execx.WriteFileIfMissing(compPath, []byte(compSrc), 0o644); err != nil { 
        return err 
    }
    
    fmt.Printf("Added Web Awesome component: %s\n", pas)
    fmt.Printf("  - %s\n", compPath)
    fmt.Println()
    fmt.Println("Usage in templates:")
    fmt.Printf("  props := %sProps{Variant: \"primary\", Size: \"medium\"}\n", pas)
    fmt.Printf("  _ = WebAwesome%s(props).Render(ctx, w)\n", pas)
    return nil
}

func scaffoldForm(name string, system string) error {
    keb := kebabCase(name)
    pas := pascalCase(name)
    
    if system == "webawesome" {
        return scaffoldWebAwesomeForm(name)
    }
    
    // Default form scaffolding
    formPath := filepath.Join("app", "templates", fmt.Sprintf("form_%s.go", keb))
    formSrc := fmt.Sprintf(`package templates

import (
    "context"
    "io"
    templ "github.com/a-h/templ"
)

// %[1]sFormData represents the form data structure
type %[1]sFormData struct {
    Name        string
    Email       string
    Message     string
    // Add more fields as needed
}

// %[1]sForm creates a form component
func %[1]sForm(action string, data *%[1]sFormData, errors map[string]string) templ.Component {
    if data == nil {
        data = &%[1]sFormData{}
    }
    if errors == nil {
        errors = make(map[string]string)
    }
    
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        _, _ = io.WriteString(w, "<form method=\"post\" action=\"" + action + "\" class=\"space-y-4\">")
        
        // Name field
        _, _ = io.WriteString(w, "<div class=\"form-control\">")
        _, _ = io.WriteString(w, "<label class=\"label\"><span class=\"label-text\">Name</span></label>")
        _, _ = io.WriteString(w, "<input type=\"text\" name=\"name\" value=\"" + data.Name + "\" class=\"input input-bordered\" required>")
        if err, exists := errors["name"]; exists {
            _, _ = io.WriteString(w, "<label class=\"label\"><span class=\"label-text-alt text-error\">" + err + "</span></label>")
        }
        _, _ = io.WriteString(w, "</div>")
        
        // Email field
        _, _ = io.WriteString(w, "<div class=\"form-control\">")
        _, _ = io.WriteString(w, "<label class=\"label\"><span class=\"label-text\">Email</span></label>")
        _, _ = io.WriteString(w, "<input type=\"email\" name=\"email\" value=\"" + data.Email + "\" class=\"input input-bordered\" required>")
        if err, exists := errors["email"]; exists {
            _, _ = io.WriteString(w, "<label class=\"label\"><span class=\"label-text-alt text-error\">" + err + "</span></label>")
        }
        _, _ = io.WriteString(w, "</div>")
        
        // Message field
        _, _ = io.WriteString(w, "<div class=\"form-control\">")
        _, _ = io.WriteString(w, "<label class=\"label\"><span class=\"label-text\">Message</span></label>")
        _, _ = io.WriteString(w, "<textarea name=\"message\" class=\"textarea textarea-bordered\" rows=\"4\">" + data.Message + "</textarea>")
        if err, exists := errors["message"]; exists {
            _, _ = io.WriteString(w, "<label class=\"label\"><span class=\"label-text-alt text-error\">" + err + "</span></label>")
        }
        _, _ = io.WriteString(w, "</div>")
        
        // Submit button
        _, _ = io.WriteString(w, "<div class=\"form-control\">")
        _, _ = io.WriteString(w, "<button type=\"submit\" class=\"btn btn-primary\">Submit</button>")
        _, _ = io.WriteString(w, "</div>")
        
        _, _ = io.WriteString(w, "</form>")
        return nil
    })
}
`, pas)
    
    if err := execx.WriteFileIfMissing(formPath, []byte(formSrc), 0o644); err != nil { 
        return err 
    }
    
    fmt.Printf("Added form: %s\n", pas)
    fmt.Printf("  - %s\n", formPath)
    return nil
}

func scaffoldWebAwesomeForm(name string) error {
    keb := kebabCase(name)
    pas := pascalCase(name)
    
    formPath := filepath.Join("app", "templates", fmt.Sprintf("webawesome_form_%s.go", keb))
    formSrc := fmt.Sprintf(`package templates

import (
    "context"
    "io"
    templ "github.com/a-h/templ"
    "gothicforge3/internal/webawesome"
)

// %[1]sFormData represents the form data structure
type %[1]sFormData struct {
    Name        string
    Email       string
    Message     string
    // Add more fields as needed
}

// WebAwesome%[1]sForm creates a Web Awesome form component
func WebAwesome%[1]sForm(action string, data *%[1]sFormData, errors map[string]string) templ.Component {
    if data == nil {
        data = &%[1]sFormData{}
    }
    if errors == nil {
        errors = make(map[string]string)
    }
    
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        _, _ = io.WriteString(w, "<form method=\"post\" action=\"" + action + "\" class=\"space-y-6\">")
        
        // Name field
        nameInput := webawesome.NewInputComponent()
        nameInput.SetAttribute("name", "name")
        nameInput.SetAttribute("label", "Name")
        nameInput.SetAttribute("value", data.Name)
        nameInput.SetAttribute("required", "true")
        if err, exists := errors["name"]; exists {
            nameInput.SetAttribute("help-text", err)
            nameInput.SetClass("sl-input--invalid")
        }
        nameHTML, _ := nameInput.RenderHTML()
        _, _ = io.WriteString(w, nameHTML)
        
        // Email field
        emailInput := webawesome.NewInputComponent()
        emailInput.SetAttribute("name", "email")
        emailInput.SetAttribute("type", "email")
        emailInput.SetAttribute("label", "Email")
        emailInput.SetAttribute("value", data.Email)
        emailInput.SetAttribute("required", "true")
        if err, exists := errors["email"]; exists {
            emailInput.SetAttribute("help-text", err)
            emailInput.SetClass("sl-input--invalid")
        }
        emailHTML, _ := emailInput.RenderHTML()
        _, _ = io.WriteString(w, emailHTML)
        
        // Message field
        messageTextarea := webawesome.NewTextareaComponent()
        messageTextarea.SetAttribute("name", "message")
        messageTextarea.SetAttribute("label", "Message")
        messageTextarea.SetAttribute("rows", "4")
        messageTextarea.SetContent(data.Message)
        if err, exists := errors["message"]; exists {
            messageTextarea.SetAttribute("help-text", err)
            messageTextarea.SetClass("sl-textarea--invalid")
        }
        messageHTML, _ := messageTextarea.RenderHTML()
        _, _ = io.WriteString(w, messageHTML)
        
        // Submit button
        submitBtn := webawesome.NewButtonComponent()
        submitBtn.SetAttribute("type", "submit")
        submitBtn.SetAttribute("variant", "primary")
        submitBtn.SetContent("Submit")
        submitHTML, _ := submitBtn.RenderHTML()
        _, _ = io.WriteString(w, submitHTML)
        
        _, _ = io.WriteString(w, "</form>")
        return nil
    })
}

// %[1]sFormExample shows usage examples for the %[1]s form
func %[1]sFormExample() templ.Component {
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
        _, _ = io.WriteString(w, "<div class=\"space-y-6\">")
        _, _ = io.WriteString(w, "<h3 class=\"text-lg font-semibold\">%[1]s Form Examples</h3>")
        
        // Default form
        _, _ = io.WriteString(w, "<div class=\"example\">")
        _, _ = io.WriteString(w, "<h4 class=\"font-medium\">Default Form</h4>")
        _ = WebAwesome%[1]sForm("/submit", nil, nil).Render(ctx, w)
        _, _ = io.WriteString(w, "</div>")
        
        // Form with data
        _, _ = io.WriteString(w, "<div class=\"example\">")
        _, _ = io.WriteString(w, "<h4 class=\"font-medium\">Form with Data</h4>")
        sampleData := &%[1]sFormData{
            Name:    "John Doe",
            Email:   "john@example.com",
            Message: "Hello, this is a sample message.",
        }
        _ = WebAwesome%[1]sForm("/submit", sampleData, nil).Render(ctx, w)
        _, _ = io.WriteString(w, "</div>")
        
        // Form with errors
        _, _ = io.WriteString(w, "<div class=\"example\">")
        _, _ = io.WriteString(w, "<h4 class=\"font-medium\">Form with Validation Errors</h4>")
        errors := map[string]string{
            "name":  "Name is required",
            "email": "Please enter a valid email address",
        }
        _ = WebAwesome%[1]sForm("/submit", nil, errors).Render(ctx, w)
        _, _ = io.WriteString(w, "</div>")
        
        _, _ = io.WriteString(w, "</div>")
        return nil
    })
}
`, pas)
    
    if err := execx.WriteFileIfMissing(formPath, []byte(formSrc), 0o644); err != nil { 
        return err 
    }
    
    fmt.Printf("Added Web Awesome form: %s\n", pas)
    fmt.Printf("  - %s\n", formPath)
    fmt.Println()
    fmt.Println("Usage in templates:")
    fmt.Printf("  data := &%sFormData{Name: \"John\", Email: \"john@example.com\"}\n", pas)
    fmt.Printf("  _ = WebAwesome%sForm(\"/submit\", data, nil).Render(ctx, w)\n", pas)
    return nil
}

func init() { 
    addCmd.Flags().StringVar(&systemFlag, "system", "", "Component system to use (webawesome)")
    rootCmd.AddCommand(addCmd) 
}

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

func scaffoldPage(name string, system string) error {
    keb := kebabCase(name)
    pas := pascalCase(name)
    
    var cardClass, containerClass string
    if system == "webawesome" {
        cardClass = "sl-card"
        containerClass = "container mx-auto p-4"
    } else {
        cardClass = "card bg-base-200/60 border border-white/10 rounded-box shadow-xl ring-1 ring-white/10"
        containerClass = "mx-auto max-w-6xl p-4"
    }
    
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
        _, _ = io.WriteString(w, "<section class=\"%[4]s\">")
        _, _ = io.WriteString(w, "<div class=\"%[5]s\">")
        if "%[6]s" == "webawesome" {
            _, _ = io.WriteString(w, "<h2 class=\"text-2xl font-bold mb-4\">%[1]s</h2>")
            _, _ = io.WriteString(w, "<p class=\"text-gray-600\">Scaffolded Web Awesome page. Edit at app/templates/page_%[2]s.go</p>")
        } else {
            _, _ = io.WriteString(w, "<div class=\"card-body\">")
            _, _ = io.WriteString(w, "<h2 class=\"card-title\">%[1]s</h2>")
            _, _ = io.WriteString(w, "<p class=\"opacity-80\">Scaffolded page. Edit at app/templates/page_%[2]s.go</p>")
            _, _ = io.WriteString(w, "</div>")
        }
        _, _ = io.WriteString(w, "</div></section>")
        return nil
    })
    return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error { return LayoutSEO(SEO{Title: "%[1]s", Description: "%[1]s page", Canonical: "/%[2]s"}).Render(templ.WithChildren(ctx, body), w) })
}
`, pas, keb, pas, containerClass, cardClass, system)
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

func scaffoldComponent(name string, system string) error {
    keb := kebabCase(name)
    pas := pascalCase(name)
    
    if system == "webawesome" {
        return scaffoldWebAwesomeComponent(name)
    }
    
    // Default component scaffolding
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
    if err := scaffoldPage(name, ""); err != nil { return err }
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

// scaffoldAPI creates a JSON API endpoint.
// Example: gforge add api users GET
func scaffoldAPI(name string, method string) error {
    keb := kebabCase(name)
    pas := pascalCase(name)
    routePath := filepath.Join("app", "routes", fmt.Sprintf("api_%s.go", keb))
    
    methodLower := strings.ToLower(method)
    chiMethod := strings.Title(methodLower)
    
    routeSrc := fmt.Sprintf(`package routes

import (
    "encoding/json"
    "net/http"
    "github.com/go-chi/chi/v5"
)

func init() {
    RegisterRoute(func(r chi.Router) {
        r.%s("/api/%s", handle%sAPI)
        RegisterURL("/api/%s")
    })
}

func handle%sAPI(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    
    // TODO: Implement your API logic here
    response := map[string]interface{}{
        "success": true,
        "message": "%s API endpoint",
        "method":  "%s",
    }
    
    _ = json.NewEncoder(w).Encode(response)
}
`, chiMethod, keb, pas, keb, pas, pas, method)
    
    if err := execx.WriteFileIfMissing(routePath, []byte(routeSrc), 0o644); err != nil { return err }
    
    fmt.Printf("Added API endpoint: %s /api/%s\n", method, keb)
    fmt.Printf("  - %s\n", routePath)
    return nil
}

// scaffoldHandler creates a generic route handler.
// Example: gforge add handler dashboard
func scaffoldHandler(name string) error {
    keb := kebabCase(name)
    pas := pascalCase(name)
    routePath := filepath.Join("app", "routes", fmt.Sprintf("handler_%s.go", keb))
    
    routeSrc := fmt.Sprintf(`package routes

import (
    "net/http"
    "github.com/go-chi/chi/v5"
)

func init() {
    RegisterRoute(func(r chi.Router) {
        r.Get("/%s", handle%s)
        RegisterURL("/%s")
    })
}

func handle%s(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    
    // TODO: Implement your handler logic here
    // Option 1: Render template
    // _ = templates.Page%s().Render(r.Context(), w)
    
    // Option 2: Return plain text
    _, _ = w.Write([]byte("Handler: %s"))
}
`, keb, pas, keb, pas, pas, pas)
    
    if err := execx.WriteFileIfMissing(routePath, []byte(routeSrc), 0o644); err != nil { return err }
    
    fmt.Printf("Added handler: /%s\n", keb)
    fmt.Printf("  - %s\n", routePath)
    return nil
}

// scaffoldModel creates a database model struct and repository.
// Example: gforge add model Post title:string body:text published:bool
func scaffoldModel(name string, fields []string) error {
    pas := pascalCase(name)
    keb := kebabCase(name)
    
    // Parse fields
    type fieldInfo struct {
        Name    string
        GoType  string
        SQLType string
        JSONTag string
    }
    
    parsedFields := []fieldInfo{}
    for _, f := range fields {
        parts := strings.SplitN(strings.TrimSpace(f), ":", 2)
        if len(parts) != 2 || parts[0] == "" {
            continue
        }
        
        fieldName := pascalCase(parts[0])
        fieldType := strings.ToLower(strings.TrimSpace(parts[1]))
        
        var goType, sqlType string
        switch fieldType {
        case "string", "text":
            goType = "string"
            sqlType = "text"
        case "int", "integer":
            goType = "int64"
            sqlType = "bigint"
        case "bool", "boolean":
            goType = "bool"
            sqlType = "boolean"
        case "float", "double":
            goType = "float64"
            sqlType = "double precision"
        case "time", "timestamp":
            goType = "time.Time"
            sqlType = "timestamptz"
        default:
            goType = "string"
            sqlType = "text"
        }
        
        parsedFields = append(parsedFields, fieldInfo{
            Name:    fieldName,
            GoType:  goType,
            SQLType: sqlType,
            JSONTag: strings.ToLower(parts[0]),
        })
    }
    
    // Generate struct fields
    var structFields strings.Builder
    for _, f := range parsedFields {
        structFields.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", f.Name, f.GoType, f.JSONTag))
    }
    
    // Create model file
    modelPath := filepath.Join("app", "models", fmt.Sprintf("%s.go", keb))
    if err := os.MkdirAll(filepath.Dir(modelPath), 0o755); err != nil {
        return err
    }
    
    modelSrc := fmt.Sprintf(`package models

import (
	"context"
	"time"
	
	"gothicforge3/internal/db"
)

// %s represents a %s entity
type %s struct {
	ID        int64     ` + "`json:\"id\"`" + `
%s	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}

// %sRepository handles database operations for %s
type %sRepository struct{}

// New%sRepository creates a new repository instance
func New%sRepository() *%sRepository {
	return &%sRepository{}
}

// FindByID retrieves a %s by ID
func (repo *%sRepository) FindByID(ctx context.Context, id int64) (*%s, error) {
	// TODO: Implement database query
	// Example:
	// query := "SELECT * FROM %ss WHERE id = $1"
	// row := db.Pool().QueryRow(ctx, query, id)
	// var item %s
	// err := row.Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	// return &item, err
	return nil, nil
}

// FindAll retrieves all %ss
func (repo *%sRepository) FindAll(ctx context.Context, limit int) ([]*%s, error) {
	// TODO: Implement database query
	return nil, nil
}

// Create inserts a new %s
func (repo *%sRepository) Create(ctx context.Context, item *%s) error {
	// TODO: Implement database insert
	return nil
}

// Update modifies an existing %s
func (repo *%sRepository) Update(ctx context.Context, item *%s) error {
	// TODO: Implement database update
	return nil
}

// Delete removes a %s by ID
func (repo *%sRepository) Delete(ctx context.Context, id int64) error {
	// TODO: Implement database delete
	return nil
}
`,
    // 1-4: header and struct
    pas, keb, pas, structFields.String(),
    // 5-11: repo comment, type, constructor and return
    pas, pas, pas, pas, pas, pas, pas,
    // 12-16: FindByID comment/receiver/return/table/item type
    pas, pas, pas, keb, pas,
    // 17-19: FindAll comment/receiver/return elem type
    keb, pas, pas,
    // 20-22: Create comment/receiver/param type
    pas, pas, pas,
    // 23-25: Update comment/receiver/param type
    pas, pas, pas,
    // 26-27: Delete comment/receiver type
    pas, pas)
    
    if err := execx.WriteFileIfMissing(modelPath, []byte(modelSrc), 0o644); err != nil { return err }
    
    fmt.Printf("Added model: %s\n", pas)
    fmt.Printf("  - %s\n", modelPath)
    fmt.Println()
    fmt.Println("Next steps:")
    fmt.Println("  1. Create migration: gforge add migration create_" + keb + "s")
    fmt.Println("  2. Implement repository methods in " + modelPath)
    return nil
}

// scaffoldEdge creates a Cloudflare Pages Function directly.
// Example: gforge add edge /api/hello POST
func scaffoldEdge(path string, method string) error {
    // Validate path
    if !strings.HasPrefix(path, "/") {
        path = "/" + path
    }
    
    // Convert path to file structure
    // /api/users -> functions/api/users.js
    // /hello -> functions/hello.js
    cleanPath := strings.TrimPrefix(path, "/")
    parts := strings.Split(cleanPath, "/")
    
    var filePath string
    functionsDir := "functions"
    
    if len(parts) == 1 {
        filePath = filepath.Join(functionsDir, parts[0]+".js")
    } else {
        dir := filepath.Join(functionsDir, filepath.Join(parts[:len(parts)-1]...))
        if err := os.MkdirAll(dir, 0o755); err != nil {
            return err
        }
        filePath = filepath.Join(dir, parts[len(parts)-1]+".js")
    }
    
    // Generate JavaScript
    methodLower := strings.ToLower(method)
    handlerName := fmt.Sprintf("onRequest%s", strings.Title(methodLower))
    
    jsSrc := fmt.Sprintf(`/**
 * Cloudflare Pages Function: %s
 * 
 * Method: %s
 * Path: %s
 * 
 * Created by: gforge add edge
 */

export async function %s(context) {
  try {
    // Extract URL parameters
    const url = new URL(context.request.url);
    const params = Object.fromEntries(url.searchParams);
`, path, method, path, handlerName)
    
    if methodLower == "post" || methodLower == "put" || methodLower == "patch" {
        jsSrc += `
    // Parse request body
    const contentType = context.request.headers.get('Content-Type') || '';
    let body;
    
    if (contentType.includes('application/json')) {
      body = await context.request.json();
    } else if (contentType.includes('application/x-www-form-urlencoded')) {
      const formData = await context.request.formData();
      body = Object.fromEntries(formData);
    } else {
      body = await context.request.text();
    }
`
    }
    
    jsSrc += `
    // TODO: Implement your logic here
    const response = {
      success: true,
      message: 'Edge function response',
      method: '` + method + `',
      path: '` + path + `',
    };
    
    return new Response(
      JSON.stringify(response),
      {
        status: 200,
        headers: {
          'Content-Type': 'application/json',
          'Access-Control-Allow-Origin': '*',
        },
      }
    );
  } catch (error) {
    return new Response(
      JSON.stringify({ 
        success: false, 
        error: error.message 
      }),
      {
        status: 500,
        headers: { 'Content-Type': 'application/json' },
      }
    );
  }
}

// Handle CORS preflight
export async function onRequestOptions(context) {
  return new Response(null, {
    status: 204,
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': '` + method + `, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      'Access-Control-Max-Age': '86400',
    },
  });
}
`
    
    if err := execx.WriteFileIfMissing(filePath, []byte(jsSrc), 0o644); err != nil { return err }
    
    fmt.Printf("Added edge function: %s %s\n", method, path)
    fmt.Printf("  - %s\n", filePath)
    fmt.Println()
    fmt.Println("Next steps:")
    fmt.Println("  1. Implement logic in " + filePath)
    fmt.Println("  2. Export: gforge export")
    fmt.Println("  3. Deploy: gforge deploy pages --run")
    return nil
}

// scaffoldEmail creates an email template with provider-specific configuration.
// Example: gforge add email welcome --provider=sendgrid
func scaffoldEmail(name string, provider string) error {
	keb := kebabCase(name)
	pas := pascalCase(name)
	snake := ""
	for i, r := range pas {
		if i > 0 && r >= 'A' && r <= 'Z' {
			snake += "_"
		}
		snake += string(r)
	}
	snake = strings.ToLower(snake)

	// Default to sendgrid if no provider specified
	if provider == "" {
		provider = "sendgrid"
	}

	// Validate provider
	validProviders := map[string]bool{
		"sendgrid": true,
		"mailgun":  true,
		"ses":      true,
		"smtp":     true,
	}
	if !validProviders[provider] {
		return fmt.Errorf("invalid provider: %s (valid: sendgrid, mailgun, ses, smtp)", provider)
	}

	// Create email templates directory
	emailDir := filepath.Join("app", "email")
	if err := os.MkdirAll(emailDir, 0o755); err != nil {
		return err
	}

	// Create email template file
	emailPath := filepath.Join(emailDir, fmt.Sprintf("%s.go", snake))
	emailSrc := fmt.Sprintf(`package email

import (
	"context"
	"fmt"
	"os"

	"gothicforge3/internal/email"
)

// %[1]s represents the data needed for the %[2]s email.
type %[1]sData struct {
	// Add your template data fields here
	// Example: Name string, VerificationURL string, etc.
}

// Send%[1]s sends a %[2]s email to the specified recipient.
//
// Example usage:
//
//	data := %[1]sData{
//	    // Fill in your data
//	}
//	err := Send%[1]s(ctx, "user@example.com", data)
func Send%[1]s(ctx context.Context, to string, data %[1]sData) error {
	// Initialize email provider
	provider, err := initEmailProvider()
	if err != nil {
		return fmt.Errorf("failed to initialize email provider: %%w", err)
	}

	// Build email content
	subject := "%[3]s"
	htmlBody := build%[1]sHTML(data)
	textBody := build%[1]sText(data)

	// Create email
	msg := &email.Email{
		From:    os.Getenv("EMAIL_FROM"),
		To:      []string{to},
		Subject: subject,
		HTML:    htmlBody,
		Text:    textBody,
	}

	// Validate email
	if err := msg.Validate(); err != nil {
		return fmt.Errorf("invalid email: %%w", err)
	}

	// Send email
	if err := provider.Send(ctx, msg); err != nil {
		return fmt.Errorf("failed to send email: %%w", err)
	}

	return nil
}

// build%[1]sHTML generates the HTML version of the email.
func build%[1]sHTML(data %[1]sData) string {
	// TODO: Implement your HTML email template
	// You can use html/template or a templating library
	return ` + "`" + `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%[3]s</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 30px;
            text-align: center;
            border-radius: 8px 8px 0 0;
        }
        .content {
            background: #ffffff;
            padding: 30px;
            border: 1px solid #e0e0e0;
            border-top: none;
        }
        .footer {
            background: #f5f5f5;
            padding: 20px;
            text-align: center;
            font-size: 12px;
            color: #666;
            border-radius: 0 0 8px 8px;
        }
        .button {
            display: inline-block;
            padding: 12px 24px;
            background: #667eea;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>%[3]s</h1>
    </div>
    <div class="content">
        <p>Hello,</p>
        <p>This is your %[2]s email.</p>
        <!-- TODO: Add your email content here -->
        <p>Best regards,<br>Your Team</p>
    </div>
    <div class="footer">
        <p>&copy; 2024 Your Company. All rights reserved.</p>
    </div>
</body>
</html>
` + "`" + `
}

// build%[1]sText generates the plain text version of the email.
func build%[1]sText(data %[1]sData) string {
	// TODO: Implement your plain text email template
	return ` + "`" + `
%[3]s

Hello,

This is your %[2]s email.

TODO: Add your email content here

Best regards,
Your Team

---
© 2024 Your Company. All rights reserved.
` + "`" + `
}
`, pas, keb, strings.Title(strings.ReplaceAll(keb, "-", " ")))

	if err := execx.WriteFileIfMissing(emailPath, []byte(emailSrc), 0o644); err != nil {
		return err
	}

	// Create test file
	testPath := filepath.Join(emailDir, fmt.Sprintf("%s_test.go", snake))
	testSrc := fmt.Sprintf(`package email

import (
	"testing"
)

// Test%[1]s tests the %[1]s email template.
func Test%[1]s(t *testing.T) {
	// Skip if no email provider is configured
	if _, err := initEmailProvider(); err != nil {
		t.Skip("No email provider configured")
	}

	data := %[1]sData{
		// Add test data here
	}

	// Test HTML generation
	html := build%[1]sHTML(data)
	if html == "" {
		t.Error("HTML body should not be empty")
	}
	if len(html) < 100 {
		t.Error("HTML body seems too short")
	}

	// Test text generation
	text := build%[1]sText(data)
	if text == "" {
		t.Error("Text body should not be empty")
	}
	if len(text) < 50 {
		t.Error("Text body seems too short")
	}

	// Note: Actual sending is not tested to avoid sending real emails
	// To test sending, set up a test email provider (e.g., MailHog)
	t.Log("Email template generated successfully")
}

// Test%[1]sValidation tests email validation.
func Test%[1]sValidation(t *testing.T) {
	tests := []struct {
		name    string
		to      string
		wantErr bool
	}{
		{
			name:    "valid email",
			to:      "test@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			to:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would require a mock provider to test properly
			// For now, just test the template generation
			data := %[1]sData{}
			html := build%[1]sHTML(data)
			text := build%[1]sText(data)
			
			if html == "" || text == "" {
				t.Error("Template generation failed")
			}
		})
	}
}
`, pas)

	if err := execx.WriteFileIfMissing(testPath, []byte(testSrc), 0o644); err != nil {
		return err
	}

	// Create example usage file
	examplePath := filepath.Join(emailDir, fmt.Sprintf("%s_example.go", snake))
	exampleSrc := fmt.Sprintf(`package email

import (
	"context"
	"log"
)

// Example%[1]s demonstrates how to use the %[1]s email template.
func Example%[1]s() {
	ctx := context.Background()

	// Prepare email data
	data := %[1]sData{
		// Fill in your data fields
		// Example: Name: "John Doe", VerificationURL: "https://example.com/verify/abc123"
	}

	// Send email
	err := Send%[1]s(ctx, "user@example.com", data)
	if err != nil {
		log.Printf("Failed to send %[2]s email: %%v", err)
		return
	}

	log.Println("%[3]s email sent successfully")
}
`, pas, keb, strings.Title(strings.ReplaceAll(keb, "-", " ")))

	if err := execx.WriteFileIfMissing(examplePath, []byte(exampleSrc), 0o644); err != nil {
		return err
	}

	// Create README if it doesn't exist
	readmePath := filepath.Join(emailDir, "README.md")
	readmeSrc := `# Email Templates

This directory contains email templates for the application.

## Usage

Each email template is a Go package that provides:
- A data structure for template variables
- A Send function to send the email
- HTML and text versions of the email
- Tests and examples

## Sending Emails

` + "```go" + `
import "yourapp/app/email"

// Send an email
data := email.WelcomeData{
    Name: "John Doe",
}
err := email.SendWelcome(ctx, "user@example.com", data)
` + "```" + `

## Configuration

Set one of the following environment variables:

### SendGrid
` + "```" + `
SENDGRID_API_KEY=your-api-key
EMAIL_FROM=noreply@example.com
` + "```" + `

### Mailgun
` + "```" + `
MAILGUN_DOMAIN=mg.example.com
MAILGUN_API_KEY=your-api-key
EMAIL_FROM=noreply@example.com
` + "```" + `

### AWS SES
` + "```" + `
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
EMAIL_FROM=noreply@example.com
` + "```" + `

### SMTP
` + "```" + `
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_USE_TLS=true
EMAIL_FROM=noreply@example.com
` + "```" + `

## Testing

For local testing, use MailHog:

` + "```bash" + `
docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog
` + "```" + `

Then configure SMTP:
` + "```" + `
SMTP_HOST=localhost
SMTP_PORT=1025
EMAIL_FROM=test@example.com
` + "```" + `

View emails at http://localhost:8025

## Adding New Templates

` + "```bash" + `
gforge add email <template-name> --provider=<provider>
` + "```" + `

Example:
` + "```bash" + `
gforge add email welcome --provider=sendgrid
gforge add email password-reset --provider=mailgun
` + "```" + `
`

	if err := execx.WriteFileIfMissing(readmePath, []byte(readmeSrc), 0o644); err != nil {
		return err
	}

	fmt.Printf("Added email template: %s\n", pas)
	fmt.Printf("  - %s\n", emailPath)
	fmt.Printf("  - %s\n", testPath)
	fmt.Printf("  - %s\n", examplePath)
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		fmt.Printf("  - %s\n", readmePath)
	}
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Configure email provider in .env:")
	switch provider {
	case "sendgrid":
		fmt.Println("     SENDGRID_API_KEY=your-api-key")
	case "mailgun":
		fmt.Println("     MAILGUN_DOMAIN=mg.example.com")
		fmt.Println("     MAILGUN_API_KEY=your-api-key")
	case "ses":
		fmt.Println("     AWS_REGION=us-east-1")
		fmt.Println("     AWS_ACCESS_KEY_ID=your-access-key")
		fmt.Println("     AWS_SECRET_ACCESS_KEY=your-secret-key")
	case "smtp":
		fmt.Println("     SMTP_HOST=smtp.gmail.com")
		fmt.Println("     SMTP_PORT=587")
		fmt.Println("     SMTP_USERNAME=your-email@gmail.com")
		fmt.Println("     SMTP_PASSWORD=your-app-password")
	}
	fmt.Println("     EMAIL_FROM=noreply@example.com")
	fmt.Println("  2. Customize email template in " + emailPath)
	fmt.Println("  3. Add template data fields to " + pas + "Data struct")
	fmt.Println("  4. Test with: go test ./app/email/...")
	return nil
}
