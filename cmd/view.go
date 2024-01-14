package cmd

import (
	"fmt"
	"io"
	entryUtils "journal/internal/utils"
	"os"
	"path"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

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
	paperInfoStyle    = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return paperTitleStyle.Copy().BorderStyle(b).BorderForeground(primaryColor)
	}()
	paperTitleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1).BorderForeground(primaryColor)
	}()
	paperLineStyle = lipgloss.NewStyle().Foreground(primaryColor).Render
)

type ViewModel struct {
	body         string
	list         list.Model
	selectedFile string
	quitting     bool
	err          error
	entryPicker  bool
	paper        viewport.Model
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
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
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
		case "tab":
			m.entryPicker = !m.entryPicker
		}
	// case tea.WindowSizeMsg:
	// 	headerHeight := lipgloss.Height(m.headerView())
	// 	footerHeight := lipgloss.Height(m.footerView())
	// 	verticalMarginHeight := headerHeight + footerHeight

	// 	if !m.ready {
	// 		m.paper = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
	// 		m.paper.YPosition = headerHeight
	// 		m.paper.SetContent(m.bodyRendered)
	// 		m.ready = true
	// 		m.paper.YPosition = headerHeight + 1
	// 	} else {
	// 		m.paper.Width = msg.Width
	// 		m.paper.Height = msg.Height - verticalMarginHeight
	// 	}
	// 	return m, tea.Batch(cmds...)
	case clearErrorMsg:
		m.err = nil
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
		s.WriteString("\n\n" + m.headerView() + "\n" + m.paper.View() + m.footerView())
	}

	return docStyle.Render(s.String())
}

func (m ViewModel) headerView() string {
	title := paperTitleStyle.Render(m.selectedFile)
	line := strings.Repeat("─", max(0, m.paper.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, paperLineStyle(line))
}

func (m ViewModel) footerView() string {
	info := paperInfoStyle.Render(fmt.Sprintf("%3.f%%", m.paper.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.paper.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, paperLineStyle(line), info)
}

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Run = func(cmd *cobra.Command, args []string) {
		lipgloss.SetHasDarkBackground(termenv.HasDarkBackground())

		ex, _ := os.Executable()
		entries := []list.Item{}
		c, err := os.ReadDir(path.Dir(ex))
		cobra.CheckErr(err)
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

		const width = 78

		paper := viewport.New(width, 20)
		paper.Style = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderTop(false).
			BorderBottom(false).
			BorderForeground(lipgloss.Color("62")).
			PaddingRight(2)

		m := ViewModel{
			list:        l,
			entryPicker: true,
			paper:       paper,
		}

		_, err = tea.NewProgram(&m, tea.WithOutput(os.Stderr)).Run()
		if err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}

}
