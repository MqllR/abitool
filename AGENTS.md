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

## Repository layout

```
cmd/                   Cobra command definitions (entry points only — no business logic)
  abi.go               Parent "abi" command; registers persistent flags (chainid, abi-store)
  abi/
    download.go        abitool abi download <address>
    view.go            abitool abi view <address>
    list.go            abitool abi list
    delete.go          abitool abi delete <address>
  root.go              Root command; loads config via PersistentPreRun

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

pkg/
  abiparser/
    types.go           ABI element types (FunctionType, EventType, …)
    parser.go          ParseABI(), Element.Signature(), Element.Selector()
    print.go           PrettyPrinter (JSON) and TablePrinter with options
  etherscan/
    types.go           ContractSourceCodeResponse
    client.go          HTTP client; call() helper
    contract.go        GetABI(), GetSourceCode()
  storage/
    abi/local.go       File-per-contract ABI storage (Write/Read/Delete/GetPath)
    contract/local.go  contracts.json index (Add/Get/Delete/List); ErrNotFound/ErrAlreadyExists
```

## Key design decisions

- **`internal/` vs `pkg/`** — `internal/` contains app-specific orchestration (ABIManager, config). `pkg/` contains reusable, app-agnostic packages (abiparser, etherscan client, storage backends).
- **Storage** — ABIs are stored as raw JSON files named after their contract address. A `contracts.json` index file holds metadata (contract name, ABI file path). Both live under `<abi-store>/<chainid>/`.
- **Config** — Loaded once via `sync.Once` in `internal/abitool/app.go`, then accessed globally through `ConfigInstance()`. Cobra binds CLI flags into Viper; config file values are the fallback.
- **No external ABI codec** — ABI encoding/decoding is intentionally not implemented yet. Function selectors are computed in-house with Keccak-256 over the canonical signature string.

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

1. **Tuple types in signatures** — `Signature()` emits `tuple` instead of expanding components, producing wrong selectors for structs. (`pkg/abiparser/parser.go`)
2. **`sync.Once` error not persisted** — a failed `Load()` call cannot be retried, but subsequent calls return `nil`. (`internal/abitool/app.go`)
3. **Etherscan API-level errors ignored** — only HTTP status is checked; `status:"0"` in JSON body is not validated. (`pkg/etherscan/client.go`)
4. **URL injection** — contract addresses are concatenated raw into query strings. (`pkg/etherscan/contract.go`)
5. **Orphaned ABI files** — if metadata save fails after ABI file is written, no rollback occurs. (`internal/contract/storage.go`)
6. **`DeleteWithABI` error comparison** — uses `!=` on a wrapped error; should use `errors.Is`. (`internal/contract/abi.go`)
7. **Implicit `return c, err`** — final return in `getContracts` should be `return c, nil`. (`pkg/storage/contract/local.go`)

## Adding a new command

1. Create `cmd/abi/<name>.go` with a `cobra.Command` exported as `<Name>Cmd`.
2. Register it in `cmd/abi.go` with `abiCmd.AddCommand(abi.<Name>Cmd)`.
3. Business logic goes in `internal/contract/` (new method on `ABIManager` if it involves stored contracts).
4. New Etherscan endpoints go in `pkg/etherscan/`.

## Adding a new chain

Add the chain ID to `SupportedChainIDs` in `internal/contract/config.go`. The Etherscan v2 API already supports multiple chains via the `chainid` query parameter.

## Running / building

```bash
go build -o abitool .
go test ./...
```
