package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/liuyukai/agentskills/internal/api"
	"github.com/liuyukai/agentskills/internal/bundle"
	"github.com/liuyukai/agentskills/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pullCmd)
}

var pullCmd = &cobra.Command{
	Use:   "pull <name>[@version]",
	Short: "Download a Skill bundle",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameVersion := args[0]

		name, version := parseNameVersion(nameVersion)

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		client := api.New(cfg.APIURL, cfg.Token)

		// If no version specified, get latest
		if version == "" {
			fmt.Printf("Fetching latest version of %s...\n", name)
			info, err := client.GetSkill(name)
			if err != nil {
				return err
			}
			if lv, ok := info["latest_version"].(map[string]interface{}); ok {
				if v, ok := lv["version"].(string); ok {
					version = v
				}
			}
			if version == "" {
				return fmt.Errorf("no versions found for %s", name)
			}
		}

		fmt.Printf("Downloading %s@%s...  ", name, version)
		body, serverChecksum, _, err := client.Download(name, version)
		if err != nil {
			fmt.Println("✗")
			return err
		}
		defer body.Close()

		// Save to temp file
		tmpFile, err := os.CreateTemp("", "agentskills-pull-*.tar.gz")
		if err != nil {
			return err
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		if _, err := io.Copy(tmpFile, body); err != nil {
			tmpFile.Close()
			fmt.Println("✗")
			return err
		}
		tmpFile.Close()
		fmt.Println("✓")

		// Verify checksum
		fmt.Print("Verifying checksum...          ")
		localChecksum, err := bundle.SHA256File(tmpPath)
		if err != nil {
			return err
		}
		if serverChecksum != "" && serverChecksum != localChecksum {
			fmt.Println("✗")
			return fmt.Errorf("checksum mismatch: server=%s local=%s", serverChecksum, localChecksum)
		}
		fmt.Println("✓")

		// Unpack
		destDir := "./" + name
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return err
		}
		if err := bundle.Unpack(tmpPath, destDir); err != nil {
			return fmt.Errorf("unpack: %w", err)
		}

		fmt.Printf("Extracted to %s/\n", destDir)
		return nil
	},
}

func parseNameVersion(s string) (name, version string) {
	if idx := strings.LastIndex(s, "@"); idx > 0 {
		return s[:idx], s[idx+1:]
	}
	return s, ""
}
