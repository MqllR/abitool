// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package chains_test

import (
	"testing"

	"github.com/MqllR/abitool/pkg/chains"
)

func TestName_KnownChains(t *testing.T) {
	cases := []struct {
		id   int
		want string
	}{
		{1, "Ethereum Mainnet"},
		{137, "Polygon Mainnet"},
		{10, "Optimism Mainnet"},
		{56, "BNB Chain"},
		{8453, "Base Mainnet"},
		{42161, "Arbitrum One Mainnet"},
		{43114, "Avalanche Mainnet"},
		{11155111, "Ethereum Sepolia"},
	}
	for _, tc := range cases {
		got := chains.Name(tc.id)
		if got != tc.want {
			t.Errorf("Name(%d): got %q, want %q", tc.id, got, tc.want)
		}
	}
}

func TestName_UnknownChain(t *testing.T) {
	cases := []struct {
		id   int
		want string
	}{
		{99999, "Chain 99999"},
		{0, "Chain 0"},
		{-1, "Chain -1"},
		{999999999, "Chain 999999999"},
	}
	for _, tc := range cases {
		got := chains.Name(tc.id)
		if got != tc.want {
			t.Errorf("Name(%d): got %q, want %q", tc.id, got, tc.want)
		}
	}
}

func TestKnown_SpotCheck(t *testing.T) {
	checks := []struct {
		id          int
		wantName    string
		wantRPCHas  string
	}{
		{1, "Ethereum Mainnet", "ethereum"},
		{137, "Polygon Mainnet", "polygon"},
		{8453, "Base Mainnet", "base.org"},
	}
	for _, tc := range checks {
		info, ok := chains.Known[tc.id]
		if !ok {
			t.Errorf("Known[%d]: not found", tc.id)
			continue
		}
		if info.Name != tc.wantName {
			t.Errorf("Known[%d].Name: got %q, want %q", tc.id, info.Name, tc.wantName)
		}
		if info.DefaultRPCURL == "" {
			t.Errorf("Known[%d].DefaultRPCURL: empty", tc.id)
		}
	}
}

func TestKnown_HasExpectedEntries(t *testing.T) {
	expectedIDs := []int{1, 10, 56, 137, 8453, 42161, 43114, 11155111}
	for _, id := range expectedIDs {
		if _, ok := chains.Known[id]; !ok {
			t.Errorf("Known map missing chain ID %d", id)
		}
	}
}
