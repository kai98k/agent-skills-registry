package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	appVersion = "dev"
	appBuild   = "unknown"
)

func SetVersionInfo(version, buildTime string) {
	appVersion = version
	appBuild = buildTime
}

var rootCmd = &cobra.Command{
	Use:   "agentskills",
	Short: "AgentSkills — AI Agent Skill Registry CLI",
	Long:  "A CLI tool for publishing, downloading, and managing AI Agent Skills.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
