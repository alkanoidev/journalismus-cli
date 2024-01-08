package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

type item string

func (i item) FilterValue() string { return string(i) }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

var (
	titleStyle        = lipgloss.NewStyle().Padding(0)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(1)
	selectedItemStyle = lipgloss.NewStyle().Foreground(primaryColor).PaddingLeft(1)
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(1)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(0)
)

type ViewModel struct {
	body         string
	filepicker   filepicker.Model
	list         list.Model
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
	return nil
}

func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case clearErrorMsg:
		m.err = nil
	}

	var cmd tea.Cmd
	// m.filepicker, cmd = m.filepicker.Update(msg)
	m.list, cmd = m.list.Update(msg)

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
			glamour.WithStylesFromJSONFile("theme.json"),
		)
		out, _ := r.Render(m.body)
		s.WriteString(out)
	} else {
		// s.WriteString("\n\n" + m.filepicker.View() + "\n")
		s.WriteString(m.list.View())
	}

	return docStyle.Render(s.String())
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

		entries := []list.Item{}
		c, err := os.ReadDir(path.Dir(ex))
		cobra.CheckErr(err)
		for _, entry := range c {
			if path.Ext(entry.Name()) == ".md" {
				entries = append(entries, item(entry.Name()))
			}
		}

		const defaultWidth = 20

		l := list.New(entries, itemDelegate{}, defaultWidth, 14)
		l.Title = "Pick a entry:"
		l.SetShowStatusBar(false)
		l.Styles.Title = titleStyle
		l.Styles.PaginationStyle = paginationStyle
		l.Styles.HelpStyle = helpStyle

		m := ViewModel{
			filepicker: fp,
			list:       l,
		}

		_, err = tea.NewProgram(&m, tea.WithOutput(os.Stderr)).Run()
		if err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}

}
