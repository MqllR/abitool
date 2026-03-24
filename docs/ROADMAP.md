# Roadmap

This document tracks planned features, their motivation, and rough implementation notes.

---

## In progress / next up

### Bug fixes (code review)

The following correctness issues are queued for fixing before new features land:

- Tuple type expansion in function signatures (`pkg/abiparser/parser.go`)
- Etherscan API-level error handling (`pkg/etherscan/client.go`)
- URL injection protection (`pkg/etherscan/contract.go`)
- Orphaned ABI file rollback (`internal/contract/storage.go`)
- `DeleteWithABI` error comparison (`internal/contract/abi.go`)
- `sync.Once` error persistence (`internal/abitool/app.go`)

See `AGENTS.md` for details on each issue.

### `abitool rpc call` — Read-only contract calls (`eth_call`) ⚙️ In progress

Implemented in `internal/contract/call.go`, `pkg/ethclient/`, `pkg/abicodec/`, `cmd/rpc/call.go`.

Interactive argument prompting uses a bubbletea TUI form (`internal/ui/form.go`).

Remaining work:
- Array / tuple argument support in `pkg/abicodec/codec.go`
- Named output display (show output parameter names alongside values)

### `outputs` field in stored ABIs ✅ Done

`Outputs []Output` added to `abiparser.Element`. Return values from `eth_call` are now decoded using stored output types.

---

## Planned features

### `abitool abi encode` — ABI calldata encoding

Encode a function call into EVM-compatible calldata (ABI-encoded hex).

**Motivation:** Needed as a prerequisite for transaction sending. Useful standalone for debugging or building raw calldata by hand.

**Scope:**
- Accept a stored contract address + function name + argument list
- Resolve the function from the stored ABI
- ABI-encode the arguments (including tuple/struct support)
- Output the hex calldata to stdout

**Implementation notes:**
- The ABI encoding spec is defined in the [Solidity ABI documentation](https://docs.soliditylabs.io/docs/abi-spec).
- Consider using `github.com/ethereum/go-ethereum/accounts/abi` to avoid reimplementing the codec.
- Must handle all basic types (`uint<M>`, `int<M>`, `bytes<M>`, `bool`, `address`) and dynamic types (`bytes`, `string`, arrays, tuples).

---

### `abitool abi decode` — ABI calldata / return value decoding

Decode raw hex calldata or return data back into human-readable form.

**Motivation:** When inspecting transactions or RPC responses, raw calldata is unreadable without the ABI. This command makes it easy to decode without external tools.

**Scope:**
- Decode transaction input data given a contract address and the stored ABI (auto-detect function from 4-byte selector)
- Decode return data from `eth_call` responses

**Implementation notes:**
- Selector lookup: match the first 4 bytes of the input against all computed selectors for the stored ABI.
- Return data decoding requires knowing the output types; these are not stored today — the `Element` struct and ABI storage will need to include `outputs`.

---

### `abitool rpc call` — Read-only contract calls (`eth_call`) ✅ Implemented

See *In progress* section above for current status and remaining work.

**Configuration:**
```yaml
rpc:
  url: "https://mainnet.infura.io/v3/YOUR_KEY"
```

---

### `abitool rpc estimate` — Gas estimation (`eth_estimateGas`)

Estimate the gas cost of a state-changing call without sending a transaction.

**Scope:**
- Accept contract address, function name, arguments, and optionally `from` address
- Build and send `eth_estimateGas` JSON-RPC request
- Display estimated gas units

---

### `abitool rpc send` — Send a signed transaction (`eth_sendRawTransaction`)

Sign and broadcast a state-changing transaction.

**Motivation:** Completes the read/write interaction loop for smart contracts.

**Scope:**
- Accept contract address, function name, arguments, gas limit / gas price overrides
- ABI-encode calldata
- Sign the transaction with a private key (loaded from config or env var — **never a CLI argument**)
- Broadcast via `eth_sendRawTransaction`
- Display the transaction hash

**Security requirements:**
- Private key must never be passed as a CLI argument (would appear in shell history)
- Acceptable sources: `ABITOOL_PRIVATE_KEY` environment variable, or a key file path in config
- Consider hardware wallet / keystore support in a later iteration

**Configuration additions needed:**
```yaml
wallet:
  private_key_env: "ABITOOL_PRIVATE_KEY"  # env var name, not the key itself
```

---

### Multi-chain support

Extend `SupportedChainIDs` in `internal/contract/config.go` to include commonly used networks.

| Chain | ID |
|---|---|
| Ethereum Mainnet | 1 |
| Sepolia | 11155111 |
| Base | 8453 |
| Arbitrum One | 42161 |
| Optimism | 10 |
| Polygon | 137 |

The Etherscan v2 API already supports all these via the `chainid` parameter. RPC URLs will need to be configurable per chain.

---

### `outputs` field in stored ABIs ✅ Done

`Outputs []Output` added to `abiparser.Element` in `pkg/abiparser/parser.go`. Return values from `eth_call` are decoded using the stored output types.

---

### Interactive mode / TUI

A terminal UI (e.g. using [bubbletea](https://github.com/charmbracelet/bubbletea)) for browsing stored contracts and exploring their interfaces without memorising addresses.
