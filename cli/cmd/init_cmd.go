package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/liuyukai/agentskills-cli/internal/provider"
)

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Create a new Skill skeleton directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cwd, _ := os.Getwd()

		// Resolve provider
		p := resolveProvider(cwd)

		// Create directory structure
		skillDir := filepath.Join(cwd, name)
		dirs := []string{
			skillDir,
			filepath.Join(skillDir, "scripts"),
			filepath.Join(skillDir, "references"),
			filepath.Join(skillDir, "assets"),
		}

		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("creating directory %s: %w", dir, err)
			}
		}

		// Generate SKILL.md template
		template := provider.SkillTemplate(p, name)
		skillMDPath := filepath.Join(skillDir, "SKILL.md")
		if err := os.WriteFile(skillMDPath, []byte(template), 0o644); err != nil {
			return fmt.Errorf("writing SKILL.md: %w", err)
		}

		fmt.Printf("Created %s/\n", name)
		fmt.Printf("  ├── SKILL.md        (template for %s)\n", p)
		fmt.Println("  ├── scripts/")
		fmt.Println("  ├── references/")
		fmt.Println("  └── assets/")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
