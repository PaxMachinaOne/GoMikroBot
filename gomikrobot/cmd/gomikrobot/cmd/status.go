package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gomikrobot version %s\n", version)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show system status",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(color.CyanString(logo))
		fmt.Println("📊 GoMikroBot Status")
		fmt.Println("─────────────────────")
		fmt.Printf("Version: %s\n", version)

		// Check config
		home, _ := os.UserHomeDir()
		configPath := filepath.Join(home, ".gomikrobot", "config.json")
		if _, err := os.Stat(configPath); err == nil {
			fmt.Println("Config:  ✓ Found (" + configPath + ")")
		} else {
			fmt.Println("Config:  ✗ Not found (run 'gomikrobot onboard' first)")
		}

		fmt.Println("Status:  Ready")
	},
}
