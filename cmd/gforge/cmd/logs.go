package cmd

import (
  "context"
  "fmt"
  "strings"
  "github.com/spf13/cobra"
  "gothicforge3/internal/execx"
)

var (
  logsFollow bool
  logsSince  string
)

var logsCmd = &cobra.Command{
  Use:   "logs",
  Short: "Tail application logs (provider-integrated)",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()
    ctx := context.Background()
    if _, ok := execx.Look("railway"); !ok {
      fmt.Println("Railway CLI not found.")
      printRailwayInstallHelp()
      return nil
    }
    if !isRailwayLinkedCLI(ctx) {
      fmt.Println("No Railway project link detected. Run 'railway link' first, then re-run: gforge logs")
      return nil
    }
    argv := []string{"logs"}
    if logsFollow { argv = append(argv, "--follow") }
    if s := strings.TrimSpace(logsSince); s != "" { argv = append(argv, "--since", s) }
    fmt.Println("Streaming logs via Railway...")
    full := append([]string{"railway", "logs"}, argv[1:]...)
    if err := execx.RunInteractive(ctx, "railway logs", full...); err != nil { return err }
    return nil
  },
}

func init() {
  logsCmd.Flags().BoolVar(&logsFollow, "follow", false, "follow output")
  logsCmd.Flags().StringVar(&logsSince, "since", "", "show logs since timestamp")
  rootCmd.AddCommand(logsCmd)
}
