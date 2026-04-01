// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package contract_test

import (
	"encoding/json"
	"errors"
	"testing"

	contractstore "github.com/MqllR/abitool/pkg/storage/contract"
)

type meta struct {
	Name    string `json:"name"`
	ABIPath string `json:"abiPath"`
}

func newLocal(t *testing.T) *contractstore.Local {
	t.Helper()
	store, err := contractstore.NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocal: %v", err)
	}
	return store
}

func metaJSON(t *testing.T, name, abiPath string) []byte {
	t.Helper()
	b, err := json.Marshal(meta{Name: name, ABIPath: abiPath})
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return b
}

func TestAdd_Get_RoundTrip(t *testing.T) {
	store := newLocal(t)
	const addr = "0xabc123"

	if err := store.Add(addr, metaJSON(t, "MyContract", "/path/to/abi")); err != nil {
		t.Fatalf("Add: %v", err)
	}

	raw, err := store.Get(addr)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	var got meta
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if got.Name != "MyContract" {
		t.Errorf("Name: got %q, want %q", got.Name, "MyContract")
	}
	if got.ABIPath != "/path/to/abi" {
		t.Errorf("ABIPath: got %q, want %q", got.ABIPath, "/path/to/abi")
	}
}

func TestAdd_Duplicate(t *testing.T) {
	store := newLocal(t)
	const addr = "0xdup"

	if err := store.Add(addr, metaJSON(t, "Contract", "/abi")); err != nil {
		t.Fatalf("Add (first): %v", err)
	}

	err := store.Add(addr, metaJSON(t, "Contract2", "/abi2"))
	if !errors.Is(err, contractstore.ErrAlreadyExists) {
		t.Errorf("Add duplicate: expected ErrAlreadyExists, got %v", err)
	}
}

func TestGet_NotFound(t *testing.T) {
	store := newLocal(t)
	_, err := store.Get("0xnonexistent")
	if !errors.Is(err, contractstore.ErrNotFound) {
		t.Errorf("Get non-existent: expected ErrNotFound, got %v", err)
	}
}

func TestDelete_Existing(t *testing.T) {
	store := newLocal(t)
	const addr = "0xdel"

	if err := store.Add(addr, metaJSON(t, "Del", "/abi")); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := store.Delete(addr); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := store.Get(addr)
	if !errors.Is(err, contractstore.ErrNotFound) {
		t.Errorf("Get after delete: expected ErrNotFound, got %v", err)
	}
}

func TestDelete_NotFound(t *testing.T) {
	store := newLocal(t)
	err := store.Delete("0xgone")
	if !errors.Is(err, contractstore.ErrNotFound) {
		t.Errorf("Delete non-existent: expected ErrNotFound, got %v", err)
	}
}

func TestList_Empty(t *testing.T) {
	store := newLocal(t)
	seq, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	count := 0
	for range seq {
		count++
	}
	if count != 0 {
		t.Errorf("List empty store: expected 0 items, got %d", count)
	}
}

func TestList_MultipleContracts(t *testing.T) {
	store := newLocal(t)
	addrs := []string{"0xaaa", "0xbbb", "0xccc"}

	for i, addr := range addrs {
		if err := store.Add(addr, metaJSON(t, addr, "/abi")); err != nil {
			t.Fatalf("Add[%d] %q: %v", i, addr, err)
		}
	}

	seq, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	found := map[string]bool{}
	for a := range seq {
		found[a] = true
	}

	if len(found) != len(addrs) {
		t.Errorf("List: expected %d items, got %d", len(addrs), len(found))
	}
	for _, addr := range addrs {
		if !found[addr] {
			t.Errorf("List: address %q not found", addr)
		}
	}
}

func TestGetContracts_MissingFile(t *testing.T) {
	// A brand-new store (no contracts.json) should behave as empty.
	store := newLocal(t)

	seq, err := store.List()
	if err != nil {
		t.Fatalf("List on empty store: %v", err)
	}
	count := 0
	for range seq {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 contracts, got %d", count)
	}

	// Adding after empty state should work without error (no contracts.json yet).
	if err := store.Add("0xnew", metaJSON(t, "New", "/abi")); err != nil {
		t.Fatalf("Add to fresh store: %v", err)
	}
}
