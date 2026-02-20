package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/liuyukai/agentskills-cli/internal/config"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Save API token to local config",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		fmt.Print("Enter API token: ")
		reader := bufio.NewReader(os.Stdin)
		token, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading token: %w", err)
		}
		token = strings.TrimSpace(token)

		if token == "" {
			return fmt.Errorf("token cannot be empty")
		}

		cfg.Token = token
		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("Token saved to %s\n", config.ConfigPath())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
