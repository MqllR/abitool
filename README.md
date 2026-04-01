# abitool

> A human-friendly CLI for Ethereum smart contracts — browse ABIs, inspect selectors, call read-only functions, all from your terminal.

[![Go](https://img.shields.io/badge/Go-1.25-blue)](https://go.dev) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## Features

- 🖥️ **Interactive TUI** — full-screen dashboard to browse, download and explore stored ABIs
- ⬇️ **Download** — fetch and cache contract ABIs from Etherscan by address
- 📋 **View** — inspect functions, events, constructors and their 4-byte selectors / topic hashes in coloured JSON or table format
- 📥 **Import** — load a local ABI JSON file without hitting Etherscan
- 📞 **RPC Call** — call read-only contract functions directly from the CLI (or via interactive TUI form)
- 🗂️ **Multi-chain** — Ethereum, Optimism, Base, Arbitrum, Polygon, BNB Chain, Avalanche, Linea, and more
- 💾 **Local storage** — ABIs cached per chain ID; no repeated downloads

---

## Quick Start

**1. Install**

```bash
go install github.com/MqllR/abitool@latest
```

**2. Get a free Etherscan API key**

Sign up at [etherscan.io/apis](https://etherscan.io/apis) — the free tier is sufficient.

**3. Create config**

```bash
mkdir -p ~/.config/abitool
cat > ~/.config/abitool/config.yaml << EOF
etherscan:
  api_key: "YOUR_ETHERSCAN_API_KEY"
EOF
```

**4. Download a contract ABI** (USDC on Ethereum mainnet)

```bash
abitool abi download 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48
```

**5. Explore it**

```bash
# Open interactive TUI
abitool

# Or view in the terminal directly
abitool abi view -o table 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48
```

---

## 🖥️ Interactive TUI

Run `abitool` with no arguments to launch the full-screen dashboard:

```bash
abitool
```

### Screens

| Screen | Description |
|---|---|
| **Home** | Choose between browsing stored contracts or downloading a new one |
| **Contracts** | Filterable list of all cached ABIs — press `Enter` to explore any contract |
| **ABI Browse** | Split-pane view: element list on the left (colour-coded by type) + detail panel on the right showing selector, mutability, and parameter types |
| **Download** | Enter a contract address to fetch its ABI from Etherscan on the spot |

### Navigation

| Key | Action |
|---|---|
| `↑` / `↓` or `j` / `k` | Move selection |
| `Enter` | Select / drill in |
| `Esc` / `Backspace` | Go back |
| `/` | Focus filter input |
| `Esc` (while filtering) | Clear filter |
| `q` | Quit |

### Terminal mock-up

```
╭──────────────────────────────────────────────────────╮
│  ⬡  Ethereum ABI Tool                                │
│                                                      │
│    ❯  Contracts   3 stored                           │
│       Download    fetch a new ABI                    │
│                                                      │
│  ↑↓/jk navigate · enter select · q quit             │
╰──────────────────────────────────────────────────────╯

╭── Contracts ──────────────────────────────────────────╮
│  / filter...                                          │
│  ❯ 0xA0b8…eB48  FiatTokenV2_1              ✓ ABI    │
│    0xdAC1…1ec7  UniswapV2Router02           ✓ ABI    │
│    0x1f98…0505  UniswapV3Factory            ✓ ABI    │
│                                                      │
│  enter browse · / filter · backspace back · q quit   │
╰──────────────────────────────────────────────────────╯

╭── FiatTokenV2_1 ──────────────┬─ Detail ─────────────╮
│  [fn] transfer          view  │  Selector             │
│  [fn] transferFrom      view  │  0xa9059cbb           │
│  [fn] approve           view  │                       │
│  [fn] balanceOf         view  │  Signature            │
│  [ev] Transfer               │  transfer(address,    │
│  [ev] Approval               │    uint256)           │
│                               │                       │
│  ↑↓ move · / filter · esc back│  Inputs               │
│                               │  to      address      │
│                               │  amount  uint256      │
╰───────────────────────────────┴───────────────────────╯
```

---

## CLI Reference

### Commands

| Command | Description |
|---|---|
| `abitool` | Launch interactive TUI |
| `abitool abi download <address>` | Download ABI from Etherscan and store locally |
| `abitool abi view <address>` | Print ABI to stdout (JSON or table) |
| `abitool abi list` | List all stored contracts |
| `abitool abi delete <address>` | Delete a stored ABI |
| `abitool abi import <address> <file>` | Import a local ABI JSON file |
| `abitool rpc call <address> <function> [args...]` | Call a read-only contract function via RPC |

### Examples

```bash
# Download USDC ABI on Ethereum mainnet (chain 1, default)
abitool abi download 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48

# Download on Base (chain 8453)
abitool abi download -c 8453 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913

# View as coloured table with selectors
abitool abi view -o table 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48

# Show parameter names too
abitool abi view -o table --with-input-name 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48

# Filter to events only
abitool abi view -o table -t event 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48

# Import a local ABI file
abitool abi import 0xdeadbeef... ./MyContract.abi.json --name MyContract

# Call a read-only function (balanceOf on USDC)
abitool rpc call --rpc-url https://eth.llamarpc.com \
  0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 \
  "balanceOf(address)" \
  0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045

# Interactive RPC call form
abitool rpc call --interactive \
  --rpc-url https://eth.llamarpc.com \
  0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 \
  "balanceOf(address)"
```

---

## Flags Reference

### Global flags

| Flag | Default | Description |
|---|---|---|
| `-f, --config` | `~/.config/abitool/config.yaml` | Config file path |
| `-c, --chainid` | `1` | Chain ID |
| `-s, --abi-store` | `~/.config/abitool/abis/` | ABI storage directory |

### `abi view` flags

| Flag | Default | Description |
|---|---|---|
| `-o, --output` | `json` | Output format: `json` or `table` |
| `-t, --type` | `all` | Filter by type: `all`, `function`, `event`, `constructor`, `fallback`, `receive` |
| `--with-input-name` | `false` | Show parameter names in table output |

### `rpc` flags

| Flag | Description |
|---|---|
| `--rpc-url` | RPC endpoint URL |
| `--interactive` | Use interactive TUI form for arguments |

### `abi import` flags

| Flag | Description |
|---|---|
| `--name` | Contract name to store in metadata |
| `--force` | Overwrite an existing stored ABI |

---

## Supported Chains

| Chain ID | Network |
|---|---|
| `1` | Ethereum Mainnet |
| `10` | Optimism |
| `56` | BNB Smart Chain |
| `137` | Polygon |
| `8453` | Base |
| `42161` | Arbitrum One |
| `43114` | Avalanche C-Chain |
| `59144` | Linea |
| `11155111` | Sepolia (testnet) |
| `84532` | Base Sepolia (testnet) |
| `421614` | Arbitrum Sepolia (testnet) |
| `11155420` | Optimism Sepolia (testnet) |
| `80002` | Polygon Amoy (testnet) |

Pass the chain ID with `-c`:

```bash
abitool abi -c 8453 download 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913
```

---

## Configuration

abitool looks for a YAML config at `$HOME/.config/abitool/config.yaml` by default. Override with `-f`:

```bash
abitool -f /path/to/config.yaml abi list
```

**Full config reference:**

```yaml
etherscan:
  api_key: "YOUR_ETHERSCAN_API_KEY"   # required for download / TUI download screen
```

---

## Storage Layout

ABIs are stored as plain files on disk — one JSON file per contract, plus a `contracts.json` metadata index per chain.

```
$HOME/.config/abitool/
├── config.yaml
└── abis/
    ├── 1/                              # Ethereum mainnet
    │   ├── contracts.json              # metadata index (name, path, …)
    │   └── 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48   # raw ABI JSON
    └── 8453/                           # Base
        ├── contracts.json
        └── 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913
```

---

## Development

### Prerequisites

- Go 1.25+
- [`golangci-lint`](https://golangci-lint.run/usage/install/) for linting

### Build & test

```bash
git clone https://github.com/MqllR/abitool.git
cd abitool

go build -o abitool .   # compile

make test               # go test ./...
make lint               # golangci-lint run
```

### Etherscan API key for tests

Some tests that hit the Etherscan API read the key from the standard config file. Set it up once:

```bash
mkdir -p ~/.config/abitool
echo 'etherscan:\n  api_key: "YOUR_KEY"' > ~/.config/abitool/config.yaml
```

Get a free key at [etherscan.io/apis](https://etherscan.io/apis).

### Tech stack

| Concern | Library |
|---|---|
| CLI framework | [cobra](https://github.com/spf13/cobra) |
| Config | [viper](https://github.com/spf13/viper) |
| TUI | [bubbletea](https://github.com/charmbracelet/bubbletea) + [bubbles](https://github.com/charmbracelet/bubbles) |
| TUI styling | [lipgloss](https://github.com/charmbracelet/lipgloss) |
| Ethereum | [go-ethereum](https://github.com/ethereum/go-ethereum) (keccak, ABI codec, RPC) |

---

## License

MIT — see [LICENSE](LICENSE) for details.
