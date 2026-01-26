package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rakunlabs/into"
	"github.com/rytsh/kup/internal/config"
	"github.com/rytsh/kup/internal/tui"
)

func main() {
	into.Init(
		run,
		into.WithMsgf("kup"),
		into.WithStartFn(nil), into.WithStopFn(nil),
	)
}

func run(ctx context.Context) error {
	// Load configuration
	cfg, err := config.Load(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		// Use default config if loading fails
		cfg = config.DefaultConfig()
	}

	// Ensure bin directory exists
	if err := os.MkdirAll(cfg.BinPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not create bin directory: %v\n", err)
	}

	// Run the TUI
	return tui.Run(cfg)
}
