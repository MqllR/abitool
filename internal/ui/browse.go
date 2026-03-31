package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"

	"github.com/MqllR/abitool/internal/abitool"
	"github.com/MqllR/abitool/pkg/abiparser"
	"github.com/MqllR/abitool/pkg/chains"
	abistore "github.com/MqllR/abitool/pkg/storage/abi"
)

// ─── browseModel ──────────────────────────────────────────────────────────────

type browseModel struct {
	address  string
	name     string
	basePath string
	chainID  int

	elements []abiparser.Element
	filtered []abiparser.Element

	cursor int
	offset int

	filter   textinput.Model
	focusing bool

	loaded bool
	err    error

	width  int
	height int
}

type abiLoadedMsg struct{ elements []abiparser.Element }
type abiLoadErrMsg struct{ err error }

func newBrowseModel(address, name, basePath string, chainID int) browseModel {
	fi := textinput.New()
	fi.Placeholder = "filter by name or type..."
	fi.CharLimit = 64
	return browseModel{
		address:  address,
		name:     name,
		basePath: basePath,
		chainID:  chainID,
		filter:   fi,
	}
}

func (m browseModel) setSize(w, h int) screen { m.width, m.height = w, h; return m }

func (m browseModel) Init() tea.Cmd {
	return loadABICmd(m.basePath, m.address)
}

func loadABICmd(basePath, address string) tea.Cmd {
	return func() tea.Msg {
		store, err := abistore.NewLocal(basePath)
		if err != nil {
			return abiLoadErrMsg{err}
		}
		raw, err := store.Read(address)
		if err != nil {
			return abiLoadErrMsg{err}
		}
		parsed, err := abiparser.ParseABI(raw)
		if err != nil {
			return abiLoadErrMsg{err}
		}
		var elements []abiparser.Element
		for el := range parsed.All() {
			elements = append(elements, el)
		}
		return abiLoadedMsg{elements}
	}
}

func (m browseModel) applyFilter() []abiparser.Element {
	q := strings.ToLower(m.filter.Value())
	if q == "" {
		return m.elements
	}
	var out []abiparser.Element
	for _, el := range m.elements {
		if strings.Contains(strings.ToLower(el.Name), q) ||
			strings.Contains(strings.ToLower(string(el.Type)), q) {
			out = append(out, el)
		}
	}
	return out
}

func (m browseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case abiLoadedMsg:
		m.elements = msg.elements
		m.filtered = m.applyFilter()
		m.loaded = true
		return m, nil

	case abiLoadErrMsg:
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
				m.filtered = m.elements
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
		case "enter", "c":
			if m.loaded && len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				el := m.filtered[m.cursor]
				if el.IsFunction() && isReadOnly(el.StateMutability) {
					rpcURL := resolveRPCURL(m.chainID)
					next := newCallFormScreen(m.address, el, rpcURL)
					return m, func() tea.Msg { return pushMsg{next} }
				}
			}
		}
	}
	return m, nil
}

func (m browseModel) visibleRows() int {
	// border(2) + title(1) + filter(1) + sep(1) + sep(1) + status(1) = 7
	rows := m.height - 7
	if rows < 1 {
		rows = 1
	}
	return rows
}

// ─── View routing ─────────────────────────────────────────────────────────────

func (m browseModel) View() string {
	w, h := m.width, m.height
	if w < 10 {
		w = 80
	}
	if h < 5 {
		h = 24
	}

	if !m.loaded {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center,
			dimStyle.Render("  Loading ABI..."))
	}
	if m.err != nil {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center,
			errorStyle.Render("  Error: "+m.err.Error()))
	}

	if w < 80 {
		return m.renderNarrow(w, h)
	}
	return m.renderSplit(w, h)
}

// ─── renderSplit: two-pane layout ─────────────────────────────────────────────

func (m browseModel) renderSplit(w, h int) string {
	// -2 for left+right border, no padding
	innerW := w - 2
	leftW := innerW * 38 / 100
	// -3 for " │ " separator between panes
	rightW := innerW - leftW - 3

	vr := m.visibleRows()

	// Filter bar
	var filterBar string
	if m.focusing {
		filterBar = "  🔍 " + m.filter.View()
	} else if m.filter.Value() != "" {
		filterBar = dimStyle.Render("  filter: ") + m.filter.Value() + dimStyle.Render("  (esc to clear)")
	} else {
		filterBar = dimStyle.Render("  / to filter")
	}

	sep := dimStyle.Render(strings.Repeat("─", innerW))

	// Left pane: element list
	leftLines := m.buildListLines(leftW, vr)

	// Right pane: detail view
	var rightLines []string
	if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
		rightLines = m.buildDetailLines(m.filtered[m.cursor], rightW)
	}

	// Pad both panes to exactly vr lines
	for len(leftLines) < vr {
		leftLines = append(leftLines, "")
	}
	for len(rightLines) < vr {
		rightLines = append(rightLines, "")
	}

	// Combine panes row by row
	var body strings.Builder
	for i := 0; i < vr; i++ {
		l := lipgloss.NewStyle().Width(leftW).MaxWidth(leftW).Render(leftLines[i])
		r := lipgloss.NewStyle().Width(rightW).MaxWidth(rightW).Render(rightLines[i])
		body.WriteString(l + dimStyle.Render(" │ ") + r + "\n")
	}

	// Status bar with item counter
	status := dimStyle.Render("  ↑↓/jk navigate  / filter  esc back  q quit")
	if len(m.filtered) > 0 {
		hint := "  ↑↓/jk navigate  / filter  esc back  q quit"
		if m.cursor < len(m.filtered) {
			el := m.filtered[m.cursor]
			if el.IsFunction() && isReadOnly(el.StateMutability) {
				hint = "  ↑↓/jk navigate  enter/c call  / filter  esc back  q quit"
			}
		}
		status = dimStyle.Render(fmt.Sprintf("  [%d/%d]  %s",
			m.cursor+1, len(m.filtered), strings.TrimPrefix(hint, "  ")))
	}

	title := titleStyle.Render("  "+m.name) + "  " +
		dimStyle.Render(fmt.Sprintf("[%s (%d)]", chains.Name(m.chainID), m.chainID))
	content := title + "\n" +
		filterBar + "\n" +
		sep + "\n" +
		strings.TrimRight(body.String(), "\n") + "\n" +
		sep + "\n" +
		status

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Width(innerW).
		Height(h - 2).
		Render(content)
}

// ─── renderNarrow: single-column fallback for narrow terminals ────────────────

func (m browseModel) renderNarrow(w, h int) string {
	innerW := w - 2
	vr := m.visibleRows()

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("  "+m.name) + "  " +
		dimStyle.Render(fmt.Sprintf("[%s (%d)]", chains.Name(m.chainID), m.chainID)) + "\n")

	if m.focusing {
		sb.WriteString("  🔍 " + m.filter.View() + "\n")
	} else {
		sb.WriteString(dimStyle.Render("  / to filter") + "\n")
	}
	sb.WriteString(dimStyle.Render(strings.Repeat("─", innerW)) + "\n")

	listLines := m.buildListLines(innerW, vr)
	for _, l := range listLines {
		sb.WriteString(l + "\n")
	}

	sb.WriteString(dimStyle.Render(strings.Repeat("─", innerW)) + "\n")
	sb.WriteString(dimStyle.Render("  ↑↓ navigate  / filter  esc back  q quit"))

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Width(innerW).
		Height(h - 2).
		Render(sb.String())
}

// ─── buildListLines ───────────────────────────────────────────────────────────

func (m browseModel) buildListLines(colW, maxRows int) []string {
	if len(m.filtered) == 0 {
		return []string{dimStyle.Render("  (no elements)")}
	}

	end := m.offset + maxRows
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	lines := make([]string, 0, end-m.offset)
	for i := m.offset; i < end; i++ {
		el := m.filtered[i]
		badge := elementBadge(el)
		name := el.Name
		if name == "" {
			name = string(el.Type)
		}
		// Ensure name fits in the column (badge=5 + spaces=2 = 7 overhead, rest for name).
		maxName := colW - 7
		if maxName > 0 && len(name) > maxName {
			name = name[:maxName-1] + "…"
		}

		var line string
		if i == m.cursor {
			line = selectedStyle.Render(" " + badge + "  " + name)
		} else {
			line = " " + badge + "  " + name
		}
		lines = append(lines, line)
	}
	return lines
}

// ─── buildDetailLines ─────────────────────────────────────────────────────────

func (m browseModel) buildDetailLines(el abiparser.Element, colW int) []string {
	var lines []string

	// Header: type  name  [anon]
	typePart := lipgloss.NewStyle().Foreground(colorDim).Bold(true).Render(string(el.Type))
	namePart := lipgloss.NewStyle().Foreground(colorWhite).Bold(true).Render(el.Name)
	header := " " + typePart
	if el.Name != "" {
		header += "  " + namePart
	}
	if el.Anonymous {
		header += "  " + dimStyle.Render("[anon]")
	}
	lines = append(lines, header)
	lines = append(lines, dimStyle.Render(" "+strings.Repeat("─", colW-2)))

	// Selector (functions and errors) / Topic hash (events)
	switch {
	case el.HasSelector():
		selector := dimStyle.Render("N/A")
		if sel, err := el.Selector(); err == nil {
			selector = lipgloss.NewStyle().Foreground(colorBlue).Render(sel)
		}
		lines = append(lines, detailRow("Selector", selector))
	case el.HasTopicHash():
		topic := dimStyle.Render("N/A")
		if th, err := el.TopicHash(); err == nil {
			topic = lipgloss.NewStyle().Foreground(colorBlue).Render(th)
		}
		lines = append(lines, detailRow("Topic[0]", topic))
	default:
		lines = append(lines, detailRow("Selector", dimStyle.Render("N/A")))
	}

	// State mutability
	lines = append(lines, detailRow("Mutability", mutabilityStyled(el.StateMutability)))

	lines = append(lines, "")

	// Inputs
	lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Bold(true).Render(" Inputs"))
	if len(el.Inputs) == 0 {
		lines = append(lines, dimStyle.Render("  (none)"))
	} else {
		for _, inp := range el.Inputs {
			lines = append(lines, formatParam(inp.Parameter, false, colW))
		}
	}

	// Outputs — only for functions and fallback
	if el.Type == abiparser.FunctionType || el.Type == abiparser.FallbackType {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDim).Bold(true).Render(" Outputs"))
		if len(el.Outputs) == 0 {
			lines = append(lines, dimStyle.Render("  (none)"))
		} else {
			for _, out := range el.Outputs {
				lines = append(lines, formatParam(out.Parameter, true, colW))
			}
		}
	}

	return lines
}

func detailRow(label, value string) string {
	labelS := lipgloss.NewStyle().Foreground(colorDim).Width(12).Render(label)
	return " " + labelS + value
}

func mutabilityStyled(sm abiparser.StateMutability) string {
	switch sm {
	case "view", "pure":
		return lipgloss.NewStyle().Foreground(colorGreen).Render(string(sm))
	case "payable":
		return lipgloss.NewStyle().Foreground(colorYellow).Render(string(sm))
	case "":
		return dimStyle.Render("—")
	default:
		return string(sm)
	}
}

func formatParam(p abiparser.Parameter, isOutput bool, _ int) string {
	typeColor := colorBlue
	if isOutput {
		typeColor = colorGreen
	}
	typeS := lipgloss.NewStyle().Foreground(typeColor).Render(p.Type)

	name := p.Name
	if name == "" {
		name = "_"
	}
	nameS := lipgloss.NewStyle().Foreground(colorWhite).Render(name)

	line := "  " + nameS + "  " + typeS

	if p.Indexed {
		line += dimStyle.Render("  [idx]")
	}

	if p.InternalType != "" && p.InternalType != p.Type {
		line += dimStyle.Render("  // "+p.InternalType)
	}
	return line
}

// ─── elementBadge ─────────────────────────────────────────────────────────────

func elementBadge(el abiparser.Element) string {
	switch el.Type {
	case abiparser.FunctionType:
		return badgeFunction.Render("[fn]")
	case abiparser.EventType:
		return badgeEvent.Render("[ev]")
	case abiparser.ErrorType:
		return badgeError.Render("[er]")
	case abiparser.ConstructorType:
		return badgeConstructor.Render("[co]")
	default:
		return badgeFallback.Render("[fb]")
	}
}

// isReadOnly reports whether a state mutability is read-only (view or pure).
func isReadOnly(sm abiparser.StateMutability) bool {
	return sm == "view" || sm == "pure"
}

// resolveRPCURL returns the best available RPC URL for the given chain ID.
// Resolution order: --rpc-url flag → rpc.url config → hardcoded chain default → "".
func resolveRPCURL(chainID int) string {
	if url := viper.GetString("rpc-url"); url != "" {
		return url
	}
	if url := abitool.ConfigInstance().RPC.URL; url != "" {
		return url
	}
	if info, ok := chains.Known[chainID]; ok {
		return info.DefaultRPCURL
	}
	return ""
}
