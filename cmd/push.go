package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/liuyukai/agentskills/internal/api"
	"github.com/liuyukai/agentskills/internal/bundle"
	"github.com/liuyukai/agentskills/internal/config"
	"github.com/liuyukai/agentskills/internal/parser"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push [path]",
	Short: "Pack and upload a Skill bundle",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		// Validate SKILL.md locally first
		skillMDPath := filepath.Join(path, "SKILL.md")
		content, err := os.ReadFile(skillMDPath)
		if err != nil {
			return fmt.Errorf("cannot read %s: %w", skillMDPath, err)
		}

		fmt.Print("Validating SKILL.md...        ")
		meta, _, err := parser.ParseSKILLMD(content)
		if err != nil {
			fmt.Println("✗")
			return fmt.Errorf("parse error: %w", err)
		}
		if err := parser.Validate(meta); err != nil {
			fmt.Println("✗")
			return fmt.Errorf("validation error: %w", err)
		}
		fmt.Println("✓")

		// Pack to temp file
		fmt.Print("Packing bundle...             ")
		tmpFile, err := os.CreateTemp("", "agentskills-*.tar.gz")
		if err != nil {
			return err
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpPath)

		if err := bundle.Pack(path, tmpPath); err != nil {
			fmt.Println("✗")
			return fmt.Errorf("pack error: %w", err)
		}

		info, _ := os.Stat(tmpPath)
		fmt.Printf("✓ (%s)\n", formatSize(info.Size()))

		// Compute local checksum
		localChecksum, err := bundle.SHA256File(tmpPath)
		if err != nil {
			return err
		}

		// Upload
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		client := api.New(cfg.APIURL, cfg.Token)
		fmt.Printf("Uploading %s@%s...   ", meta.Name, meta.Version)

		result, err := client.Publish(tmpPath)
		if err != nil {
			fmt.Println("✗")
			return err
		}
		fmt.Println("✓")

		fmt.Printf("Checksum: sha256:%s\n", localChecksum)

		// Verify server checksum matches
		if serverChecksum, ok := result["checksum"].(string); ok {
			expected := "sha256:" + localChecksum
			if serverChecksum != expected {
				return fmt.Errorf("checksum mismatch: local=%s server=%s", expected, serverChecksum)
			}
		}

		fmt.Printf("\nPublished %s@%s successfully.\n", meta.Name, meta.Version)
		return nil
	},
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
