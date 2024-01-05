/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

type RootModel struct{}

var rootCmd = &cobra.Command{
	Use: "journal",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func (m RootModel) Init() tea.Cmd {
	return nil
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
		// TODO: onaj event za resize https://youtu.be/Gl31diSVP8M?si=qyI3layE6KGR7A36&t=328
	}

	return m, nil
}

var (
	welcomeMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#72cedd")).
				Align(lipgloss.Center).
				BorderStyle(lipgloss.RoundedBorder())
	docStyle = lipgloss.NewStyle().
			Padding(1)
)

func (m RootModel) View() string {
	doc := strings.Builder{}

	msg := `                 Welcome to                     
			▀▀█ █▀█ █ █ █▀▄ █▀█ █▀█ █   ▀█▀ █▀▀ █▄█ █ █ █▀▀
			  █ █ █ █ █ █▀▄ █ █ █▀█ █    █  ▀▀█ █ █ █ █ ▀▀█
			▀▀  ▀▀▀ ▀▀▀ ▀ ▀ ▀ ▀ ▀ ▀ ▀▀▀ ▀▀▀ ▀▀▀ ▀ ▀ ▀▀▀ ▀▀▀
			          Capture thoughts effortlessly          
			             in the command line.               `
	doc.WriteString(welcomeMessageStyle.Render(strings.ReplaceAll(msg, "\t", "")))

	return docStyle.Render(doc.String())
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Run = func(cmd *cobra.Command, args []string) {

		lipgloss.SetHasDarkBackground(termenv.HasDarkBackground())

		m := RootModel{}

		if _, err := tea.NewProgram(m).Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}
}
