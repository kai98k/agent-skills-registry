package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/liuyukai/agentskills-cli/internal/api"
)

var searchTag string

var searchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search for Skills on the registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword := args[0]

		client, err := api.NewClient()
		if err != nil {
			return fmt.Errorf("creating API client: %w", err)
		}

		result, err := client.Search(keyword, searchTag, providerFlag, 1, 20)
		if err != nil {
			return err
		}

		if len(result.Results) == 0 {
			fmt.Println("No skills found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tVERSION\tDOWNLOADS\tPROVIDERS\tDESCRIPTION")
		for _, r := range result.Results {
			providers := strings.Join(r.Providers, ",")
			desc := r.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
				r.Name, r.LatestVersion, r.Downloads, providers, desc)
		}
		w.Flush()

		return nil
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchTag, "tag", "", "Filter by tag")
	rootCmd.AddCommand(searchCmd)
}
