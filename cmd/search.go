package cmd

import (
	"fmt"
	"strings"

	"github.com/liuyukai/agentskills/internal/api"
	"github.com/liuyukai/agentskills/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search for Skills",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword := args[0]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		client := api.New(cfg.APIURL, cfg.Token)

		result, err := client.Search(keyword)
		if err != nil {
			return err
		}

		results, _ := result["results"].([]interface{})
		if len(results) == 0 {
			fmt.Println("No skills found.")
			return nil
		}

		// Print header
		fmt.Printf("%-25s %-10s %-12s %s\n", "NAME", "VERSION", "DOWNLOADS", "DESCRIPTION")

		for _, r := range results {
			m, ok := r.(map[string]interface{})
			if !ok {
				continue
			}
			name, _ := m["name"].(string)
			version, _ := m["latest_version"].(string)
			desc, _ := m["description"].(string)
			downloads, _ := m["downloads"].(float64)

			// Truncate description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}

			fmt.Printf("%-25s %-10s %-12s %s\n",
				truncate(name, 25),
				truncate(version, 10),
				fmt.Sprintf("%.0f", downloads),
				desc,
			)
		}
		return nil
	},
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s + strings.Repeat(" ", max-len(s))
	}
	return s[:max-3] + "..."
}
