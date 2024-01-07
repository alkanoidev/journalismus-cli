package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

var viewTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func init() {
	viewCmd.AddCommand(viewTodayCmd)
	viewTodayCmd.Run = func(cmd *cobra.Command, args []string) {
		lipgloss.SetHasDarkBackground(termenv.HasDarkBackground())

		path := time.Now().Format("January-02-2006") + ".md"
		f, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		r, _ := glamour.NewTermRenderer(
			glamour.WithStylesFromJSONFile("glamour_theme.json"),
		)
		out, _ := r.Render(string(f))
		fmt.Print(out)
	}
}
