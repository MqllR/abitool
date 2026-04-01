// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package abi_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	abistore "github.com/MqllR/abitool/pkg/storage/abi"
)

func newLocal(t *testing.T) *abistore.Local {
	t.Helper()
	store, err := abistore.NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocal: %v", err)
	}
	return store
}

func TestWrite_Read_RoundTrip(t *testing.T) {
	store := newLocal(t)
	const id = "0xabc"
	const data = `[{"type":"function"}]`

	if err := store.Write(id, data); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := store.Read(id)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got != data {
		t.Errorf("Read: got %q, want %q", got, data)
	}
}

func TestRead_NonExistent(t *testing.T) {
	store := newLocal(t)
	_, err := store.Read("does-not-exist")
	if err == nil {
		t.Fatal("Read non-existent file: expected error, got nil")
	}
}

func TestDelete_Existing(t *testing.T) {
	store := newLocal(t)
	const id = "0xdel"

	if err := store.Write(id, "data"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := store.Delete(id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := store.Read(id)
	if err == nil {
		t.Fatal("Read after delete: expected error, got nil")
	}
}

func TestDelete_NonExistent(t *testing.T) {
	store := newLocal(t)
	err := store.Delete("does-not-exist")
	if err == nil {
		t.Fatal("Delete non-existent file: expected error, got nil")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected os.ErrNotExist, got %v", err)
	}
}

func TestGetPath_ReturnsCorrectPath(t *testing.T) {
	dir := t.TempDir()
	store, err := abistore.NewLocal(dir)
	if err != nil {
		t.Fatalf("NewLocal: %v", err)
	}

	const id = "0x1234"
	got := store.GetPath(id)
	want := filepath.Join(dir, id)
	if got != want {
		t.Errorf("GetPath: got %q, want %q", got, want)
	}
}

func TestWrite_Overwrite(t *testing.T) {
	store := newLocal(t)
	const id = "0xover"

	if err := store.Write(id, "original"); err != nil {
		t.Fatalf("Write (first): %v", err)
	}
	if err := store.Write(id, "updated"); err != nil {
		t.Fatalf("Write (overwrite): %v", err)
	}

	got, err := store.Read(id)
	if err != nil {
		t.Fatalf("Read after overwrite: %v", err)
	}
	if got != "updated" {
		t.Errorf("got %q, want %q", got, "updated")
	}
}
