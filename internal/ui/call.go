package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/MqllR/abitool/pkg/abiparser"
)

// callFormScreen is a TUI screen that collects an RPC URL and function arguments
// before executing a read-only eth_call. It is pushed onto the screen stack from
// the browse screen when the user presses Enter on a view/pure function.
//
// Layout: first field = RPC URL (pre-filled), subsequent fields = one per input param.
type callFormScreen struct {
	address string
	element abiparser.Element

	labels []string
	inputs []textinput.Model
	cursor int

	width  int
	height int
}

func newCallFormScreen(address string, el abiparser.Element, rpcURL string) callFormScreen {
	// Build one label+input per field: [rpc url, param0, param1, ...]
	count := 1 + len(el.Inputs)
	labels := make([]string, count)
	inputs := make([]textinput.Model, count)

	// RPC URL field
	labels[0] = "RPC URL"
	rpcInput := textinput.New()
	rpcInput.Placeholder = "https://..."
	rpcInput.SetValue(rpcURL)
	rpcInput.CharLimit = 256
	rpcInput.Width = 60
	inputs[0] = rpcInput

	// One field per function input
	for i, inp := range el.Inputs {
		labels[i+1] = fmt.Sprintf("%s (%s)", inp.Name, inp.Type)
		ti := textinput.New()
		ti.Placeholder = inp.Type
		ti.CharLimit = 256
		ti.Width = 60
		inputs[i+1] = ti
	}

	inputs[0].Focus()

	return callFormScreen{
		address: address,
		element: el,
		labels:  labels,
		inputs:  inputs,
	}
}

func (m callFormScreen) setSize(w, h int) screen { m.width, m.height = w, h; return m }

func (m callFormScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m callFormScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			return m, func() tea.Msg { return popMsg{} }

		case tea.KeyEnter, tea.KeyTab:
			if m.cursor < len(m.inputs)-1 {
				m.inputs[m.cursor].Blur()
				m.cursor++
				m.inputs[m.cursor].Focus()
				return m, textinput.Blink
			}
			// Last field — submit: push the result screen which will fire the call.
			rpcURL := m.inputs[0].Value()
			args := make([]string, len(m.inputs)-1)
			for i, inp := range m.inputs[1:] {
				args[i] = inp.Value()
			}
			next := newCallResultScreen(m.address, m.element, rpcURL, args)
			return m, func() tea.Msg { return pushMsg{next} }

		case tea.KeyShiftTab:
			if m.cursor > 0 {
				m.inputs[m.cursor].Blur()
				m.cursor--
				m.inputs[m.cursor].Focus()
				return m, textinput.Blink
			}
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.inputs[m.cursor].Blur()
				m.cursor--
				m.inputs[m.cursor].Focus()
				return m, textinput.Blink
			}
		}
	}

	var cmd tea.Cmd
	m.inputs[m.cursor], cmd = m.inputs[m.cursor].Update(msg)
	return m, cmd
}

func (m callFormScreen) View() string {
	w, h := m.width, m.height
	if w < 10 {
		w = 80
	}
	if h < 5 {
		h = 24
	}

	fnName := m.element.Name
	if fnName == "" {
		fnName = string(m.element.Type)
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("  Call: "+fnName) + "\n")
	sb.WriteString(dimStyle.Render("  "+strings.Repeat("─", max(w-10, 10))) + "\n\n")

	for i, inp := range m.inputs {
		prefix := "  "
		if i == m.cursor {
			prefix = "▸ "
		}
		sb.WriteString(dimStyle.Render(prefix+m.labels[i]) + "\n")
		sb.WriteString("  " + inp.View() + "\n\n")
	}

	sb.WriteString(dimStyle.Render("  tab: next  shift-tab: prev  enter: submit  esc: back"))

	boxW := w - 10
	if boxW < 50 {
		boxW = 50
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Width(boxW).
		Render(sb.String())

	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}
