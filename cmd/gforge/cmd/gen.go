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
  Short: "Scaffold in app/ (page/component) — alias of 'add'",
  RunE: func(cmd *cobra.Command, args []string) error {
    banner()
    if genPage == "" && genComponent == "" {
      fmt.Println("Usage: gforge gen --page Name | --component Name")
      return nil
    }
    if genPage != "" {
      if !isValidName(genPage) { return fmt.Errorf("invalid page name: %s", genPage) }
      if err := scaffoldPage(genPage); err != nil { return err }
    }
    if genComponent != "" {
      if !isValidName(genComponent) { return fmt.Errorf("invalid component name: %s", genComponent) }
      if err := scaffoldComponent(genComponent); err != nil { return err }
    }
    return nil
  },
}

func init() {
  genCmd.Flags().StringVar(&genPage, "page", "", "page name")
  genCmd.Flags().StringVar(&genComponent, "component", "", "component name")
  rootCmd.AddCommand(genCmd)
}
