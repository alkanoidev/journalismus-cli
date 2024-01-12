package cmd

import (
	"fmt"
	entryUtils "journal/internal/utils"
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
	return textarea.Blink
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
			if len(strings.TrimSpace(m.body)) == 0 {
				err := entryUtils.DeleteEntry()
				if err != nil {
					m.err = err
					os.Exit(1)
				}
			} else {
				m.msg, m.err = entryUtils.WriteEntryToFile(m.body)
				clearSuccessMsgAfter(2 * time.Second)
			}
			return m, tea.Quit
		case tea.KeyEnter:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		case tea.KeyCtrlS:
			m.body = m.textarea.Value()
			m.msg, m.err = entryUtils.WriteEntryToFile(m.body)
			return m, tea.Batch(cmd, clearSuccessMsgAfter(2*time.Second))
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
			m.body = m.textarea.Value()
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
	s := strings.Builder{}

	var successMsg string
	if len(m.msg) > 0 {
		successMsg = checkMark + success.Render(m.msg)
	}

	column := lipgloss.JoinVertical(lipgloss.Left,
		highlight.Render(m.date)+"\n",
		m.textarea.View(),
		successMsg,
		helpStyle.Render("ctrl+s: save • ctrl+c: quit"))

	s.WriteString(column)

	return docStyle.Render(s.String())
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

		entry, err := entryUtils.ReadOrCreateEntry()
		cobra.CheckErr(err)
		ta.SetValue(entry)

		m := WriteModel{
			textarea: ta,
			body:     ta.Value(),
			date:     time.Now().Format("January 02 2006"),
		}

		if _, err := tea.NewProgram(m).Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}
}
