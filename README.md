# abitool

A CLI tool that provides a simple human-friendly interface to Ethereum smart contracts — pulling ABIs from Etherscan, inspecting functions and events, and (coming soon) calling contract functions.

## Features

- **Interactive TUI** — full-screen dashboard to browse, download and explore stored ABIs
- **Download** — fetch and cache a contract's ABI from Etherscan by address
- **View** — inspect functions, events, constructors and their selectors in coloured JSON or table format
- **List** — show all locally stored contracts
- **Delete** — remove a stored ABI and its metadata

## Installation

```bash
go install github.com/MqllR/abitool@latest
```

## Development

Build from source and run pre-checks before submitting changes:

```bash
make build  # compile the binary
make test   # run all tests
make lint   # run golangci-lint (requires golangci-lint to be installed)
```

Or build from source:

```bash
git clone https://github.com/MqllR/abitool.git
cd abitool
go build -o abitool .
```

## Configuration

abitool expects a YAML config file at `$HOME/.config/abitool/config.yaml` by default:

```yaml
etherscan:
  api_key: "YOUR_ETHERSCAN_API_KEY"
```

Get a free API key at [etherscan.io/apis](https://etherscan.io/apis).

You can point to a different file with the `-f` flag:

```bash
abitool -f /path/to/config.yaml abi list
```

## Usage

### Interactive TUI

Running `abitool` with no arguments opens the interactive dashboard:

```bash
abitool
```

The TUI provides a full-screen terminal interface:

```
╭─────────────────────────────────────╮
│  ⬡  Ethereum ABI Tool               │
│                                     │
│    ❯  Contracts   3 stored          │
│       Download    fetch a new ABI   │
│                                     │
│  ↑↓/jk navigate  enter select  q quit
╰─────────────────────────────────────╯
```

**Navigation:**

| Key | Action |
|---|---|
| `↑` / `↓` or `j` / `k` | Move selection |
| `Enter` | Select / drill in |
| `Esc` or `Backspace` | Go back |
| `/` | Filter the current list |
| `q` | Quit |

**Screens:**

1. **Home** — choose between browsing stored contracts or downloading a new one
2. **Contracts** — filterable list of all stored ABIs; press `Enter` to explore any contract
3. **ABI Browse** — split-pane view: element list on the left (color-coded by type), detail panel on the right showing selector, mutability, and parameter types
4. **Download** — enter a contract address to fetch its ABI from Etherscan on the spot

> **Note:** The Download screen requires `etherscan.api_key` to be set in your config file.

### Download a contract ABI

```bash
abitool abi download 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48
```

ABIs are stored locally under `$HOME/.config/abitool/abis/<chainid>/`. Subsequent downloads of the same address are skipped.

### View a contract ABI

```bash
# Pretty-printed JSON (default)
abitool abi view 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48

# Coloured table with selectors
abitool abi view -o table 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48

# Table with parameter names visible
abitool abi view -o table --with-input-name 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48

# Filter to functions only
abitool abi view -o table -t function 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48

# Available type filters: all, function, event, constructor, fallback, receive
```

The table output uses colour to aid readability:

| Column | Colour |
|---|---|
| `function` type | Blue |
| `event` type | Yellow |
| `error` type | Red |
| `constructor` type | Green |
| Selector | Purple |
| `view` / `pure` mutability | Green |
| `payable` mutability | Yellow |

### List all stored contracts

```bash
abitool abi list
```

Outputs a coloured table with address, contract name, and ABI presence indicator.

### Delete a stored contract

```bash
abitool abi delete 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48
```

### Use a different chain

abitool defaults to Ethereum Mainnet (chain ID 1). Use `-c` to specify another supported chain:

```bash
abitool abi -c 1 download 0x...
```

## Flags reference

| Flag | Default | Description |
|---|---|---|
| `-f, --config` | `$HOME/.config/abitool/config.yaml` | Config file path |
| `-c, --chainid` | `1` | Chain ID |
| `-s, --abi-store` | `$HOME/.config/abitool/abis/` | ABI storage directory |
| `-o, --output` | `json` | View output format: `json` or `table` |
| `-t, --type` | `all` | View type filter: `all`, `function`, `event`, `constructor`, `fallback`, `receive` |
| `--with-input-name` | `false` | Show parameter names in table output |

## Storage layout

```
$HOME/.config/abitool/
├── config.yaml
└── abis/
    └── 1/                    # chain ID
        ├── contracts.json    # metadata index
        └── <address>         # raw ABI JSON per contract
```

## Roadmap

See [docs/ROADMAP.md](docs/ROADMAP.md) for planned features including ABI encoding/decoding and direct RPC calls.
