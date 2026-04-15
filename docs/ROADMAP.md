# Roadmap

This document tracks planned features, their motivation, and rough implementation notes.

---

## Planned features

### `abitool rpc call` — Read-only contract calls (`eth_call`) ✅ Done

Implemented in `internal/contract/call.go`, `pkg/ethclient/`, `pkg/abicodec/`, `cmd/rpc/call.go`.

- Interactive argument prompting via bubbletea TUI form (`internal/ui/form.go`)
- Full ABI type support: scalars, `address`, `bool`, `uint/int`, `bytes`, fixed bytes, dynamic arrays, fixed-size arrays, tuples
- Named output display (parameter names shown alongside decoded values)

**Configuration:**

```yaml
rpc:
  url: "https://mainnet.infura.io/v3/YOUR_KEY"
```

---

### `abitool encode` — ABI calldata encoding ✅ Done

Implemented as a top-level command in `cmd/encode.go` and `internal/contract/encode.go`.

**Usage:**

```
abitool encode <address> <function> [arg...]
```

- Accepts a stored contract address + function name + argument list
- Resolves the function from the stored ABI
- ABI-encodes the arguments (including tuple/struct and array support via go-ethereum codec)
- Outputs the hex calldata to stdout by default; `--output json` includes signature and selector

**TUI integration:**

Pressing `Enter` or `c` on any function in the ABI browser now opens an action menu:
- **view/pure functions:** ① Call (eth_call) ② Generate calldata
- **nonpayable/payable functions:** ① Generate calldata

"Generate calldata" pushes an `encodeFormScreen` (argument input) followed by an `encodeResultScreen` (calldata display with signature and selector).

---

### `abitool decode` — ABI calldata / return value decoding ✅ Done

Implemented as a top-level command in `cmd/decode.go` and `internal/contract/decode.go`.

**Modes:**

- `abitool decode <address> <calldata-hex>` — raw calldata (auto-detects function from 4-byte selector)
- `abitool decode --eth-call <json>` — parse an `eth_call` JSON request body
- `abitool decode --from-tx <hex>` — parse a RLP-encoded signed transaction
- `abitool decode --return-data <address> <function-name> <hex>` — decode return data

**Implementation notes:**

- `pkg/abicodec.DecodeInput` strips the 4-byte selector and unpacks inputs via go-ethereum.
- `internal/contract.DecodeManager` handles all four modes and provides clear error messages with suggested fix when the ABI is not in the local store.
- Raw transaction parsing uses `go-ethereum/core/types.Transaction.UnmarshalBinary` (supports legacy + EIP-1559 + EIP-4844 transactions).

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

### Persistent chain selection ✅ Done

Users can now persist their default chain ID without editing the config file manually.

- **CLI:** `abitool chain use <chainID>` — sets and saves the default chain.
- **TUI:** "Switch Chain" from the home screen — chain selection is silently persisted to the config file.

The config write-back marshals only the `Config` struct (not all viper keys) to avoid flushing transient CLI flags. See `internal/abitool/app.go` (`SaveChainID`, `saveConfig`).

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
