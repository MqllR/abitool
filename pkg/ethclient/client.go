// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

// Package ethclient provides a minimal JSON-RPC client for Ethereum node interactions.
// It uses the go-ethereum RPC transport (github.com/ethereum/go-ethereum/rpc).
// go-ethereum is Copyright 2014 The go-ethereum Authors, licensed under LGPL-3.0.
package ethclient

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	gorpc "github.com/ethereum/go-ethereum/rpc"
)

// Client wraps a go-ethereum JSON-RPC connection.
type Client struct {
	inner *gorpc.Client
}

// Dial connects to the given JSON-RPC endpoint URL.
func Dial(ctx context.Context, url string) (*Client, error) {
	c, err := gorpc.DialContext(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("dialing RPC endpoint %q: %w", url, err)
	}

	return &Client{inner: c}, nil
}

// Close terminates the underlying connection.
func (c *Client) Close() {
	c.inner.Close()
}

// ethCallParams mirrors the JSON object sent as the first argument of eth_call.
type ethCallParams struct {
	To   string `json:"to"`
	Data string `json:"data"`
}

// CallContract executes an eth_call against the given contract address with the
// provided ABI-encoded calldata. block should be "latest", "pending", or a hex
// block number string (e.g. "0x10d4f"). Returns the raw response bytes.
func (c *Client) CallContract(ctx context.Context, to string, calldata []byte, block string) ([]byte, error) {
	params := ethCallParams{
		To:   to,
		Data: hexutil.Encode(calldata),
	}

	// eth_call returns a 0x-prefixed hex string; hexutil.Bytes unmarshals it correctly.
	var result hexutil.Bytes
	if err := c.inner.CallContext(ctx, &result, "eth_call", params, block); err != nil {
		return nil, fmt.Errorf("eth_call failed: %w", err)
	}

	return result, nil
}
