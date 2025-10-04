package cmd

import (
    "os"
    "path/filepath"
    "testing"
)

func chdirTemp(t *testing.T) string {
    t.Helper()
    cwd, _ := os.Getwd()
    dir := t.TempDir()
    if err := os.Chdir(dir); err != nil { t.Fatalf("chdir: %v", err) }
    t.Cleanup(func() { _ = os.Chdir(cwd) })
    return dir
}

func TestScaffoldPageAndRoute(t *testing.T) {
	_ = chdirTemp(t)
	if err := scaffoldPage("about"); err != nil { t.Fatalf("scaffoldPage: %v", err) }
	if _, err := os.Stat(filepath.Join("app", "templates", "page_about.go")); err != nil {
		t.Fatalf("missing template: %v", err)
	}
	if _, err := os.Stat(filepath.Join("app", "routes", "page_about.go")); err != nil {
		t.Fatalf("missing route: %v", err)
	}
}

func TestScaffoldComponent(t *testing.T) {
	_ = chdirTemp(t)
	if err := scaffoldComponent("Card"); err != nil { t.Fatalf("scaffoldComponent: %v", err) }
	if _, err := os.Stat(filepath.Join("app", "templates", "component_card.go")); err != nil {
		t.Fatalf("missing component: %v", err)
	}
}

func TestScaffoldAuth(t *testing.T) {
	_ = chdirTemp(t)
	if err := scaffoldAuth(); err != nil { t.Fatalf("scaffoldAuth: %v", err) }
	if _, err := os.Stat(filepath.Join("app", "routes", "auth.go")); err != nil {
		t.Fatalf("missing auth route: %v", err)
	}
	if _, err := os.Stat(filepath.Join("app", "templates", "auth_login.go")); err != nil {
		t.Fatalf("missing auth template: %v", err)
	}
}

func TestScaffoldOAuth(t *testing.T) {
	_ = chdirTemp(t)
	if err := scaffoldOAuth("github"); err != nil { t.Fatalf("scaffoldOAuth: %v", err) }
	if _, err := os.Stat(filepath.Join("app", "routes", "oauth_github.go")); err != nil {
		t.Fatalf("missing oauth route: %v", err)
	}
}

func TestScaffoldDB(t *testing.T) {
	_ = chdirTemp(t)
	if err := scaffoldDB("appdata"); err != nil { t.Fatalf("scaffoldDB: %v", err) }
	if _, err := os.Stat(filepath.Join("app", "db", "appdata.sql")); err != nil {
		t.Fatalf("missing sql: %v", err)
	}
}

func TestScaffoldModule(t *testing.T) {
	_ = chdirTemp(t)
	if err := scaffoldModule("blog"); err != nil { t.Fatalf("scaffoldModule: %v", err) }
	if _, err := os.Stat(filepath.Join("app", "templates", "page_blog.go")); err != nil {
		t.Fatalf("missing module page: %v", err)
	}
	if _, err := os.Stat(filepath.Join("app", "db", "blog.sql")); err != nil {
		t.Fatalf("missing module sql: %v", err)
	}
}
