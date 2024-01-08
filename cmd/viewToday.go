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
	Short: "",
	Long:  ``,
}

func init() {
	viewCmd.AddCommand(viewTodayCmd)
	viewTodayCmd.Run = func(cmd *cobra.Command, args []string) {
		lipgloss.SetHasDarkBackground(termenv.HasDarkBackground())
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

		path := time.Now().Format("January-02-2006") + ".md"
		f, err := os.ReadFile(path)
		if err != nil {
			fmt.Print(helpStyle.Render(" No entry for today detected...\n"))
			writeCmd.Run(cmd, args)

		}

		r, _ := glamour.NewTermRenderer(
			glamour.WithStylesFromJSONFile("theme.json"),
		)
		out, _ := r.Render(string(f))

		fmt.Print(docStyle.Render(out))
	}
}
