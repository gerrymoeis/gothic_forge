package cmd

import (
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Scaffold optional features (page, api, db, auth)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}
