// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package abi

import (
	"os"
	"path/filepath"
	"strings"
)

type Local struct {
	path string
}

func NewLocal(storePath string) (*Local, error) {
	storePathWithEnv := os.ExpandEnv(storePath)

	if err := os.MkdirAll(storePathWithEnv, 0755); err != nil {
		return nil, err
	}

	return &Local{
		path: storePathWithEnv,
	}, nil
}

// Write writes the ABI data for a given contract ID.
func (l *Local) Write(id string, data string) error {
	return os.WriteFile(filepath.Join(l.path, strings.ToLower(id)), []byte(data), 0644)
}

// Read reads the ABI data for a given contract ID.
func (l *Local) Read(id string) (string, error) {
	data, err := os.ReadFile(filepath.Join(l.path, strings.ToLower(id)))
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (l *Local) Delete(id string) error {
	return os.Remove(filepath.Join(l.path, strings.ToLower(id)))
}

// GetPath returns the file path of the ABI for a given contract ID.
func (l *Local) GetPath(id string) string {
	return filepath.Join(l.path, strings.ToLower(id))
}
