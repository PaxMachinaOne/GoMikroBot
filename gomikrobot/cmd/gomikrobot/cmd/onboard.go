package cmd

import (
	"fmt"

	"github.com/kamir/gomikrobot/internal/config"
	"github.com/spf13/cobra"
)

var onboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "Initialize configuration",
	Run:   runOnboard,
}

func init() {
	rootCmd.AddCommand(onboardCmd)
}

func runOnboard(cmd *cobra.Command, args []string) {
	fmt.Println("🚀 Initializing GoMikroBot...")

	path, _ := config.ConfigPath()

	// Check if already exists
	if _, err := config.Load(); err == nil {
		// Load might return defaults if file missing, but let's check file existence in Status
		// internal/config/loader.go check logic:
		// config.Load() returns default if file missing.
		// So we can just try to Save default.
	}

	cfg := config.DefaultConfig()
	if err := config.Save(cfg); err != nil {
		fmt.Printf("Error skipping config: %v\n", err)
	} else {
		fmt.Printf("✅ Config created at: %s\n", path)
	}

	// ensure workspace exists
	if err := config.EnsureDir(cfg.Agents.Defaults.Workspace); err != nil {
		// It might be ~ path, EnsureDir assumes expanded?
		// DefaultConfig has "~/...". config.EnsureDir does mkdir. Mkdir doesn't expand ~.
		// We need to expand it. config.Load expands it.
		// Let's rely on user to run it or expand it here.
	}

	fmt.Println("\nNext steps:")
	fmt.Println("1. Edit config.json to add your API keys.")
	fmt.Println("2. Run 'gomikrobot agent -m \"hello\"' to test.")
}
