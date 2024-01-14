package cmd

import (
	"fmt"
	"io"
	entryUtils "journal/internal/utils"
	"os"
	"path"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "",
}

type item string

func (i item) FilterValue() string { return string(i) }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
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
	listHelpStyle     = list.DefaultStyles().HelpStyle.PaddingLeft(0)
)

type ViewModel struct {
	body         string
	list         list.Model
	selectedFile string
	quitting     bool
	err          error
	entryPicker  bool
	paper        viewport.Model
	ready        bool
}

func (m ViewModel) Init() tea.Cmd {
	return nil
}

func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape, tea.KeyTab:
			if !m.entryPicker {
				m.entryPicker = true
			}
			return m, nil

		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEnter:
			if m.entryPicker {
				entry, ok := m.list.SelectedItem().(item)
				m.entryPicker = false
				if ok {
					m.selectedFile = string(entry)

					body, err := entryUtils.ReadFile(m.selectedFile)
					if err != nil {
						fmt.Fprintln(os.Stderr, "Failed opening file: ", err)
						os.Exit(1)
					}
					m.body = string(body)

					r, _ := glamour.NewTermRenderer(
						glamour.WithStylesFromJSONFile("theme.json"),
					)
					out, _ := r.Render(m.body)

					m.paper.SetContent(string(out))
				}
			}
		}
		// case tea.WindowSizeMsg:
		// 	headerHeight := lipgloss.Height(m.headerView())
		// 	footerHeight := lipgloss.Height(m.footerView())
		// 	verticalMarginHeight := headerHeight + footerHeight

		// 	if !m.ready {
		// 		m.paper = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
		// 		m.paper.YPosition = headerHeight
		// 		m.ready = true
		// 		m.paper.YPosition = headerHeight + 1
		// 	} else {
		// 		m.paper.Width = msg.Width
		// 		m.paper.Height = msg.Height - verticalMarginHeight
		// 	}
	}
	var cmd tea.Cmd
	if m.entryPicker {
		m.list, cmd = m.list.Update(msg)
	} else {
		m.paper, cmd = m.paper.Update(msg)
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m ViewModel) View() string {
	if m.quitting {
		return ""
	}
	s := strings.Builder{}
	if m.entryPicker {
		s.WriteString(m.list.View())
	} else {
		s.WriteString(lipgloss.JoinVertical(lipgloss.Left, m.paper.View(), helpStyle.Render("tab: toggle list of entries • ctrl+c: quit")))
	}

	return s.String()
}

func newModel() (*ViewModel, error) {
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}
	entries := []list.Item{}
	c, err := os.ReadDir(path.Dir(ex))
	if err != nil {
		return nil, err
	}
	for _, entry := range c {
		if path.Ext(entry.Name()) == ".md" {
			entries = append(entries, item(entry.Name()))
		}
	}

	l := list.New(entries, itemDelegate{}, 20, 14)
	l.Title = "Pick a entry:"
	l.SetShowStatusBar(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = listHelpStyle

	paper := viewport.New(78, 20)
	paper.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor)

	return &ViewModel{
		list:        l,
		entryPicker: true,
		paper:       paper,
	}, nil
}

func init() {
	rootCmd.AddCommand(viewCmd)

	m, err := newModel()
	if err != nil {
		fmt.Println("Could not initialize Bubble Tea model:", err)
		os.Exit(1)
	}

	viewCmd.Run = func(cmd *cobra.Command, args []string) {
		lipgloss.SetHasDarkBackground(termenv.HasDarkBackground())

		_, err := tea.NewProgram(m, tea.WithOutput(os.Stderr)).Run()
		if err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}
}
