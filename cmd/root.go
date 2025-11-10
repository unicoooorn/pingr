package cmd

import (
	"github.com/spf13/cobra"
	"github.com/unicoooorn/pingr/cmd/run"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:              "bio-account",
		Short:            "REST-API service for downloading vectors from EBS",
		TraverseChildren: true,
	}

	rootCmd.AddCommand(run.Register())

	rootCmd.PersistentFlags().StringP("config", "c", "config/config.yaml", "Specify a config file")

	return rootCmd
}
