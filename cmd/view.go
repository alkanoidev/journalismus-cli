package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

type ViewModel struct {
	body         string
	filepicker   filepicker.Model
	selectedFile string
	quitting     bool
	err          error
}

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "",
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (m ViewModel) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	case clearErrorMsg:
		m.err = nil
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		m.selectedFile = path

		f, err := os.ReadFile(m.selectedFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		m.body = string(f)
	}

	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		m.err = errors.New(path + " is not valid.")
		m.selectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return m, cmd
}

func (m ViewModel) View() string {
	if m.quitting {
		return ""
	}
	s := strings.Builder{}
	if m.err != nil {
		s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
	}

	if len(m.body) > 0 {
		r, _ := glamour.NewTermRenderer(
			glamour.WithStylesFromJSONFile("glamour_theme.json"),
		)
		out, _ := r.Render(m.body)
		s.WriteString(out)
	} else {
		s.WriteString("Pick a entry:")
		s.WriteString("\n\n" + m.filepicker.View() + "\n")
	}

	return s.String()
}

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Run = func(cmd *cobra.Command, args []string) {
		lipgloss.SetHasDarkBackground(termenv.HasDarkBackground())

		fp := filepicker.New()
		fp.AllowedTypes = []string{".md"}
		ex, _ := os.Executable()
		fp.CurrentDirectory = path.Dir(ex)
		fp.ShowHidden = false
		fp.Styles.Selected.Foreground(primaryColor)
		fp.Styles.Cursor.Foreground(primaryColor)

		m := ViewModel{
			filepicker: fp,
		}

		_, err := tea.NewProgram(&m, tea.WithOutput(os.Stderr)).Run()
		if err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}

}
