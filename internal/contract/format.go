// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package contract

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
)

// jsonMarshaler and textMarshaler are local aliases to avoid importing
// encoding/json and encoding in the same file as their consumers.
type jsonMarshaler interface {
	MarshalJSON() ([]byte, error)
}

type textMarshaler interface {
	MarshalText() ([]byte, error)
}

// formatValue returns a human-friendly string representation of a decoded ABI
// value. []byte and [N]byte values are rendered as 0x-prefixed hex strings.
// Structs (tuples), slices, and arrays are rendered recursively.
func formatValue(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	return formatReflect(reflect.ValueOf(v))
}

func formatReflect(rv reflect.Value) string {
	// If the value (or its pointer) implements Stringer, use it directly.
	// This must happen before pointer dereferencing so that types like *big.Int
	// (which has a pointer-receiver String()) are caught here.
	if rv.CanInterface() {
		if s, ok := rv.Interface().(fmt.Stringer); ok {
			return s.String()
		}
	}

	// Dereference pointers.
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return "<nil>"
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Slice:
		if rv.IsNil() {
			return "<nil>"
		}
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return "0x" + hex.EncodeToString(rv.Bytes())
		}
		parts := make([]string, rv.Len())
		for i := range parts {
			parts[i] = formatReflect(rv.Index(i))
		}
		return "[" + strings.Join(parts, " ") + "]"

	case reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			b := make([]byte, rv.Len())
			for i := range b {
				b[i] = byte(rv.Index(i).Uint())
			}
			return "0x" + hex.EncodeToString(b)
		}
		parts := make([]string, rv.Len())
		for i := range parts {
			parts[i] = formatReflect(rv.Index(i))
		}
		return "[" + strings.Join(parts, " ") + "]"

	case reflect.Struct:
		t := rv.Type()
		var parts []string
		for i := 0; i < rv.NumField(); i++ {
			if !t.Field(i).IsExported() {
				continue
			}
			parts = append(parts, formatReflect(rv.Field(i)))
		}
		return "{" + strings.Join(parts, " ") + "}"

	default:
		if rv.CanInterface() {
			return fmt.Sprintf("%v", rv.Interface())
		}
		return fmt.Sprintf("%v", rv)
	}
}

// normalizeForJSON recursively transforms decoded ABI values so that
// json.Marshal produces human-friendly output:
//   - []byte  → 0x-prefixed hex string
//   - [N]byte → 0x-prefixed hex string
//   - structs → map[string]interface{} (preserving ABI field names)
//   - slices / arrays → []interface{} with normalized elements
//   - everything else → passed through unchanged
func normalizeForJSON(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	return normalizeReflect(reflect.ValueOf(v))
}

func normalizeReflect(rv reflect.Value) interface{} {
	// If the value knows how to marshal itself for JSON, leave it alone.
	// This handles *big.Int (MarshalJSON), common.Address (MarshalText), etc.
	if rv.CanInterface() {
		v := rv.Interface()
		if _, ok := v.(jsonMarshaler); ok {
			return v
		}
		if _, ok := v.(textMarshaler); ok {
			return v
		}
	}

	// Dereference pointers.
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Slice:
		if rv.IsNil() {
			return nil
		}
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return "0x" + hex.EncodeToString(rv.Bytes())
		}
		result := make([]interface{}, rv.Len())
		for i := range result {
			result[i] = normalizeReflect(rv.Index(i))
		}
		return result

	case reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			b := make([]byte, rv.Len())
			for i := range b {
				b[i] = byte(rv.Index(i).Uint())
			}
			return "0x" + hex.EncodeToString(b)
		}
		result := make([]interface{}, rv.Len())
		for i := range result {
			result[i] = normalizeReflect(rv.Index(i))
		}
		return result

	case reflect.Struct:
		t := rv.Type()
		m := make(map[string]interface{})
		for i := 0; i < rv.NumField(); i++ {
			if !t.Field(i).IsExported() {
				continue
			}
			m[t.Field(i).Name] = normalizeReflect(rv.Field(i))
		}
		return m

	default:
		if rv.CanInterface() {
			return rv.Interface()
		}
		return nil
	}
}
