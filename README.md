# abitool

> A human-friendly CLI for Ethereum smart contracts — browse ABIs, inspect selectors, call read-only functions, all from your terminal.

[![Go](https://img.shields.io/badge/Go-1.25-blue)](https://go.dev) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## Features

- 🖥️ **Interactive TUI** — full-screen dashboard to browse, download and explore stored ABIs
- ⬇️ **Download** — fetch and cache contract ABIs from Etherscan by address
- 🏷️ **Labels** — give contracts human-friendly names to avoid confusion with identically-named proxies
- 📋 **View** — inspect functions, events, constructors and their 4-byte selectors / topic hashes in coloured JSON or table format
- 📥 **Import** — load a local ABI JSON file without hitting Etherscan
- 📞 **RPC Call** — call read-only contract functions directly from the CLI (or via interactive TUI form)
- 🔍 **Decode** — decode raw calldata, `eth_call` request bodies, signed transactions, or return data into human-readable form
- 🔏 **Encode** — ABI-encode a function call into calldata hex (useful for multisigs, scripting, and debugging)
- 🗂️ **Multi-chain** — Ethereum, Optimism, Base, Arbitrum, Polygon, BNB Chain, Avalanche, Linea, and more
- 💾 **Local storage** — ABIs cached per chain ID; no repeated downloads

---

## Quick Start

**1. Install**

```bash
go install github.com/MqllR/abitool/cmd@latest
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
| **Home** | Choose between browsing stored contracts, downloading a new one, or switching the active chain |
| **Contracts** | Filterable list of all cached ABIs — press `Enter` to explore any contract |
| **ABI Browse** | Split-pane view: element list on the left (colour-coded by type) + detail panel on the right showing selector, mutability, and parameter types |
| **Download** | Two-step form: enter a contract address, then optionally set a label — fetches the ABI from Etherscan on confirmation |
| **Chain Selector** | Pick from the known chain list or enter a custom chain ID; selection is persisted to the config file |


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
| `abitool chain use <chainID>` | Set the default chain ID (persisted to config) |
| `abitool abi download <address>` | Download ABI from Etherscan and store locally |
| `abitool abi view <address>` | Print ABI to stdout (JSON or table) |
| `abitool abi list` | List all stored contracts |
| `abitool abi rename <address> <label>` | Set or update the display label for a contract |
| `abitool abi delete <address>` | Delete a stored ABI |
| `abitool abi import <address> <file>` | Import a local ABI JSON file |
| `abitool rpc call <address> <function> [args...]` | Call a read-only contract function via RPC |
| `abitool decode <address> <calldata-hex>` | Decode raw calldata using the stored ABI |
| `abitool decode --eth-call <json>` | Decode an `eth_call` JSON request body |
| `abitool decode --from-tx <hex>` | Decode calldata from a signed transaction |
| `abitool decode --return-data <address> <fn> <hex>` | Decode return data from a function call |
| `abitool encode <address> <function> [args...]` | ABI-encode a function call into calldata hex |

### Examples

```bash
# Set Base as the default chain (persisted to config)
abitool chain use 8453

# Download USDC ABI on Ethereum mainnet (chain 1, default)
abitool abi download 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48

# Download on Base (chain 8453)
abitool abi --chainid 8453 download 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913

# Download and immediately set a label (avoids ambiguous proxy names)
abitool abi download -c 84532 --label "USDC Base Sepolia" 0x036CbD53842c5426634e7929541eC2318f3dCF7e

# Rename a contract you already downloaded
abitool abi rename 0x036CbD53842c5426634e7929541eC2318f3dCF7e "USDC Base Sepolia"

# List contracts — duplicate display names are flagged with ⚠; labels shown with [EtherscanName]
abitool abi list --chainid 84532
# Chain: Base Sepolia (84532)
#
# Address                                     Contract Name                           ABI
# ─────────────────────────────────────────────────────────────────────────────────────────
# 0x036CbD53842c5426634e7929541eC2318f3dCF7e  USDC Base Sepolia [FiatTokenProxy]      true
# 0x808456652fdb597867f38412077A9182bf77359F  ⚠ FiatTokenProxy                        true
# 0xd74cc5d436923b8ba2c179b4bCA2841D8A52C5B5  FiatTokenV2_2                           true

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

# Decode raw calldata (USDC transfer)
abitool decode 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 \
  0xa9059cbb000000000000000000000000d8da6bf26964af9d7eed9e03e53415d37aa960450000000000000000000000000000000000000000000000000000000005f5e100

# Decode an eth_call JSON body
abitool decode --eth-call '{"to":"0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48","data":"0xa9059cbb..."}'

# Decode a signed transaction
abitool decode --from-tx 0x02f8...

# Decode return data from a function
abitool decode --return-data 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 totalSupply \
  0x0000000000000000000000000000000000000000000000000000000005f5e100

# Encode calldata for a state-changing function (e.g. for a multisig proposal)
abitool encode 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 transfer \
  0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045 1000000

# Encode with JSON output (includes signature and selector)
abitool encode --output json 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 transfer \
  0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045 1000000
```

---

## Flags Reference

### Global flags

| Flag | Default | Description |
|---|---|---|
| `-c, --config` | `~/.config/abitool/config.yaml` | Config file path |
| `--chainid` | `1` (or value from config) | Chain ID for this invocation |
| `-s, --abi-store` | `~/.config/abitool/abis/` | ABI storage directory |

> **Tip:** Use `abitool chain use <id>` to persist a default chain so you don't have to pass `--chainid` every time.
### `abi download` flags

| Flag | Description |
|---|---|
| `--label` | Optional human-friendly display name (overrides the Etherscan contract name in list output) |

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

### `decode` flags

| Flag | Default | Description |
|---|---|---|
| `--eth-call` | | Parse an `eth_call` JSON body `{"to":"0x...","data":"0x..."}` |
| `--from-tx` | | Parse a RLP-encoded signed transaction hex |
| `--return-data` | | Decode return data; positional args become `<address> <function-name> <return-hex>` |
| `-o, --output` | `text` | Output format: `text` or `json` |

### `encode` flags

| Flag | Default | Description |
|---|---|---|
| `-o, --output` | `hex` | Output format: `hex` (bare calldata string) or `json` (includes signature and selector) |

---

## Contract Labels

Etherscan names are often misleading — many proxy contracts share the same name (e.g. `FiatTokenProxy`), making it hard to tell them apart in a large library.

**Labels** let you assign a human-friendly name to any stored contract. The original Etherscan name is preserved and shown in brackets when a label is set.

```
Address                                     Contract Name                           ABI
─────────────────────────────────────────────────────────────────────────────────────────
0x036CbD53842c5426634e7929541eC2318f3dCF7e  USDC Base Sepolia [FiatTokenProxy]      true
0x808456652fdb597867f38412077A9182bf77359F  ⚠ FiatTokenProxy                        true
0xd74cc5d436923b8ba2c179b4bCA2841D8A52C5B5  FiatTokenV2_2                           true
```

- Rows whose display name is still ambiguous (no label, same name as another contract) are flagged with **⚠**.
- Set a label at download time with `--label`, or update it any time with `abi rename`.
- Labels are stored in `contracts.json` under the `label` key and have no effect on the stored ABI itself.

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

Pass the chain ID with `--chainid`, or set it permanently with `abitool chain use`:

```bash
# One-off override
abitool abi --chainid 8453 download 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913

# Or set Base as the default so you never need the flag again
abitool chain use 8453
abitool abi download 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913
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
chainid: 137                           # optional — persisted by `abitool chain use`
rpc:
  url: "https://mainnet.infura.io/v3/YOUR_KEY"  # optional RPC fallback
```

`chainid` and `rpc.url` are written back to the config file automatically by `abitool chain use` and can also be edited by hand.

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

go build -o abitool ./cmd/   # compile

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
