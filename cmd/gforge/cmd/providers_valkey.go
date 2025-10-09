package cmd

import (
  "bufio"
  "context"
  "errors"
  "fmt"
  "os"
  "strings"
  "time"
)

// valkeyInteractiveProvision prompts for a Redis/Valkey connection string (REDIS_URL)
// and writes it to .env. Returns the URL if provided.
func valkeyInteractiveProvision(_ context.Context, dryRun bool) (string, error) {
  cur := strings.TrimSpace(os.Getenv("REDIS_URL"))
  if cur != "" { return cur, nil }
  if dryRun {
    fmt.Println("  • Valkey (dry-run): would prompt for REDIS_URL and write to .env")
    return "", nil
  }
  fmt.Println("  • Valkey: Please paste your Redis/Valkey connection URI (REDIS_URL)")
  fmt.Println("    Examples:")
  fmt.Println("      - rediss://default:<password>@<host>:<port>/0")
  fmt.Println("      - redis://:<password>@<host>:<port>/0 (non-TLS)")
  fmt.Print("    REDIS_URL= ")
  reader := bufio.NewReader(os.Stdin)
  url, _ := reader.ReadString('\n')
  url = strings.TrimSpace(url)
  if url == "" { return "", errors.New("no REDIS_URL provided") }
  kv := map[string]string{"REDIS_URL": url}
  if err := updateEnvFileInPlace(".env", kv); err != nil {
    if f, ferr := os.OpenFile(".env", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600); ferr == nil {
      defer f.Close()
      _, _ = f.WriteString("\n# Added by gforge deploy wizard\nREDIS_URL=" + url + "\n")
    }
  }
  _ = os.Setenv("REDIS_URL", url)
  fmt.Println("    → REDIS_URL saved to .env")
  return url, nil
}

// valkeyAutoProvision plans auto-provision of a Valkey (Redis-compatible) instance on Aiven.
// For now, this function prints a plan in dry-run and returns an informative error when executed
// without full implementation. Future work: use Aiven API to create service and compose REDIS_URL.
func valkeyAutoProvision(ctx context.Context, dryRun bool) (string, error) {
  // If already present, nothing to do
  if cur := strings.TrimSpace(os.Getenv("REDIS_URL")); cur != "" { return cur, nil }

  project := strings.TrimSpace(os.Getenv("AIVEN_PROJECT"))
  cloud := strings.TrimSpace(os.Getenv("AIVEN_CLOUD")) // e.g., aws-us-east-1
  plan := strings.TrimSpace(os.Getenv("AIVEN_PLAN"))   // e.g., startup-4, hobbyist
  serviceName := strings.TrimSpace(os.Getenv("AIVEN_SERVICE_NAME"))
  if serviceName == "" { serviceName = "gforge-valkey" }

  if dryRun {
    fmt.Println("  • Valkey (dry-run): would create/find Aiven Valkey service")
    fmt.Printf("    - project: %s\n", project)
    fmt.Printf("    - cloud: %s\n", cloud)
    fmt.Printf("    - plan: %s\n", plan)
    fmt.Printf("    - service: %s\n", serviceName)
    return "", nil
  }

  // Guard: require token for any non-dry-run auto provisioning
  tok := strings.TrimSpace(os.Getenv("AIVEN_TOKEN"))
  if tok == "" {
    return "", errors.New("AIVEN_TOKEN is not set; cannot auto-provision Valkey")
  }
  if project == "" { return "", errors.New("AIVEN_PROJECT is required for Valkey auto-provision") }
  if plan == "" { return "", errors.New("AIVEN_PLAN is required for Valkey auto-provision") }

  cli, err := newAivenClientFromEnv()
  if err != nil { return "", err }

  // 1) Check existing service
  var svc *aivenService
  if s, err := cli.getService(ctx, project, serviceName); err == nil {
    svc = s
  }
  // 2) Create if missing
  if svc == nil {
    req := aivenCreateServiceReq{
      Cloud:       cloud,
      Plan:        plan,
      ServiceName: serviceName,
      ServiceType: "valkey",
    }
    s, err := cli.createService(ctx, project, req)
    if err != nil { return "", err }
    if err := cli.waitServiceRunning(ctx, project, s.Name, 15*time.Minute); err != nil { return "", err }
    svc = s
  } else {
    if err := cli.waitServiceRunning(ctx, project, svc.Name, 10*time.Minute); err != nil { return "", err }
  }

  // 3) Compose REDIS_URL
  url := strings.TrimSpace(svc.ServiceURI)
  if url == "" {
    return "", errors.New("Aiven service_uri is empty; cannot compose REDIS_URL")
  }

  // 4) Persist and reflect
  kv := map[string]string{"REDIS_URL": url}
  if err := updateEnvFileInPlace(".env", kv); err != nil {
    if f, ferr := os.OpenFile(".env", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600); ferr == nil {
      defer f.Close()
      _, _ = f.WriteString("\n# Added by gforge deploy wizard\nREDIS_URL=" + url + "\n")
    }
  }
  _ = os.Setenv("REDIS_URL", url)
  return url, nil
}
