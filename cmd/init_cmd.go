package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Create a new Skill skeleton",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		base := filepath.Join(".", name)

		dirs := []string{
			base,
			filepath.Join(base, "scripts"),
			filepath.Join(base, "references"),
			filepath.Join(base, "assets"),
		}
		for _, d := range dirs {
			if err := os.MkdirAll(d, 0755); err != nil {
				return fmt.Errorf("mkdir %s: %w", d, err)
			}
		}

		skillMD := fmt.Sprintf(`---
name: "%s"
version: "0.1.0"
description: ""
author: ""
tags: []
---

# %s

Describe your skill here.
`, name, name)

		if err := os.WriteFile(filepath.Join(base, "SKILL.md"), []byte(skillMD), 0644); err != nil {
			return fmt.Errorf("write SKILL.md: %w", err)
		}

		fmt.Printf("Created %s/\n", name)
		fmt.Printf("  ├── SKILL.md\n")
		fmt.Printf("  ├── scripts/\n")
		fmt.Printf("  ├── references/\n")
		fmt.Printf("  └── assets/\n")
		return nil
	},
}
