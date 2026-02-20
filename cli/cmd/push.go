package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/liuyukai/agentskills-cli/internal/api"
	"github.com/liuyukai/agentskills-cli/internal/bundle"
	"github.com/liuyukai/agentskills-cli/internal/parser"
	"github.com/liuyukai/agentskills-cli/internal/provider"
)

var pushCmd = &cobra.Command{
	Use:   "push [path]",
	Short: "Pack and upload a Skill Bundle",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		// Read and validate SKILL.md locally
		skillMDPath := filepath.Join(absPath, "SKILL.md")
		content, err := os.ReadFile(skillMDPath)
		if err != nil {
			return fmt.Errorf("reading SKILL.md: %w", err)
		}

		meta, _, err := parser.ParseSkillMD(string(content))
		if err != nil {
			return fmt.Errorf("validating SKILL.md: %w", err)
		}
		fmt.Println("Validating SKILL.md...        OK")

		// Resolve provider
		p := resolveProvider(absPath)

		// Provider-specific name validation
		if err := provider.ValidateName(p, meta.Name); err != nil {
			return err
		}

		providerStr := string(p)
		if p == provider.Generic {
			providerStr = ""
		}
		if p != provider.Generic {
			fmt.Printf("Provider: %s\n", p)
		}

		// Pack the bundle
		bundleData, err := bundle.Pack(absPath)
		if err != nil {
			return err
		}
		sizeKB := float64(len(bundleData)) / 1024.0
		fmt.Printf("Packing bundle...             OK (%.1f KB)\n", sizeKB)

		// Compute local checksum
		localChecksum := bundle.Checksum(bundleData)

		// Upload
		client, err := api.NewClient()
		if err != nil {
			return fmt.Errorf("creating API client: %w", err)
		}

		fmt.Printf("Uploading %s@%s...   ", meta.Name, meta.Version)
		result, err := client.Publish(bundleData, fmt.Sprintf("%s-%s.tar.gz", meta.Name, meta.Version), providerStr)
		if err != nil {
			fmt.Println("FAILED")
			return err
		}
		fmt.Println("OK")

		// Verify checksum
		serverChecksum := strings.TrimPrefix(result.Checksum, "sha256:")
		if serverChecksum != localChecksum {
			return fmt.Errorf("checksum mismatch: local=%s server=%s", localChecksum, serverChecksum)
		}
		fmt.Printf("Checksum: sha256:%s\n", localChecksum)

		fmt.Printf("\nPublished %s@%s successfully.\n", result.Name, result.Version)
		if len(result.Providers) > 0 {
			fmt.Printf("  Providers: %s\n", strings.Join(result.Providers, ", "))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
