package run

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/unicoooorn/pingr/internal/app"
	"github.com/unicoooorn/pingr/internal/config"
)

func Register() *cobra.Command {
	return &cobra.Command{
		Use:  "run",
		RunE: run,
	}
}

func run(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if err := config.ValidateConfig(cfg); err != nil {
    return fmt.Errorf("Config error: %v", err)
	}

	return app.Run(ctx, *cfg)
}
