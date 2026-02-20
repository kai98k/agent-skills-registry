package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/liuyukai/agentskills-cli/internal/provider"
)

var providerFlag string

var rootCmd = &cobra.Command{
	Use:   "agentskills",
	Short: "AgentSkills CLI — publish and pull AI Agent Skill bundles",
	Long: `AgentSkills is a CLI tool for managing AI Agent Skills.
It supports publishing, pulling, searching, and initializing skill bundles
for various AI agent providers (Claude, Gemini, Codex, Copilot, Cursor, Windsurf, Antigravity).`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&providerFlag, "provider", "",
		"Target agent provider: claude, gemini, codex, copilot, cursor, windsurf, antigravity, generic (auto-detected if omitted)")
}

// resolveProvider determines the provider from flag, auto-detection, or config fallback
func resolveProvider(dir string) provider.Provider {
	if providerFlag != "" {
		if !provider.IsValidProvider(providerFlag) {
			fmt.Fprintf(os.Stderr, "Warning: unknown provider '%s', using 'generic'\n", providerFlag)
			return provider.Generic
		}
		return provider.Provider(providerFlag)
	}

	// Auto-detect
	result := provider.Detect(dir)
	if result.Provider != provider.Generic {
		fmt.Fprintf(os.Stderr, "Detected provider: %s (%s)\n", result.Provider, joinIndicators(result.Indicators))
		fmt.Fprintln(os.Stderr, "Use --provider to override.")
	}
	return result.Provider
}

func joinIndicators(indicators []string) string {
	if len(indicators) == 0 {
		return "no indicators"
	}
	result := indicators[0]
	for i := 1; i < len(indicators); i++ {
		result += ", " + indicators[i]
	}
	return result
}
