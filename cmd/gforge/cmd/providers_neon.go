package cmd

import (
  "bufio"
  "context"
  "errors"
  "fmt"
  "net/http"
  "os"
  "strings"
  "time"
  "encoding/json"
  "bytes"
  "io"
)

// neonInteractiveProvision prompts the user to paste a Neon Postgres connection string (DATABASE_URL)
// and writes it to .env. This is a safe, minimal Phase 2 implementation without invoking
// unverified Neon API endpoints. It returns the DSN if provided.
func neonInteractiveProvision(_ context.Context, dryRun bool) (string, error) {
  // If already present, nothing to do
  cur := strings.TrimSpace(os.Getenv("DATABASE_URL"))
  if cur != "" { return cur, nil }
  if dryRun {
    fmt.Println("  • Neon (dry-run): would prompt for DATABASE_URL and write to .env, then run migrations")
    return "", nil
  }
  fmt.Println("  • Neon: Please paste your Neon Postgres connection string (DATABASE_URL)")
  fmt.Println("    Example: postgres://user:password@host/dbname?sslmode=require")
  fmt.Print("    DATABASE_URL= ")
  reader := bufio.NewReader(os.Stdin)
  dsn, _ := reader.ReadString('\n')
  dsn = strings.TrimSpace(dsn)
  if dsn == "" { return "", errors.New("no DATABASE_URL provided") }
  // Persist to .env using helper from deploy.go
  kv := map[string]string{"DATABASE_URL": dsn}
  if err := updateEnvFileInPlace(".env", kv); err != nil {
    // best-effort fallback: append at end
    if f, ferr := os.OpenFile(".env", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600); ferr == nil {
      defer f.Close()
      _, _ = f.WriteString("\n# Added by gforge deploy wizard\nDATABASE_URL=" + dsn + "\n")
    }
  }
  _ = os.Setenv("DATABASE_URL", dsn)
  fmt.Println("    → DATABASE_URL saved to .env")
  return dsn, nil
}

// Base URL for Neon API v2
const neonAPIBase = "https://console.neon.tech/api/v2"

type neonClient struct {
  token string
  hc    *http.Client
}

func newNeonClientFromEnv() (*neonClient, error) {
  tok := strings.TrimSpace(os.Getenv("NEON_TOKEN"))
  if tok == "" { return nil, errors.New("NEON_TOKEN is not set") }
  return &neonClient{ token: tok, hc: &http.Client{ Timeout: 30 * time.Second } }, nil
}

func (c *neonClient) do(ctx context.Context, method, path string, in any, out any) error {
  var body io.Reader
  if in != nil {
    b, err := json.Marshal(in)
    if err != nil { return err }
    body = bytes.NewReader(b)
  }
  req, err := http.NewRequestWithContext(ctx, method, neonAPIBase+path, body)
  if err != nil { return err }
  req.Header.Set("Authorization", "Bearer "+c.token)
  req.Header.Set("Content-Type", "application/json")
  resp, err := c.hc.Do(req)
  if err != nil { return err }
  defer resp.Body.Close()
  if resp.StatusCode >= 300 {
    b, _ := io.ReadAll(resp.Body)
    return fmt.Errorf("neon api %s %s: %s: %s", method, path, resp.Status, string(b))
  }
  if out != nil {
    dec := json.NewDecoder(resp.Body)
    if err := dec.Decode(out); err != nil { return err }
  }
  return nil
}

// Minimal shapes (only fields we need). The Neon API returns nested envelopes.
type neonListProjects struct { Projects []neonProject `json:"projects"` }
type neonProject struct {
  ID   string `json:"id"`
  Name string `json:"name"`
}
type neonCreateProjectResp struct {
  Project  neonProject    `json:"project"`
  Branch   neonBranch     `json:"branch"`
  Endpoint neonEndpoint   `json:"endpoint"`
  Op       *neonOperation `json:"operation"`
}
type neonBranch struct {
  ID   string `json:"id"`
  Name string `json:"name"`
}
type neonEndpoint struct {
  ID       string `json:"id"`
  Host     string `json:"host"`
  Type     string `json:"type"`
  BranchID string `json:"branch_id"`
}
type neonOperation struct {
  ID     string `json:"id"`
  Status string `json:"status"`
}
type neonListBranches struct { Branches []neonBranch `json:"branches"` }
type neonListEndpoints struct { Endpoints []neonEndpoint `json:"endpoints"` }
type neonGetOperation struct { Operation neonOperation `json:"operation"` }

func (c *neonClient) listProjects(ctx context.Context) ([]neonProject, error) {
  var out neonListProjects
  if err := c.do(ctx, http.MethodGet, "/projects", nil, &out); err != nil { return nil, err }
  return out.Projects, nil
}

func (c *neonClient) createProject(ctx context.Context, name, regionID string) (*neonCreateProjectResp, error) {
  body := map[string]any{
    "project": map[string]any{
      "name": name,
    },
  }
  if strings.TrimSpace(regionID) != "" {
    body["project"].(map[string]any)["region_id"] = regionID
  }
  var out neonCreateProjectResp
  if err := c.do(ctx, http.MethodPost, "/projects", body, &out); err != nil { return nil, err }
  return &out, nil
}

func (c *neonClient) listBranches(ctx context.Context, projectID string) ([]neonBranch, error) {
  var out neonListBranches
  if err := c.do(ctx, http.MethodGet, "/projects/"+projectID+"/branches", nil, &out); err != nil { return nil, err }
  return out.Branches, nil
}

func (c *neonClient) createBranch(ctx context.Context, projectID, name string) (*neonOperation, *neonBranch, error) {
  body := map[string]any{ "branch": map[string]any{ "name": name } }
  var out struct{ Branch neonBranch `json:"branch"`; Operation *neonOperation `json:"operation"` }
  if err := c.do(ctx, http.MethodPost, "/projects/"+projectID+"/branches", body, &out); err != nil { return nil, nil, err }
  return out.Operation, &out.Branch, nil
}

func (c *neonClient) listEndpoints(ctx context.Context, projectID string) ([]neonEndpoint, error) {
  var out neonListEndpoints
  if err := c.do(ctx, http.MethodGet, "/projects/"+projectID+"/endpoints", nil, &out); err != nil { return nil, err }
  return out.Endpoints, nil
}

func (c *neonClient) createEndpoint(ctx context.Context, projectID, branchID string) (*neonOperation, *neonEndpoint, error) {
  body := map[string]any{ "endpoint": map[string]any{ "branch_id": branchID, "type": "read_write" } }
  var out struct{ Endpoint neonEndpoint `json:"endpoint"`; Operation *neonOperation `json:"operation"` }
  if err := c.do(ctx, http.MethodPost, "/projects/"+projectID+"/endpoints", body, &out); err != nil { return nil, nil, err }
  return out.Operation, &out.Endpoint, nil
}

func (c *neonClient) createRole(ctx context.Context, projectID, branchID, roleName, password string) error {
  body := map[string]any{ "role": map[string]any{ "name": roleName, "password": password } }
  var out map[string]any
  return c.do(ctx, http.MethodPost, "/projects/"+projectID+"/branches/"+branchID+"/roles", body, &out)
}

func (c *neonClient) createDatabase(ctx context.Context, projectID, branchID, dbName, ownerRole string) error {
  body := map[string]any{ "database": map[string]any{ "name": dbName, "owner_name": ownerRole } }
  var out map[string]any
  return c.do(ctx, http.MethodPost, "/projects/"+projectID+"/branches/"+branchID+"/databases", body, &out)
}

func (c *neonClient) getOperation(ctx context.Context, opID string) (*neonOperation, error) {
  var out neonGetOperation
  if err := c.do(ctx, http.MethodGet, "/operations/"+opID, nil, &out); err != nil { return nil, err }
  return &out.Operation, nil
}

func (c *neonClient) pollOperation(ctx context.Context, op *neonOperation, timeout time.Duration) error {
  if op == nil || op.ID == "" { return nil }
  deadline := time.Now().Add(timeout)
  for {
    if time.Now().After(deadline) { return errors.New("operation timeout") }
    cur, err := c.getOperation(ctx, op.ID)
    if err != nil { return err }
    st := strings.ToLower(strings.TrimSpace(cur.Status))
    if st == "finished" || st == "succeeded" || st == "ready" || st == "completed" { return nil }
    if st == "failed" || st == "error" { return fmt.Errorf("operation failed: %s", cur.Status) }
    time.Sleep(2 * time.Second)
  }
}

// neonAutoProvision provisions a Neon project/branch/endpoint/role/db and returns a DATABASE_URL.
func neonAutoProvision(ctx context.Context, dryRun bool) (string, error) {
  // If already present, nothing to do
  if cur := strings.TrimSpace(os.Getenv("DATABASE_URL")); cur != "" { return cur, nil }

  // Config
  projectName := strings.TrimSpace(os.Getenv("NEON_PROJECT_NAME"))
  if projectName == "" { projectName = "gothic-forge-v3" }
  branchName := strings.TrimSpace(os.Getenv("NEON_BRANCH_NAME"))
  if branchName == "" { branchName = "main" }
  dbName := strings.TrimSpace(os.Getenv("NEON_DB_NAME"))
  if dbName == "" { dbName = "appdb" }
  roleName := strings.TrimSpace(os.Getenv("NEON_DB_USER"))
  if roleName == "" { roleName = "app" }
  password := strings.TrimSpace(os.Getenv("NEON_DB_PASSWORD"))
  if password == "" { password = genSecret() }
  regionID := strings.TrimSpace(os.Getenv("NEON_REGION"))

  if dryRun {
    fmt.Println("  • Neon (dry-run): would create/find project, branch, endpoint, role, database")
    fmt.Printf("    - project: %s\n", projectName)
    fmt.Printf("    - branch: %s\n", branchName)
    fmt.Printf("    - db: %s, user: %s\n", dbName, roleName)
    if regionID != "" { fmt.Printf("    - region: %s\n", regionID) } else { fmt.Println("    - region: (provider default)") }
    return "", nil
  }

  cli, err := newNeonClientFromEnv()
  if err != nil { return "", err }

  // 1) Project (find by name, else create)
  ctxT, cancel := context.WithTimeout(ctx, 2*time.Minute)
  defer cancel()

  var projectID string
  if projs, err := cli.listProjects(ctxT); err == nil {
    for _, p := range projs { if strings.EqualFold(p.Name, projectName) { projectID = p.ID; break } }
  }
  var branchID string
  var endpointHost string
  if projectID == "" {
    fmt.Printf("  • creating project: %s", projectName)
    if regionID != "" { fmt.Printf(" (region: %s)", regionID) }
    fmt.Println()
    resp, err := cli.createProject(ctxT, projectName, regionID)
    if err != nil { return "", err }
    if resp.Op != nil { if err := cli.pollOperation(ctxT, resp.Op, 2*time.Minute); err != nil { return "", err } }
    projectID = resp.Project.ID
    branchID = resp.Branch.ID
    endpointHost = resp.Endpoint.Host
    fmt.Printf("    → created project id: %s, branch id: %s, endpoint: %s\n", projectID, branchID, endpointHost)
  } else {
    fmt.Printf("  • found project: %s (id: %s)\n", projectName, projectID)
  }

  // 2) Branch ensure
  if branchID == "" {
    if brs, err := cli.listBranches(ctxT, projectID); err == nil {
      for _, b := range brs { if strings.EqualFold(b.Name, branchName) { branchID = b.ID; break } }
    }
    if branchID == "" {
      fmt.Printf("  • creating branch: %s\n", branchName)
      op, b, err := cli.createBranch(ctxT, projectID, branchName)
      if err != nil { return "", err }
      if op != nil { if err := cli.pollOperation(ctxT, op, 2*time.Minute); err != nil { return "", err } }
      branchID = b.ID
      fmt.Printf("    → branch id: %s\n", branchID)
    } else {
      fmt.Printf("  • found branch: %s (id: %s)\n", branchName, branchID)
    }
  }

  // 3) Endpoint ensure (read_write for branch)
  if endpointHost == "" {
    if eps, err := cli.listEndpoints(ctxT, projectID); err == nil {
      for _, e := range eps {
        if strings.EqualFold(e.BranchID, branchID) && (e.Type == "read_write" || e.Type == "primary" || e.Type == "RW") {
          endpointHost = e.Host
          if endpointHost != "" { break }
        }
      }
    }
    if endpointHost == "" {
      fmt.Printf("  • creating endpoint for branch id: %s\n", branchID)
      op, e, err := cli.createEndpoint(ctxT, projectID, branchID)
      if err != nil { return "", err }
      if op != nil { if err := cli.pollOperation(ctxT, op, 2*time.Minute); err != nil { return "", err } }
      endpointHost = e.Host
      fmt.Printf("    → endpoint: %s\n", endpointHost)
    } else {
      fmt.Printf("  • found endpoint host: %s\n", endpointHost)
    }
  }

  // 4) Role ensure (ignore error if already exists)
  fmt.Printf("  • ensuring role: %s\n", roleName)
  _ = cli.createRole(ctxT, projectID, branchID, roleName, password)

  // 5) Database ensure (ignore error if already exists)
  fmt.Printf("  • ensuring database: %s (owner: %s)\n", dbName, roleName)
  _ = cli.createDatabase(ctxT, projectID, branchID, dbName, roleName)

  // 6) Compose DSN
  if endpointHost == "" { return "", errors.New("neon endpoint host is empty") }
  dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=require", roleName, password, endpointHost, dbName)

  // Persist to .env
  kv := map[string]string{
    "DATABASE_URL": dsn,
    "NEON_DB_PASSWORD": password,
  }
  if err := updateEnvFileInPlace(".env", kv); err != nil {
    f, ferr := os.OpenFile(".env", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
    if ferr == nil { defer f.Close(); _, _ = f.WriteString("\n# Added by gforge deploy wizard\nDATABASE_URL="+dsn+"\nNEON_DB_PASSWORD="+password+"\n") }
  }
  _ = os.Setenv("DATABASE_URL", dsn)
  fmt.Println("  • DATABASE_URL saved to .env")
  return dsn, nil
}
