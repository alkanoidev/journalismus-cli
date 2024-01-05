package filepickerInput

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	Filepicker   filepicker.Model
	SelectedFile string
	quitting     bool
	err          error
}

func InitialFilePickerModel(selectedFile string) model {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".mod", ".sum", ".go", ".txt", ".md"}
	ex, _ := os.Executable()
	fp.CurrentDirectory = filepath.Dir(ex)

	return model{
		Filepicker:   fp,
		SelectedFile: selectedFile,
		quitting:     false,
		err:          nil,
	}
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (m model) Init() tea.Cmd {
	return m.Filepicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	m.Filepicker, cmd = m.Filepicker.Update(msg)

	// Did the user select a file?
	if didSelect, path := m.Filepicker.DidSelectFile(msg); didSelect {
		// Get the path of the selected file.
		m.SelectedFile = path
	}

	// Did the user select a disabled file?
	// This is only necessary to display an error to the user.
	if didSelect, path := m.Filepicker.DidSelectDisabledFile(msg); didSelect {
		// Let's clear the selectedFile and display an error.
		m.err = errors.New(path + " is not valid.")
		m.SelectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	var s strings.Builder
	s.WriteString("\n  ")
	if m.err != nil {
		s.WriteString(m.Filepicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if m.SelectedFile == "" {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected file: " + m.Filepicker.Styles.Selected.Render(m.SelectedFile))
	}
	s.WriteString("\n\n" + m.Filepicker.View() + "\n")
	return s.String()
}
