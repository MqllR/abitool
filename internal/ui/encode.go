// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package ui

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/MqllR/abitool/pkg/abicodec"
	"github.com/MqllR/abitool/pkg/abiparser"
)

// ─── encodeFormScreen ─────────────────────────────────────────────────────────

// encodeFormScreen collects function arguments before generating calldata.
// It mirrors callFormScreen but without the RPC URL field.
// If the function has no inputs, it skips the form and goes directly to the
// result screen via Init().
type encodeFormScreen struct {
	address string
	element abiparser.Element

	labels []string
	inputs []textinput.Model
	cursor int

	width  int
	height int
}

func newEncodeFormScreen(address string, el abiparser.Element) encodeFormScreen {
	count := len(el.Inputs)
	labels := make([]string, count)
	inputs := make([]textinput.Model, count)

	for i, inp := range el.Inputs {
		labels[i] = fmt.Sprintf("%s (%s)", inp.Name, inp.Type)
		ti := textinput.New()
		ti.Placeholder = inp.Type
		ti.CharLimit = 256
		ti.Width = 60
		inputs[i] = ti
	}

	if len(inputs) > 0 {
		inputs[0].Focus()
	}

	return encodeFormScreen{
		address: address,
		element: el,
		labels:  labels,
		inputs:  inputs,
	}
}

func (m encodeFormScreen) setSize(w, h int) screen { m.width, m.height = w, h; return m }

func (m encodeFormScreen) Init() tea.Cmd {
	// Zero-input functions: skip form, go straight to result.
	if len(m.inputs) == 0 {
		next := newEncodeResultScreen(m.address, m.element, nil)
		return func() tea.Msg { return pushMsg{next} }
	}
	return textinput.Blink
}

func (m encodeFormScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			// Last field — collect args and push result screen.
			args := make([]string, len(m.inputs))
			for i, inp := range m.inputs {
				args[i] = inp.Value()
			}
			next := newEncodeResultScreen(m.address, m.element, args)
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

func (m encodeFormScreen) View() string {
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
	sb.WriteString(titleStyle.Render("  Encode: "+fnName) + "\n")
	sb.WriteString(dimStyle.Render("  "+strings.Repeat("─", max(w-10, 10))) + "\n\n")

	for i, inp := range m.inputs {
		prefix := "  "
		if i == m.cursor {
			prefix = "▸ "
		}
		sb.WriteString(dimStyle.Render(prefix+m.labels[i]) + "\n")
		sb.WriteString("  " + inp.View() + "\n\n")
	}

	sb.WriteString(dimStyle.Render("  tab: next  shift-tab: prev  enter: encode  esc: back"))

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

// ─── encodeResultScreen ───────────────────────────────────────────────────────

// encodeResultScreen encodes the call and displays the calldata hex.
// Encoding is CPU-only, so it is performed synchronously inside Init() and
// returned immediately as an encodeDoneMsg — no loading spinner needed.
type encodeResultScreen struct {
	address string
	element abiparser.Element
	args    []string

	calldata string
	err      error
	loaded   bool

	width  int
	height int
}

// encodeDoneMsg carries the encode result back to the screen.
type encodeDoneMsg struct {
	calldata string
	err      error
}

func newEncodeResultScreen(address string, el abiparser.Element, args []string) encodeResultScreen {
	return encodeResultScreen{
		address: address,
		element: el,
		args:    args,
	}
}

func (m encodeResultScreen) setSize(w, h int) screen { m.width, m.height = w, h; return m }

func (m encodeResultScreen) Init() tea.Cmd {
	return encodeCalldataCmd(m.element, m.args)
}

func encodeCalldataCmd(el abiparser.Element, args []string) tea.Cmd {
	return func() tea.Msg {
		method, err := abicodec.ParseMethod(el)
		if err != nil {
			return encodeDoneMsg{err: fmt.Errorf("parsing method: %w", err)}
		}

		calldata, err := abicodec.EncodeInput(method, args)
		if err != nil {
			return encodeDoneMsg{err: fmt.Errorf("encoding calldata: %w", err)}
		}

		return encodeDoneMsg{calldata: "0x" + hex.EncodeToString(calldata)}
	}
}

func (m encodeResultScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case encodeDoneMsg:
		m.calldata = msg.calldata
		m.err = msg.err
		m.loaded = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "esc", "backspace":
			return m, func() tea.Msg { return popMsg{} }
		}
	}
	return m, nil
}

func (m encodeResultScreen) View() string {
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

	boxW := w - 10
	if boxW < 50 {
		boxW = 50
	}
	innerW := boxW - 6 // padding(2*2) + border(2)

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("  Encode: "+fnName) + "\n")
	sb.WriteString(dimStyle.Render("  "+strings.Repeat("─", max(innerW-2, 10))) + "\n\n")

	if sig, err := m.element.Signature(); err == nil {
		sb.WriteString(dimStyle.Render("  Signature  ") + lipgloss.NewStyle().Foreground(colorWhite).Render(sig) + "\n")
	}
	if sel, err := m.element.Selector(); err == nil {
		sb.WriteString(dimStyle.Render("  Selector   ") + lipgloss.NewStyle().Foreground(colorBlue).Render(sel) + "\n")
	}

	sb.WriteString("\n")

	switch {
	case !m.loaded:
		sb.WriteString(dimStyle.Render("  Encoding...") + "\n")
	case m.err != nil:
		sb.WriteString(errorStyle.Render("  Error: "+m.err.Error()) + "\n")
	default:
		sb.WriteString(dimStyle.Render("  Calldata") + "\n")
		// Wrap calldata at innerW so it fits in the box and can be selected easily.
		sb.WriteString(wrapHex(m.calldata, innerW) + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("  esc / backspace: back  q: quit"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Width(boxW).
		Render(sb.String())

	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

// wrapHex breaks a long hex string into lines of at most lineW characters,
// indented by two spaces.
func wrapHex(s string, lineW int) string {
	if lineW < 8 {
		lineW = 8
	}
	// Reserve 2 chars for the "  " indent.
	chunkW := lineW - 2
	if chunkW < 4 {
		chunkW = 4
	}

	var lines []string
	for len(s) > 0 {
		end := chunkW
		if end > len(s) {
			end = len(s)
		}
		lines = append(lines, "  "+lipgloss.NewStyle().Foreground(colorGreen).Render(s[:end]))
		s = s[end:]
	}
	return strings.Join(lines, "\n")
}
