package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/liuyukai/agentskills/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Configure API URL and token",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("Enter API URL [%s]: ", cfg.APIURL)
		urlInput, _ := reader.ReadString('\n')
		urlInput = strings.TrimSpace(urlInput)
		if urlInput != "" {
			cfg.APIURL = urlInput
		}

		fmt.Print("Enter API token: ")
		tokenInput, _ := reader.ReadString('\n')
		tokenInput = strings.TrimSpace(tokenInput)
		if tokenInput != "" {
			cfg.Token = tokenInput
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("save config: %w", err)
		}

		fmt.Printf("Token saved to %s\n", config.DefaultPath())
		return nil
	},
}
