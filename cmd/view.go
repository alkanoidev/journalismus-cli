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
		f, err := os.Open(m.selectedFile)
		cobra.CheckErr(err)
		defer f.Close()
		b1 := make([]byte, 255)
		_, err = f.Read(b1)
		cobra.CheckErr(err)

		m.body = string(b1)
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
	var s strings.Builder
	s.WriteString("\n  ")
	if m.err != nil {
		s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if m.selectedFile == "" {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected file: " + m.filepicker.Styles.Selected.Render(m.selectedFile))
	}
	s.WriteString("\n\n" + m.filepicker.View() + "\n")

	if len(m.body) > 0 {
		out, _ := glamour.Render(m.body, "dark")
		outputStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#3d8a96"))
		s.WriteString(outputStyle.Render(out))
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
		fp.Styles.Selected.Foreground(lipgloss.Color("#72cedd"))
		fp.Styles.Cursor.Foreground(lipgloss.Color("#3d8a96"))
		fp.Styles.Directory.Foreground(lipgloss.Color("#3d8a96"))
		fp.ShowHidden = false
		fp.Styles.FileSize.Width(0)
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
