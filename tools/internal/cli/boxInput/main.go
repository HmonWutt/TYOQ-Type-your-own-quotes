package main

// A simple program demonstrating the text area component from the Bubbles
// component library.

import (
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
}

type helloMsg string

func waitASec() tea.Msg {
	time.Sleep(time.Second)
	return helloMsg("Hi, there!")
}

type model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
}

func initialModel() model {
	ta := textarea.New()
	// ta.Placeholder = "Send a message..."
	// ta.SetVirtualCursor(false)
	// ta.Focus()

	// ta.Prompt = "┃ "
	ta.CharLimit = 28000

	ta.SetWidth(0)
	ta.SetHeight(0)

	// Remove cursor line styling
	s := ta.Styles()
	s.Focused.CursorLine = lipgloss.NewStyle()
	ta.SetStyles(s)

	ta.ShowLineNumbers = false
	vp := viewport.New(viewport.WithWidth(30), viewport.WithHeight(5))
	vp.SetContent(`Welcome to TYOQ.`)
	vp.KeyMap.Left.SetEnabled(false)
	vp.KeyMap.Right.SetEnabled(false)

	ta.KeyMap.InsertNewline.SetEnabled(true)

	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.textarea.SetWidth(msg.Width)
		// m.textarea.SetWidth(msg.Width)
		m.textarea.SetHeight(msg.Height)

		if len(m.messages) > 0 {
			// Wrap content before setting it.
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width()).Render(strings.Join(m.messages, "\n")))
		}
		m.viewport.GotoBottom()
	case tea.KeyPressMsg:
		// fmt.Println(msg.String())
		switch msg.String() {
		case "ctrl+c", "esc":
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case "enter":
			m.messages = append(m.messages, "\n")
			// fmt.Println(m.messages)
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width()).Render(strings.Join(m.messages, "")))
			m.textarea.Reset()
			m.viewport.GotoBottom()
			// return m, nil
			// Send all other keypresses to the textarea.
			// var cmd tea.Cmd

			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		case "space":
			m.textarea, cmd = m.textarea.Update(msg)
			m.messages = append(m.messages, " ")
			// fmt.Println(m.messages)
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width()).Render(strings.Join(m.messages, "")))
			m.textarea.Reset()
			m.viewport.GotoBottom()
			// return m, nil
			// Send all other keypresses to the textarea.
			// var cmd tea.Cmd
			return m, cmd

		case "backspace":
			m.textarea, cmd = m.textarea.Update(msg)
			m.messages = m.messages[:len(m.messages)-1]
			// fmt.Println(m.messages)
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width()).Render(strings.Join(m.messages, "")))
			m.textarea.Reset()
			m.viewport.GotoBottom()
			// return m, nil
			// Send all other keypresses to the textarea.
			// var cmd tea.Cmd
			return m, cmd
		default:

			m.textarea, cmd = m.textarea.Update(msg)
			m.messages = append(m.messages, m.senderStyle.Render(msg.String()))
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width()).Render(strings.Join(m.messages, "")))
			m.textarea.Reset()
			m.viewport.GotoBottom()
			// return m, nil
			// Send all other keypresses to the textarea.
			return m, cmd
		}

	case cursor.BlinkMsg:
		// Textarea should also process cursor blinks.
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() tea.View {
	viewportView := m.viewport.View()
	v := tea.NewView(viewportView + "|")
	c := v.Cursor
	if c != nil {
		c.Y += lipgloss.Height(viewportView)
	}
	v.Cursor = c
	v.AltScreen = true
	return v
}
