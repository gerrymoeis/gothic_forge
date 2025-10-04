package cmd

import (
  "bufio"
  "fmt"
  "os"
  "path/filepath"
  "strings"

  "github.com/spf13/cobra"
)

var (
  secretsSet string
  secretsGet string
)

var secretsCmd = &cobra.Command{
  Use:   "secrets",
  Short: "Manage .env secrets (set/get)",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()
    envPath := filepath.Join(".env")
    if secretsSet == "" && secretsGet == "" {
      fmt.Println("Usage: gforge secrets --set KEY=VAL | --get KEY")
      return nil
    }
    // Ensure .env exists
    if _, err := os.Stat(envPath); os.IsNotExist(err) {
      if err := os.WriteFile(envPath, []byte(""), 0o600); err != nil {
        return err
      }
    }
    // Load existing
    kv := map[string]string{}
    if f, err := os.Open(envPath); err == nil {
      defer f.Close()
      sc := bufio.NewScanner(f)
      for sc.Scan() {
        line := strings.TrimSpace(sc.Text())
        if line == "" || strings.HasPrefix(line, "#") { continue }
        if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
          kv[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
        }
      }
    }
    if secretsGet != "" {
      if v, ok := kv[secretsGet]; ok {
        fmt.Printf("%s=%s\n", secretsGet, v)
      } else {
        fmt.Println("(not set)")
      }
      return nil
    }
    if secretsSet != "" {
      parts := strings.SplitN(secretsSet, "=", 2)
      if len(parts) != 2 {
        return fmt.Errorf("--set expects KEY=VAL")
      }
      key := strings.TrimSpace(parts[0])
      val := strings.TrimSpace(parts[1])
      kv[key] = val
      // Rewrite .env
      b := &strings.Builder{}
      for k, v := range kv { fmt.Fprintf(b, "%s=%s\n", k, v) }
      return os.WriteFile(envPath, []byte(b.String()), 0o600)
    }
    return nil
  },
}

func init() {
  secretsCmd.Flags().StringVar(&secretsSet, "set", "", "set KEY=VAL")
  secretsCmd.Flags().StringVar(&secretsGet, "get", "", "get KEY")
  rootCmd.AddCommand(secretsCmd)
}
