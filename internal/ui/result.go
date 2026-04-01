// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/MqllR/abitool/pkg/abicodec"
	"github.com/MqllR/abitool/pkg/abiparser"
	"github.com/MqllR/abitool/pkg/ethclient"
)

// callResultScreen executes an eth_call and displays the decoded output.
// The call is fired asynchronously from Init() so the screen shows a loading
// placeholder until the result arrives.
type callResultScreen struct {
	address string
	element abiparser.Element
	rpcURL  string
	args    []string

	values []interface{}
	err    error
	loaded bool

	width  int
	height int
}

// callDoneMsg carries the result of an eth_call back to the screen.
type callDoneMsg struct {
	values []interface{}
	err    error
}

func newCallResultScreen(address string, el abiparser.Element, rpcURL string, args []string) callResultScreen {
	return callResultScreen{
		address: address,
		element: el,
		rpcURL:  rpcURL,
		args:    args,
	}
}

func (m callResultScreen) setSize(w, h int) screen { m.width, m.height = w, h; return m }

func (m callResultScreen) Init() tea.Cmd {
	return executeCallCmd(m.rpcURL, m.address, m.element, m.args)
}

func executeCallCmd(rpcURL, address string, el abiparser.Element, args []string) tea.Cmd {
	return func() tea.Msg {
		method, err := abicodec.ParseMethod(el)
		if err != nil {
			return callDoneMsg{err: fmt.Errorf("parsing method: %w", err)}
		}

		calldata, err := abicodec.EncodeInput(method, args)
		if err != nil {
			return callDoneMsg{err: fmt.Errorf("encoding calldata: %w", err)}
		}

		client, err := ethclient.Dial(context.Background(), rpcURL)
		if err != nil {
			return callDoneMsg{err: err}
		}
		defer client.Close()

		raw, err := client.CallContract(context.Background(), address, calldata, "latest")
		if err != nil {
			return callDoneMsg{err: err}
		}

		values, err := abicodec.DecodeOutput(method, raw)
		if err != nil {
			return callDoneMsg{err: fmt.Errorf("decoding output: %w", err)}
		}

		return callDoneMsg{values: values}
	}
}

func (m callResultScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case callDoneMsg:
		m.values = msg.values
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

func (m callResultScreen) View() string {
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
	sb.WriteString(titleStyle.Render("  Result: "+fnName) + "\n")
	sb.WriteString(dimStyle.Render("  "+strings.Repeat("─", max(w-10, 10))) + "\n\n")

	switch {
	case !m.loaded:
		sb.WriteString(dimStyle.Render("  Calling...") + "\n")
	case m.err != nil:
		sb.WriteString(errorStyle.Render("  Error: "+m.err.Error()) + "\n")
	case len(m.values) == 0:
		sb.WriteString(dimStyle.Render("  (no return value)") + "\n")
	default:
		for i, v := range m.values {
			label := dimStyle.Render(fmt.Sprintf("  [%d]  ", i))
			value := lipgloss.NewStyle().Foreground(colorGreen).Render(fmt.Sprintf("%v", v))
			sb.WriteString(label + value + "\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("  esc / backspace: back  q: quit"))

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
