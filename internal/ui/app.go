// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

// Package ui provides terminal UI components built with bubbletea.
package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"

	"github.com/MqllR/abitool/internal/abitool"
	"github.com/MqllR/abitool/pkg/chains"
	"github.com/MqllR/abitool/pkg/etherscan"
	abistore "github.com/MqllR/abitool/pkg/storage/abi"
	contractstore "github.com/MqllR/abitool/pkg/storage/contract"
)

// ─── Shared colours & styles ──────────────────────────────────────────────────

var (
	colorPrimary = lipgloss.Color("#7D56F4")
	colorDim     = lipgloss.Color("#6272A4")
	colorWhite   = lipgloss.Color("#F8F8F2")
	colorGreen   = lipgloss.Color("#50FA7B")
	colorRed     = lipgloss.Color("#FF5555")
	colorYellow  = lipgloss.Color("#F9D449")
	colorBlue    = lipgloss.Color("#4BAFED")

	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(colorPrimary)
	dimStyle      = lipgloss.NewStyle().Foreground(colorDim)
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(colorWhite).Background(colorPrimary)
	errorStyle    = lipgloss.NewStyle().Foreground(colorRed)
	successStyle  = lipgloss.NewStyle().Foreground(colorGreen)

	// Type badge styles — also used by browse.go (same package).
	badgeFunction    = lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	badgeEvent       = lipgloss.NewStyle().Bold(true).Foreground(colorYellow)
	badgeError       = lipgloss.NewStyle().Bold(true).Foreground(colorRed)
	badgeConstructor = lipgloss.NewStyle().Bold(true).Foreground(colorGreen)
	badgeFallback    = lipgloss.NewStyle().Foreground(colorDim)
)

// ─── Navigation messages ──────────────────────────────────────────────────────

type pushMsg struct{ next screen }
type popMsg struct{}

// popAndRefreshMsg pops the current screen and re-initialises the one below.
type popAndRefreshMsg struct{}

// chainSelectedMsg is emitted by the chain selector to switch the active chain.
type chainSelectedMsg struct{ chainID int }

// ─── screen interface ─────────────────────────────────────────────────────────

// screen is implemented by every TUI screen in the stack.
type screen interface {
	tea.Model
	setSize(w, h int) screen
}

// ─── appModel (root) ──────────────────────────────────────────────────────────

type appModel struct {
	stack     []screen
	width     int
	height    int
	chainID   int
	apiKey    string
	storePath string
}

func newAppModel(initial screen, chainID int, apiKey, storePath string) appModel {
	return appModel{
		stack:     []screen{initial},
		width:     80,
		height:    24,
		chainID:   chainID,
		apiKey:    apiKey,
		storePath: storePath,
	}
}

func (m appModel) top() screen { return m.stack[len(m.stack)-1] }

func (m appModel) Init() tea.Cmd { return m.top().Init() }

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		for i, s := range m.stack {
			m.stack[i] = s.setSize(m.width, m.height)
		}
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

	case pushMsg:
		next := msg.next.setSize(m.width, m.height)
		m.stack = append(m.stack, next)
		return m, next.Init()

	case popMsg:
		if len(m.stack) <= 1 {
			return m, tea.Quit
		}
		m.stack = m.stack[:len(m.stack)-1]
		return m, nil

	case popAndRefreshMsg:
		if len(m.stack) <= 1 {
			return m, tea.Quit
		}
		m.stack = m.stack[:len(m.stack)-1]
		return m, m.top().Init()

	case chainSelectedMsg:
		m.chainID = msg.chainID
		_ = abitool.SaveChainID(msg.chainID)
		basePath := filepath.Join(m.storePath, strconv.Itoa(m.chainID))
		home := newHomeModel(basePath, m.chainID, m.apiKey).setSize(m.width, m.height)
		m.stack = []screen{home}
		return m, home.Init()
	}

	// Delegate to the top screen.
	updated, cmd := m.top().Update(msg)
	m.stack[len(m.stack)-1] = updated.(screen)
	return m, cmd
}

func (m appModel) View() string {
	if len(m.stack) == 0 {
		return ""
	}
	return m.top().View()
}

// ─── homeModel ────────────────────────────────────────────────────────────────

type homeModel struct {
	cursor   int
	basePath string
	chainID  int
	apiKey   string
	count    int
	width    int
	height   int
}

type menuEntry struct {
	label string
	desc  string
}

func (m homeModel) menuItems() []menuEntry {
	countLabel := ""
	switch m.count {
	case 0:
		countLabel = "none stored"
	case 1:
		countLabel = "1 stored"
	default:
		countLabel = fmt.Sprintf("%d stored", m.count)
	}
	return []menuEntry{
		{"Contracts", countLabel},
		{"Download", "fetch a new ABI from Etherscan"},
		{"Switch Chain", fmt.Sprintf("current: %s (%d)", chains.Name(m.chainID), m.chainID)},
	}
}

func newHomeModel(basePath string, chainID int, apiKey string) homeModel {
	count := 0
	if cs, err := contractstore.NewLocal(basePath); err == nil {
		if it, err := cs.List(); err == nil {
			for range it {
				count++
			}
		}
	}
	return homeModel{
		basePath: basePath,
		chainID:  chainID,
		apiKey:   apiKey,
		count:    count,
	}
}

func (m homeModel) setSize(w, h int) screen { m.width, m.height = w, h; return m }
func (m homeModel) Init() tea.Cmd           { return nil }

func (m homeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	items := m.menuItems()
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(items)-1 {
				m.cursor++
			}
		case "enter", " ":
			switch m.cursor {
			case 0:
				next := newContractListModel(m.basePath, m.chainID)
				return m, func() tea.Msg { return pushMsg{next} }
			case 1:
				next := newDownloadModel(m.basePath, m.chainID, m.apiKey)
				return m, func() tea.Msg { return pushMsg{next} }
			case 2:
				next := newChainSelectorModel(m.chainID)
				return m, func() tea.Msg { return pushMsg{next} }
			}
		}
	}
	return m, nil
}

func (m homeModel) View() string {
	items := m.menuItems()
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("⬡  Ethereum ABI Tool") + "  " +
		dimStyle.Render(fmt.Sprintf("[%s (%d)]", chains.Name(m.chainID), m.chainID)) + "\n\n")

	for i, item := range items {
		var line string
		if i == m.cursor {
			label := selectedStyle.Render(" " + item.label + " ")
			desc := lipgloss.NewStyle().Foreground(colorDim).Render("  " + item.desc)
			line = "  ❯ " + label + desc
		} else {
			desc := dimStyle.Render("  " + item.desc)
			line = "    " + item.label + desc
		}
		sb.WriteString(line + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("↑↓/jk navigate  enter select  q quit"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 3).
		Render(sb.String())

	w, h := m.width, m.height
	if w < 10 {
		w = 80
	}
	if h < 5 {
		h = 24
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

// ─── contractListModel ────────────────────────────────────────────────────────

type contractEntry struct {
	address string
	name    string
}

type contractListModel struct {
	all      []contractEntry
	filtered []contractEntry
	cursor   int
	offset   int
	filter   textinput.Model
	focusing bool
	basePath string
	chainID  int
	loaded   bool
	err      error
	width    int
	height   int
}

type contractsLoadedMsg struct{ entries []contractEntry }
type contractsErrMsg struct{ err error }

func newContractListModel(basePath string, chainID int) contractListModel {
	fi := textinput.New()
	fi.Placeholder = "filter contracts..."
	fi.CharLimit = 64
	return contractListModel{
		basePath: basePath,
		chainID:  chainID,
		filter:   fi,
	}
}

func (m contractListModel) setSize(w, h int) screen { m.width, m.height = w, h; return m }

func (m contractListModel) Init() tea.Cmd {
	return loadContractsCmd(m.basePath)
}

func loadContractsCmd(basePath string) tea.Cmd {
	return func() tea.Msg {
		cs, err := contractstore.NewLocal(basePath)
		if err != nil {
			return contractsErrMsg{err}
		}
		it, err := cs.List()
		if err != nil {
			return contractsErrMsg{err}
		}
		var entries []contractEntry
		for address := range it {
			name := address
			if raw, err := cs.Get(address); err == nil {
				var meta struct {
					ContractName string `json:"contract_name"`
				}
				if err := json.Unmarshal(raw, &meta); err == nil && meta.ContractName != "" {
					name = meta.ContractName
				}
			}
			entries = append(entries, contractEntry{address: address, name: name})
		}
		return contractsLoadedMsg{entries}
	}
}

func (m contractListModel) applyFilter() []contractEntry {
	q := strings.ToLower(m.filter.Value())
	if q == "" {
		return m.all
	}
	var out []contractEntry
	for _, e := range m.all {
		if strings.Contains(strings.ToLower(e.address), q) ||
			strings.Contains(strings.ToLower(e.name), q) {
			out = append(out, e)
		}
	}
	return out
}

func (m contractListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case contractsLoadedMsg:
		m.all = msg.entries
		m.filtered = m.applyFilter()
		m.loaded = true
		return m, nil

	case contractsErrMsg:
		m.err = msg.err
		m.loaded = true
		return m, nil

	case tea.KeyMsg:
		if m.focusing {
			switch msg.Type {
			case tea.KeyEsc:
				m.focusing = false
				m.filter.Blur()
				m.filter.SetValue("")
				m.filtered = m.all
				m.cursor, m.offset = 0, 0
			case tea.KeyEnter:
				m.focusing = false
				m.filter.Blur()
			default:
				var cmd tea.Cmd
				m.filter, cmd = m.filter.Update(msg)
				m.filtered = m.applyFilter()
				m.cursor, m.offset = 0, 0
				return m, cmd
			}
			return m, nil
		}

		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "esc", "backspace":
			return m, func() tea.Msg { return popMsg{} }
		case "/":
			m.focusing = true
			m.filter.Focus()
			return m, textinput.Blink
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				if vr := m.visibleRows(); m.cursor >= m.offset+vr {
					m.offset = m.cursor - vr + 1
				}
			}
		case "enter", " ":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				entry := m.filtered[m.cursor]
				next := newBrowseModel(entry.address, entry.name, m.basePath, m.chainID)
				return m, func() tea.Msg { return pushMsg{next} }
			}
		}
	}
	return m, nil
}

func (m contractListModel) visibleRows() int {
	// border(2) + title(1) + filter(1) + sep(1) + sep(1) + status(1) = 7
	rows := m.height - 7
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (m contractListModel) View() string {
	w, h := m.width, m.height
	if w < 10 {
		w = 80
	}
	if h < 5 {
		h = 24
	}
	innerW := w - 2 // subtract border chars

	var sb strings.Builder

	// Title row
	sb.WriteString(titleStyle.Render("  Contracts") + "  " +
		dimStyle.Render(fmt.Sprintf("[%s (%d)]", chains.Name(m.chainID), m.chainID)) + "\n")

	// Filter bar
	if m.focusing {
		sb.WriteString("  🔍 " + m.filter.View() + "\n")
	} else if m.filter.Value() != "" {
		sb.WriteString(dimStyle.Render("  filter: ") + m.filter.Value() + dimStyle.Render("  (esc to clear)") + "\n")
	} else {
		sb.WriteString(dimStyle.Render("  / to filter") + "\n")
	}

	sb.WriteString(dimStyle.Render(strings.Repeat("─", innerW)) + "\n")

	// List content
	if !m.loaded {
		sb.WriteString(dimStyle.Render("  Loading...") + "\n")
	} else if m.err != nil {
		sb.WriteString(errorStyle.Render("  Error: "+m.err.Error()) + "\n")
	} else if len(m.all) == 0 {
		sb.WriteString(dimStyle.Render("  No contracts downloaded yet. Press esc to go back.") + "\n")
	} else if len(m.filtered) == 0 {
		sb.WriteString(dimStyle.Render("  No matches.") + "\n")
	} else {
		vr := m.visibleRows()
		end := m.offset + vr
		if end > len(m.filtered) {
			end = len(m.filtered)
		}
		for i := m.offset; i < end; i++ {
			e := m.filtered[i]
			addr := shortenAddr(e.address)
			var line string
			if i == m.cursor {
				line = selectedStyle.Render(fmt.Sprintf("  %-14s  %s", addr, e.name))
			} else {
				line = fmt.Sprintf("  %-14s  %s", addr, dimStyle.Render(e.name))
			}
			sb.WriteString(line + "\n")
		}
		if len(m.filtered) > vr {
			sb.WriteString(dimStyle.Render(fmt.Sprintf("  [%d/%d]", m.cursor+1, len(m.filtered))) + "\n")
		}
	}

	sb.WriteString(dimStyle.Render(strings.Repeat("─", innerW)) + "\n")
	sb.WriteString(dimStyle.Render("  ↑↓/jk navigate  / filter  enter browse  esc back  q quit"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Width(innerW).
		Height(h - 2).
		Render(sb.String())
}

func shortenAddr(addr string) string {
	if len(addr) <= 13 {
		return addr
	}
	return addr[:6] + "…" + addr[len(addr)-4:]
}

// ─── downloadModel ────────────────────────────────────────────────────────────

type dlState int

const (
	dlIdle dlState = iota
	dlLoading
	dlSuccess
	dlError
)

type downloadModel struct {
	input    textinput.Model
	state    dlState
	message  string
	basePath string
	chainID  int
	apiKey   string
	width    int
	height   int
}

type downloadDoneMsg struct {
	address string
	name    string
}
type downloadErrMsg struct{ err error }

func newDownloadModel(basePath string, chainID int, apiKey string) downloadModel {
	ti := textinput.New()
	ti.Placeholder = "0x..."
	ti.Prompt = "  Contract address: "
	ti.CharLimit = 42
	ti.Focus()
	return downloadModel{
		input:    ti,
		basePath: basePath,
		chainID:  chainID,
		apiKey:   apiKey,
	}
}

func (m downloadModel) setSize(w, h int) screen { m.width, m.height = w, h; return m }
func (m downloadModel) Init() tea.Cmd           { return textinput.Blink }

func (m downloadModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case downloadDoneMsg:
		m.state = dlSuccess
		m.message = fmt.Sprintf("✓  Downloaded %s  (%s)", msg.name, msg.address)
		return m, nil

	case downloadErrMsg:
		m.state = dlError
		m.message = msg.err.Error()
		return m, nil

	case tea.KeyMsg:
		if m.state == dlSuccess {
			return m, func() tea.Msg { return popAndRefreshMsg{} }
		}
		if m.state == dlError {
			if msg.String() == "esc" || msg.String() == "q" {
				return m, func() tea.Msg { return popMsg{} }
			}
			// Any other key: reset to idle to let the user try again.
			m.state = dlIdle
			m.message = ""
			return m, nil
		}
		switch msg.Type {
		case tea.KeyEsc:
			return m, func() tea.Msg { return popMsg{} }
		case tea.KeyEnter:
			if m.state != dlLoading {
				addr := strings.TrimSpace(m.input.Value())
				if addr == "" {
					return m, nil
				}
				m.state = dlLoading
				return m, doDownload(m.basePath, m.apiKey, m.chainID, addr)
			}
		}
	}

	if m.state == dlIdle {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func doDownload(basePath, apiKey string, chainID int, address string) tea.Cmd {
	return func() tea.Msg {
		if apiKey == "" {
			return downloadErrMsg{fmt.Errorf("Etherscan API key not configured (set etherscan.api_key in config)")}
		}
		client := etherscan.NewClient(apiKey, etherscan.FromInt(chainID))
		src, err := client.GetSourceCode(context.Background(), address)
		if err != nil {
			return downloadErrMsg{err}
		}

		as, err := abistore.NewLocal(basePath)
		if err != nil {
			return downloadErrMsg{err}
		}
		if err := as.Write(address, src.ABI); err != nil {
			return downloadErrMsg{err}
		}

		metaJSON, err := json.Marshal(map[string]any{
			"contract_name": src.ContractName,
			"abi_path":      as.GetPath(address),
		})
		if err != nil {
			_ = as.Delete(address)
			return downloadErrMsg{err}
		}

		cs, err := contractstore.NewLocal(basePath)
		if err != nil {
			_ = as.Delete(address)
			return downloadErrMsg{err}
		}
		if err := cs.Add(address, metaJSON); err != nil {
			_ = as.Delete(address)
			return downloadErrMsg{fmt.Errorf("saving contract metadata: %w", err)}
		}

		return downloadDoneMsg{address: address, name: src.ContractName}
	}
}

func (m downloadModel) View() string {
	w, h := m.width, m.height
	if w < 10 {
		w = 80
	}
	if h < 5 {
		h = 24
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("  Download ABI") + "  " +
		dimStyle.Render(fmt.Sprintf("[%s (%d)]", chains.Name(m.chainID), m.chainID)) + "\n\n")

	switch m.state {
	case dlIdle:
		sb.WriteString(m.input.View() + "\n\n")
		sb.WriteString(dimStyle.Render("  enter submit  esc back"))
	case dlLoading:
		sb.WriteString(dimStyle.Render("  ⏳  Downloading "+m.input.Value()+"...") + "\n")
	case dlSuccess:
		sb.WriteString(successStyle.Render(m.message) + "\n\n")
		sb.WriteString(dimStyle.Render("  Press any key to continue..."))
	case dlError:
		sb.WriteString(errorStyle.Render("  ✗  "+m.message) + "\n\n")
		sb.WriteString(dimStyle.Render("  Press any key to try again  esc to go back"))
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 3).
		Render(sb.String())

	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

// ─── chainSelectorModel ───────────────────────────────────────────────────────

// chainSelectorEntry represents a selectable chain in the picker.
type chainSelectorEntry struct {
	id   int
	name string
}

type chainSelectorModel struct {
	entries    []chainSelectorEntry
	cursor     int
	offset     int
	customInput textinput.Model
	customMode  bool
	currentID  int
	width      int
	height     int
}

func newChainSelectorModel(currentID int) chainSelectorModel {
	entries := []chainSelectorEntry{
		{1, "Ethereum Mainnet"},
		{10, "Optimism"},
		{56, "BNB Chain"},
		{100, "Gnosis"},
		{137, "Polygon"},
		{8453, "Base"},
		{42161, "Arbitrum One"},
		{43114, "Avalanche"},
		{59144, "Linea"},
		{11155111, "Sepolia"},
		{11155420, "Optimism Sepolia"},
		{84532, "Base Sepolia"},
		{421614, "Arbitrum Sepolia"},
		{80002, "Polygon Amoy"},
	}

	// Start cursor on the current chain if it's in the list.
	cursor := 0
	for i, e := range entries {
		if e.id == currentID {
			cursor = i
			break
		}
	}

	ci := textinput.New()
	ci.Placeholder = "enter chain ID..."
	ci.CharLimit = 12

	return chainSelectorModel{
		entries:     entries,
		cursor:      cursor,
		currentID:   currentID,
		customInput: ci,
	}
}

func (m chainSelectorModel) setSize(w, h int) screen { m.width, m.height = w, h; return m }
func (m chainSelectorModel) Init() tea.Cmd           { return nil }

func (m chainSelectorModel) visibleRows() int {
	// border(2) + title(1) + sep(1) + custom(1) + sep(1) + status(1) = 7
	rows := m.height - 7
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (m chainSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.customMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEsc:
				m.customMode = false
				m.customInput.Blur()
				m.customInput.SetValue("")
				return m, nil
			case tea.KeyEnter:
				id, err := strconv.Atoi(strings.TrimSpace(m.customInput.Value()))
				if err != nil || id <= 0 {
					m.customMode = false
					m.customInput.Blur()
					m.customInput.SetValue("")
					return m, nil
				}
				return m, func() tea.Msg { return chainSelectedMsg{id} }
			default:
				var cmd tea.Cmd
				m.customInput, cmd = m.customInput.Update(msg)
				return m, cmd
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "esc", "backspace":
			return m, func() tea.Msg { return popMsg{} }
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
				if vr := m.visibleRows(); m.cursor >= m.offset+vr {
					m.offset = m.cursor - vr + 1
				}
			}
		case "enter", " ":
			return m, func() tea.Msg { return chainSelectedMsg{m.entries[m.cursor].id} }
		case "c":
			m.customMode = true
			m.customInput.Focus()
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m chainSelectorModel) View() string {
	w, h := m.width, m.height
	if w < 10 {
		w = 80
	}
	if h < 5 {
		h = 24
	}
	innerW := w - 2

	var sb strings.Builder

	sb.WriteString(titleStyle.Render("  Select Chain") + "\n")
	sb.WriteString(dimStyle.Render(strings.Repeat("─", innerW)) + "\n")

	vr := m.visibleRows()
	end := m.offset + vr
	if end > len(m.entries) {
		end = len(m.entries)
	}

	for i := m.offset; i < end; i++ {
		e := m.entries[i]
		active := ""
		if e.id == m.currentID {
			active = dimStyle.Render(" ✓")
		}
		line := fmt.Sprintf("  %-20s  %d%s", e.name, e.id, active)
		if i == m.cursor {
			sb.WriteString(selectedStyle.Render(line) + "\n")
		} else {
			sb.WriteString(line + "\n")
		}
	}

	sb.WriteString(dimStyle.Render(strings.Repeat("─", innerW)) + "\n")

	if m.customMode {
		sb.WriteString("  Custom: " + m.customInput.View() + "\n")
	} else {
		sb.WriteString(dimStyle.Render("  c  custom chain ID") + "\n")
	}

	sb.WriteString(dimStyle.Render("  ↑↓/jk navigate  enter select  c custom  esc back  q quit"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Width(innerW).
		Height(h - 2).
		Render(sb.String())
}

// ─── RunApp ───────────────────────────────────────────────────────────────────

// RunApp launches the full-screen TUI dashboard. Config is read from viper
// (already populated by the root command's PersistentPreRun).
func RunApp() error {
	cfg := abitool.ConfigInstance()
	chainID := viper.GetInt("chainid")
	storePath := viper.GetString("abi-store")
	basePath := filepath.Join(storePath, strconv.Itoa(chainID))

	home := newHomeModel(basePath, chainID, cfg.EtherScan.APIKey)
	app := newAppModel(home, chainID, cfg.EtherScan.APIKey, storePath)

	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
