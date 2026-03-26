# AGENTS.md — AI Assistant Context

This file gives AI assistants the context needed to work effectively in this repository.

## What this project is

`abitool` is a Go CLI tool that bridges Ethereum smart contracts and humans. It lets users pull contract ABIs from Etherscan, inspect their interface (functions, events, selectors), and — in future — encode/decode calldata and send transactions directly.

## Tech stack

| Concern | Choice |
|---|---|
| Language | Go 1.25 |
| CLI framework | [cobra](https://github.com/spf13/cobra) |
| Config | [viper](https://github.com/spf13/viper) (YAML, bound to cobra flags) |
| Keccak-256 | `golang.org/x/crypto/sha3` (`NewLegacyKeccak256`) |
| External API | Etherscan v2 REST API |
| TUI framework | [bubbletea](https://github.com/charmbracelet/bubbletea) + [bubbles](https://github.com/charmbracelet/bubbles) |
| TUI styling | [lipgloss](https://github.com/charmbracelet/lipgloss) |

## Repository layout

```
cmd/                   Cobra command definitions (entry points only — no business logic)
  abi.go               Parent "abi" command; registers persistent flags (chainid, abi-store)
  abi/
    download.go        abitool abi download <address>
    view.go            abitool abi view <address>
    list.go            abitool abi list
    delete.go          abitool abi delete <address>
  root.go              Root command; loads config + launches TUI when called with no subcommand

internal/
  abitool/
    app.go             Config loading (sync.Once + viper)
    config.go          Config struct; ConfigInstance() accessor
  contract/
    abi.go             ABIManager — orchestrates download, view, list, delete
    storage.go         ABIManager storage helpers (getContract, saveContractWithABI, …)
    types.go           Contract, Metadata structs
    config.go          SupportedChainIDs map
    display.go         Print / PrintContractList — format-aware output (table/json, type filter)
  ui/
    app.go             Root TUI app model; screen stack; homeModel, contractListModel, downloadModel
    browse.go          browseModel — split-pane ABI element browser (list + detail)
    form.go            RunForm — generic multi-field text-input form (used by rpc call)

pkg/
  abiparser/
    types.go           ABI element types (FunctionType, EventType, …)
    parser.go          ParseABI(), Element.Signature(), Element.Selector()
    print.go           PrettyPrinter (JSON) and TablePrinter with lipgloss colours
  etherscan/
    types.go           ContractSourceCodeResponse
    client.go          HTTP client; call() helper
    contract.go        GetABI(), GetSourceCode()
  storage/
    abi/local.go       File-per-contract ABI storage (Write/Read/Delete/GetPath)
    contract/local.go  contracts.json index (Add/Get/Delete/List); ErrNotFound/ErrAlreadyExists
```

## Key design decisions

- **`internal/` vs `pkg/`** — `internal/` contains app-specific orchestration (ABIManager, config, UI). `pkg/` contains reusable, app-agnostic packages (abiparser, etherscan client, storage backends).
- **Storage** — ABIs are stored as raw JSON files named after their contract address. A `contracts.json` index file holds metadata (contract name, ABI file path). Both live under `<abi-store>/<chainid>/`.
- **Config** — Loaded once via `sync.Once` in `internal/abitool/app.go`, then accessed globally through `ConfigInstance()`. Cobra binds CLI flags into Viper; config file values are the fallback.
- **No external ABI codec** — ABI encoding/decoding is intentionally not implemented yet. Function selectors are computed in-house with Keccak-256 over the canonical signature string.
- **TUI as the default entry point** — Running `abitool` with no subcommand launches the full-screen Bubble Tea dashboard. All CLI subcommands remain fully functional for scripting and piping.

## Data model

```go
// ABI element from the parsed JSON
type Element struct {
    Type            Type            // function | event | error | constructor | receive | fallback
    Name            string
    Inputs          []Input
    StateMutability StateMutability // pure | view | nonpayable | payable
}

type Input struct {
    Name         string
    Type         string       // e.g. "uint256", "address", "tuple", "tuple[]"
    InternalType string       // Solidity-level type hint (struct Foo.Bar)
    Components   []Parameter  // Populated for tuple types (struct members)
}
```

## How function selectors work

`Element.Selector()` → `Element.Signature()` → Keccak-256 → first 4 bytes → `0x...` hex string.

The canonical signature format is `functionName(type1,type2)`. Tuple types must be expanded recursively into `((memberType1,memberType2,...))` — see [known issue #1 in the plan](/.copilot/session-state/*/plan.md).

## Configuration file

```yaml
# $HOME/.config/abitool/config.yaml
etherscan:
  api_key: "YOUR_KEY"
```

The `api_key` field is the only required config value. All other settings have defaults and can be overridden via CLI flags.

## Known open issues (as of last review)

All issues from the initial code review have been resolved. See `docs/ROADMAP.md` for planned features.

## Adding a new command

1. Create `cmd/abi/<name>.go` with a `cobra.Command` exported as `<Name>Cmd`.
2. Register it in `cmd/abi.go` with `abiCmd.AddCommand(abi.<Name>Cmd)`.
3. Business logic goes in `internal/contract/` (new method on `ABIManager` if it involves stored contracts).
4. New Etherscan endpoints go in `pkg/etherscan/`.

## Adding a new chain

Add the chain ID to `SupportedChainIDs` in `internal/contract/config.go`. The Etherscan v2 API already supports multiple chains via the `chainid` query parameter.

---

## TUI architecture & UI/UX principles

This section is the authoritative guide for AI assistants making changes to the terminal UI.

### Overview

The TUI uses the [Bubble Tea](https://github.com/charmbracelet/bubbletea) elm-architecture framework with [Lip Gloss](https://github.com/charmbracelet/lipgloss) for layout and colour. The entry point is `ui.RunApp()`, called from `cmd/root.go` when no subcommand is given.

### Screen-stack navigation model

All screens live in `internal/ui/`. Navigation is managed by a single root `appModel` in `app.go` that owns a `[]screen` stack:

```
appModel.stack = [homeModel, contractListModel, browseModel]
                              ↑ bottom                   ↑ top (active)
```

- **Push** a new screen by returning `func() tea.Msg { return pushMsg{next} }` from `Update`.
- **Pop** (go back) by returning `func() tea.Msg { return popMsg{} }`.
- **Pop + re-init** the screen below (e.g. after a download) by returning `popAndRefreshMsg{}`. This calls `Init()` on the now-top screen, so it reloads fresh data.
- `appModel.Update` propagates `tea.WindowSizeMsg` to **all** screens in the stack (not just the top), so every screen is always correctly sized when it becomes active.
- Popping the last screen quits the program (`tea.Quit`).

Every screen implements the `screen` interface:

```go
type screen interface {
    tea.Model              // Init() / Update() / View()
    setSize(w, h int) screen
}
```

`setSize` must return a copy with `width` and `height` updated — screens are value types (structs), not pointers.

### Adding a new screen

1. Create a struct in `internal/ui/` that implements `screen`.
2. Load data asynchronously: return a `tea.Cmd` from `Init()` that fetches data and returns a typed message (e.g. `type myDataMsg struct{...}`). Handle it in `Update`.
3. Push it from another screen's `Update` by emitting `pushMsg{newMyScreen(...)}`.
4. Pop back with `popMsg{}` on `Esc` / `Backspace` / `q`.
5. Never call `tea.Quit` directly inside a screen except for an explicit global quit key (`q`).

### Colour palette

All colours are defined as package-level `lipgloss.Style` vars at the top of `app.go` and reused across screens (same package). Do **not** define new colours inline — add them to the shared var block.

| Token | Hex | Semantic use |
|---|---|---|
| `colorPrimary` | `#7D56F4` | Border highlights, titles, selected-item background |
| `colorDim` | `#6272A4` | Secondary text, separators, status bars, empty states |
| `colorWhite` | `#F8F8F2` | Primary content text |
| `colorGreen` | `#50FA7B` | Success, `view`/`pure` mutability, constructor badge, ABI present |
| `colorRed` | `#FF5555` | Error states, error-type badge, ABI absent |
| `colorYellow` | `#F9D449` | Warnings, `payable` mutability, event badge |
| `colorBlue` | `#4BAFED` | Function badge, selector values |

The `TablePrinter` in `pkg/abiparser/print.go` has its own mirrored set of `table*Style` vars (same palette, slightly extended with purple `#BD93F9` for selectors in the list view).

### Layout conventions

- **Borders**: always `lipgloss.RoundedBorder()` with `BorderForeground(colorPrimary)`.
- **Full-screen screens** (contract list, browse): use `lipgloss.NewStyle().Width(innerW).Height(h-2).Render(content)` where `innerW = w - 2` (border chars). This constrains the box exactly to the terminal.
- **Centred overlay screens** (home, download): use `lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)`.
- **Split panes** (browse screen): left pane = 38% of inner width, right pane = remainder minus 3 (for `" │ "` separator). Both panes padded to exactly `visibleRows` lines so the separator stays straight.
- **Narrow fallback** (< 80 cols): switch to single-column layout. The browse screen detects `w < 80` in `View()` and calls `renderNarrow` instead of `renderSplit`.
- **Status bar**: always the last line inside the box, using `dimStyle`.

### Column sizing with Lip Gloss

Plain `fmt.Fprintf("%-*s", w, s)` breaks when `s` contains ANSI escape codes because `len(s)` counts bytes, not visible characters.

**Always use lipgloss width-constrained rendering for table cells:**

```go
cell := func(styled string, colW int) string {
    return lipgloss.NewStyle().Width(colW + gap).Render(styled)
}
```

Measure column widths from **raw (unstyled)** strings first, then apply styles only when rendering.

### Keyboard conventions

Every screen must support these bindings consistently:

| Key | Action |
|---|---|
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `Enter` | Select / drill into next screen |
| `Esc` / `Backspace` | Go back (pop screen) |
| `/` | Focus filter input |
| `Esc` (while filtering) | Clear filter and blur input |
| `q` | Quit entire program |
| `Ctrl+C` | Quit (handled at `appModel` level) |

The status bar at the bottom of each screen must display the applicable subset of these bindings as a reminder.

### Async data loading pattern

```go
func (m myScreen) Init() tea.Cmd {
    return loadDataCmd(m.someParam)
}

func loadDataCmd(param string) tea.Cmd {
    return func() tea.Msg {
        data, err := expensiveLoad(param)
        if err != nil {
            return loadErrMsg{err}
        }
        return loadedMsg{data}
    }
}

func (m myScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case loadedMsg:
        m.data = msg.data
        m.loaded = true
        return m, nil
    case loadErrMsg:
        m.err = msg.err
        m.loaded = true
        return m, nil
    }
    // ...
}
```

Show a `dimStyle.Render("  Loading...")` placeholder in `View()` while `!m.loaded`.

### Element type badges

Use the shared badge styles defined in `app.go` for consistent type labelling across all screens:

| Type | Badge | Style |
|---|---|---|
| `function` | `[fn]` | `badgeFunction` — bold blue |
| `event` | `[ev]` | `badgeEvent` — bold yellow |
| `error` | `[er]` | `badgeError` — bold red |
| `constructor` | `[co]` | `badgeConstructor` — bold green |
| `fallback` / `receive` | `[fb]` | `badgeFallback` — dim |

These are exported via the `elementBadge(el abiparser.Element) string` helper in `browse.go`.

### Table output (non-TUI)

The `TablePrinter` in `pkg/abiparser/print.go` also uses Lip Gloss for the static `abi view --output table` command. Follow the same column-sizing pattern and the same colour palette. The `table*Style` vars there mirror the TUI palette — keep them in sync if colours change.

## Running / building

```bash
go build -o abitool .
go test ./...
```

Before submitting changes, run the pre-checks:

```bash
make lint   # requires golangci-lint
make test
```
