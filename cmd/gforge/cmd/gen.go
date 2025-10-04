package cmd

import (
  "fmt"
  "github.com/spf13/cobra"
)

var (
  genPage      string
  genComponent string
)

var genCmd = &cobra.Command{
  Use:   "gen",
  Short: "Scaffold in app/ (page/component)",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()
    if genPage == "" && genComponent == "" {
      fmt.Println("Usage: gforge gen --page Name | --component Name")
      return nil
    }
    if genPage != "" {
      fmt.Printf("Scaffold page: %s (app/templates)\n", genPage)
    }
    if genComponent != "" {
      fmt.Printf("Scaffold component: %s (app/templates)\n", genComponent)
    }
    return nil
  },
}

func init() {
  genCmd.Flags().StringVar(&genPage, "page", "", "page name")
  genCmd.Flags().StringVar(&genComponent, "component", "", "component name")
  rootCmd.AddCommand(genCmd)
}
