package cmd

import (
	"fmt"
	"os"
	"strings"

	filepickerInput "journal/cmd/ui"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

type ViewModel struct {
	body       string
	err        error
	filepicker filepicker.Model
}

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "",
}

func (m ViewModel) Init() tea.Cmd {
	return nil
}

func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

func (m ViewModel) View() string {
	doc := strings.Builder{}

	out, err := glamour.Render("# Hi", "dark")
	cobra.CheckErr(err)

	doc.WriteString(out)
	doc.WriteString(m.filepicker.View())

	return docStyle.Render(doc.String())
}

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Run = func(cmd *cobra.Command, args []string) {
		lipgloss.SetHasDarkBackground(termenv.HasDarkBackground())
		fpModel := filepickerInput.InitialFilePickerModel("")

		m := ViewModel{
			body:       "",
			filepicker: fpModel.Filepicker,
		}

		if _, err := tea.NewProgram(m).Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}

}
