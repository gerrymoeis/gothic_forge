package cmd

import (
    "fmt"
    "path/filepath"
    "regexp"
    "strings"

    "gothicforge3/internal/execx"
    "github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
    Use:   "add",
    Short: "Scaffold features in app/ (page, component, auth, oauth, db, module)",
    Args:  cobra.MinimumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        banner()
        kind := strings.ToLower(args[0])
        var name string
        if kind == "page" || kind == "component" || kind == "oauth" || kind == "db" || kind == "module" || kind == "crud" {
            if len(args) < 2 {
                fmt.Println("Usage:")
                fmt.Println("  gforge add page <name>")
                fmt.Println("  gforge add component <name>")
                fmt.Println("  gforge add auth")
                fmt.Println("  gforge add oauth <provider>")
                fmt.Println("  gforge add db <name>")
                fmt.Println("  gforge add module <name>")
                fmt.Println("  gforge add crud <name>")
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
        default:
            fmt.Println("Usage:")
            fmt.Println("  gforge add page <name>")
            fmt.Println("  gforge add component <name>")
            fmt.Println("  gforge add auth")
            fmt.Println("  gforge add oauth <provider>")
            fmt.Println("  gforge add db <name>")
            fmt.Println("  gforge add module <name>")
            fmt.Println("  gforge add crud <name>")
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
