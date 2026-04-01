// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

// Package ui provides terminal UI components built with bubbletea.
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// FormField describes a single input field in the form.
type FormField struct {
	Name string
	Type string
}

// formModel is the bubbletea model for a multi-field input form.
type formModel struct {
	fields  []FormField
	inputs  []textinput.Model
	cursor  int
	done    bool
	aborted bool
}

func newFormModel(fields []FormField) formModel {
	inputs := make([]textinput.Model, len(fields))
	for i, f := range fields {
		ti := textinput.New()
		ti.Placeholder = f.Type
		ti.Prompt = fmt.Sprintf("  %s (%s): ", f.Name, f.Type)
		ti.CharLimit = 256
		inputs[i] = ti
	}
	if len(inputs) > 0 {
		inputs[0].Focus()
	}

	return formModel{
		fields: fields,
		inputs: inputs,
	}
}

func (m formModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.aborted = true
			return m, tea.Quit

		case tea.KeyEnter, tea.KeyTab:
			if m.cursor < len(m.inputs)-1 {
				m.inputs[m.cursor].Blur()
				m.cursor++
				m.inputs[m.cursor].Focus()
				return m, textinput.Blink
			}
			// Last field: submit.
			m.done = true
			return m, tea.Quit
		}
	}

	// Forward keystrokes to the focused input.
	var cmd tea.Cmd
	m.inputs[m.cursor], cmd = m.inputs[m.cursor].Update(msg)
	return m, cmd
}

func (m formModel) View() string {
	var b strings.Builder

	b.WriteString("\n  Fill in the arguments (Tab / Enter to advance, Esc to cancel)\n\n")

	for i, input := range m.inputs {
		if i == m.cursor {
			b.WriteString("▸ ")
		} else {
			b.WriteString("  ")
		}
		b.WriteString(input.View())
		b.WriteByte('\n')
	}

	b.WriteByte('\n')
	return b.String()
}

// RunForm displays an interactive form for the given fields and returns the
// values entered by the user. Returns an error if the user aborts (Esc / Ctrl-C).
func RunForm(fields []FormField) ([]string, error) {
	if len(fields) == 0 {
		return nil, nil
	}

	m := newFormModel(fields)
	p := tea.NewProgram(m)

	result, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("running input form: %w", err)
	}

	final := result.(formModel)
	if final.aborted {
		return nil, fmt.Errorf("aborted by user")
	}

	values := make([]string, len(final.inputs))
	for i, inp := range final.inputs {
		values[i] = inp.Value()
	}

	return values, nil
}
