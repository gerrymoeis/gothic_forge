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

// valkeyInteractiveProvision prompts for a Redis/Valkey connection string (VALKEY_URL)
// and writes it to .env. Returns the URL if provided.
func valkeyInteractiveProvision(_ context.Context, dryRun bool) (string, error) {
  // Check both new and legacy env vars
  cur := strings.TrimSpace(os.Getenv("VALKEY_URL"))
  if cur == "" {
    cur = strings.TrimSpace(os.Getenv("REDIS_URL"))
  }
  if cur != "" { return cur, nil }
  if dryRun {
    fmt.Println("  • Valkey (dry-run): would prompt for VALKEY_URL and write to .env")
    return "", nil
  }
  fmt.Println("  • Valkey: configuring cache connection")
  fmt.Println("")
  fmt.Println("    ╔══════════════════════════════════════════════════════════════╗")
  fmt.Println("    ║  How to get your Valkey/Redis connection URL:               ║")
  fmt.Println("    ╚══════════════════════════════════════════════════════════════╝")
  fmt.Println("")
  fmt.Println("    1. Go to Aiven Console: https://console.aiven.io/")
  fmt.Println("       (Sign up for FREE if you don't have an account)")
  fmt.Println("")
  fmt.Println("    2. Create a Valkey service:")
  fmt.Println("       • Click 'Create service'")
  fmt.Println("       • Select 'Valkey' (Redis-compatible)")
  fmt.Println("       • Choose Free/Developer plan (single-node). Confirm memory quota and limits in the console")
  fmt.Println("       • Select region closest to your users")
  fmt.Println("       • Click 'Create service' and wait ~5 minutes")
  fmt.Println("")
  fmt.Println("    3. Get connection URL:")
  fmt.Println("       • Open your Valkey service")
  fmt.Println("       • Go to 'Overview' tab")
  fmt.Println("       • Find 'Connection Information' section")
  fmt.Println("       • Copy the 'Service URI' (looks like: rediss://default:password@host:port)")
  fmt.Println("")
  fmt.Println("    4. Paste the connection URL below")
  fmt.Println("")
  fmt.Println("    Connection URL format:")
  fmt.Println("      rediss://default:<password>@<host>:<port>/0  (with TLS, recommended)")
  fmt.Println("      redis://:<password>@<host>:<port>/0          (without TLS)")
  fmt.Println("")
  fmt.Print("    VALKEY_URL (or press ENTER to skip): ")
  reader := bufio.NewReader(os.Stdin)
  url, _ := reader.ReadString('\n')
  url = strings.TrimSpace(url)
  if url == "" { 
    fmt.Println("    ⚠️  Skipped: No VALKEY_URL provided. Cache features will be disabled.")
    fmt.Println("       You can add it later to .env: VALKEY_URL=<your-connection-url>")
    return "", nil  // Return nil error to allow deployment to continue
  }
  kv := map[string]string{"VALKEY_URL": url}
  if err := updateEnvFileInPlace(".env", kv); err != nil {
    if f, ferr := os.OpenFile(".env", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600); ferr == nil {
      defer f.Close()
      _, _ = f.WriteString("\n# Added by gforge deploy wizard\nVALKEY_URL=" + url + "\n")
    }
  }
  _ = os.Setenv("VALKEY_URL", url)
  fmt.Println("    → VALKEY_URL saved to .env")
  return url, nil
}

// valkeyAutoProvision plans auto-provision of a Valkey (Redis-compatible) instance on Aiven.
// For now, this function prints a plan in dry-run and returns an informative error when executed
// without full implementation. Future work: use Aiven API to create service and compose VALKEY_URL.
func valkeyAutoProvision(ctx context.Context, dryRun bool) (string, error) {
  // If already present, nothing to do (check both new and legacy)
  cur := strings.TrimSpace(os.Getenv("VALKEY_URL"))
  if cur == "" {
    cur = strings.TrimSpace(os.Getenv("REDIS_URL"))
  }
  if cur != "" { return cur, nil }

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

  // 3) Compose VALKEY_URL
  url := strings.TrimSpace(svc.ServiceURI)
  if url == "" {
    return "", errors.New("Aiven service_uri is empty; cannot compose VALKEY_URL")
  }

  // 4) Persist and reflect
  kv := map[string]string{"VALKEY_URL": url}
  if err := updateEnvFileInPlace(".env", kv); err != nil {
    if f, ferr := os.OpenFile(".env", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600); ferr == nil {
      defer f.Close()
      _, _ = f.WriteString("\n# Added by gforge deploy wizard\nVALKEY_URL=" + url + "\n")
    }
  }
  _ = os.Setenv("VALKEY_URL", url)
  return url, nil
}
