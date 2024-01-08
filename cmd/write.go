package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

type WriteModel struct {
	textarea textarea.Model
	body     string
	err      error
	date     string
	msg      string
}

type ClearSuccessMsg struct{}

func clearSuccessMsgAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return ClearSuccessMsg{}
	})
}

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Write a new entry",
}

func (m WriteModel) Init() tea.Cmd {
	return m.textarea.Focus()
}

func (m WriteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		case tea.KeyCtrlS:
			m.body = m.textarea.Value()
			date := time.Now().Format("January-02-2006")
			f, err := os.Create("./" + date + ".md")
			defer f.Close()
			if err != nil {
				panic(err)
			}
			_, err = f.WriteString(m.body)
			if err != nil {
				panic(err)
			}
			m.msg = "Entry saved"

			return m, tea.Batch(cmd, clearSuccessMsgAfter(2*time.Second))
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}
	case error:
		m.err = msg
		return m, nil
	case ClearSuccessMsg:
		m.msg = ""
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m WriteModel) View() string {
	doc := strings.Builder{}

	highlight := lipgloss.NewStyle().Foreground(primaryColor)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	success := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
	checkMark := lipgloss.NewStyle().SetString("✓").
		Foreground(lipgloss.Color("#73F59F")).
		PaddingRight(1).
		String()

	var successMsg string
	if len(m.msg) > 0 {
		successMsg = checkMark + success.Render(m.msg)
	}

	column := lipgloss.JoinVertical(lipgloss.Left,
		highlight.Render(m.date)+"\n",
		m.textarea.View(),
		successMsg,
		helpStyle.Render("ctrl+s: save • ctrl+c: quit"))

	doc.WriteString(column)

	return docStyle.Render(doc.String())
}

func init() {
	rootCmd.AddCommand(writeCmd)
	writeCmd.Run = func(cmd *cobra.Command, args []string) {
		lipgloss.SetHasDarkBackground(termenv.HasDarkBackground())

		ta := textarea.New()
		ta.Placeholder = "How was your today?"
		ta.CharLimit = 0
		ta.SetWidth(50)
		ta.SetHeight(15)
		ta.Focus()

		m := WriteModel{
			textarea: ta,
			body:     "",
			date:     time.Now().Format("January 02 2006"),
		}

		if _, err := tea.NewProgram(m).Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}
}
