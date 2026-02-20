package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/liuyukai/agentskills-cli/internal/api"
	"github.com/liuyukai/agentskills-cli/internal/bundle"
	"github.com/liuyukai/agentskills-cli/internal/provider"
)

var pullScope string

var pullCmd = &cobra.Command{
	Use:   "pull <name>[@version]",
	Short: "Download and extract a Skill Bundle",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nameVersion := args[0]
		name, version := parseNameVersion(nameVersion)

		client, err := api.NewClient()
		if err != nil {
			return fmt.Errorf("creating API client: %w", err)
		}

		// If no version specified, get latest
		if version == "" {
			info, err := client.GetSkill(name)
			if err != nil {
				return err
			}
			if info.LatestVersion == nil {
				return fmt.Errorf("skill '%s' has no published versions", name)
			}
			version = info.LatestVersion.Version
		}

		// Download
		fmt.Printf("Downloading %s@%s...  ", name, version)
		data, serverChecksum, err := client.Download(name, version)
		if err != nil {
			fmt.Println("FAILED")
			return err
		}
		fmt.Println("OK")

		// Verify checksum
		localChecksum := bundle.Checksum(data)
		if serverChecksum != "" && serverChecksum != localChecksum {
			return fmt.Errorf("checksum mismatch: server=%s local=%s", serverChecksum, localChecksum)
		}
		fmt.Println("Verifying checksum...          OK")

		// Determine extraction path
		cwd, _ := os.Getwd()
		p := resolveProvider(cwd)

		var targetDir string
		if pullScope == "user" {
			targetDir = provider.UserInstallPath(p, name)
		} else {
			targetDir = provider.WorkspaceInstallPath(p, name, cwd)
		}

		// Create parent directory
		if err := os.MkdirAll(targetDir, 0o755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}

		// Extract
		if err := bundle.Unpack(data, targetDir); err != nil {
			return fmt.Errorf("extracting bundle: %w", err)
		}

		// Make path relative for display
		relPath := targetDir
		if rel, err := relativeDisplay(cwd, targetDir); err == nil {
			relPath = rel
		}

		if p != provider.Generic {
			fmt.Printf("Provider: %s\n", p)
		}
		fmt.Printf("Extracted to %s/\n", relPath)

		return nil
	},
}

func parseNameVersion(s string) (string, string) {
	parts := strings.SplitN(s, "@", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

func relativeDisplay(base, target string) (string, error) {
	rel, err := os.Getwd()
	if err != nil {
		return target, err
	}
	_ = rel
	// Try to make relative
	if strings.HasPrefix(target, base) {
		return "." + target[len(base):], nil
	}
	return target, nil
}

func init() {
	pullCmd.Flags().StringVar(&pullScope, "scope", "workspace",
		"Install scope: workspace (project-level) or user (user-level)")
	rootCmd.AddCommand(pullCmd)
}
