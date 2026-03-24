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

### `abitool rpc call` — Read-only contract calls (`eth_call`)

Call a view/pure contract function and display the decoded result.

**Motivation:** The most common interaction with a smart contract is a read. This should require no wallet, no gas estimation, and no signing.

**Scope:**
- Accept a contract address, function name, and arguments
- Encode calldata (via encode feature above)
- Send `eth_call` via JSON-RPC
- Decode and display the return value

**Configuration additions needed:**
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

### `outputs` field in stored ABIs

Currently only `inputs` are used. Add `outputs` to the `Element` struct and storage so that return values from `eth_call` can be decoded.

---

### Interactive mode / TUI

A terminal UI (e.g. using [bubbletea](https://github.com/charmbracelet/bubbletea)) for browsing stored contracts and exploring their interfaces without memorising addresses.
