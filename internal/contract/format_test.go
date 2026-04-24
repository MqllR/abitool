// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package contract

import (
	"encoding/json"
	"math/big"
	"strings"
	"testing"
)

// ---- formatValue ----

func TestFormatValue_Bytes(t *testing.T) {
	b := []byte{0x00, 0xe0, 0xf7, 0x83}
	got := formatValue(b)
	want := "0x00e0f783"
	if got != want {
		t.Errorf("formatValue([]byte): got %q, want %q", got, want)
	}
}

func TestFormatValue_EmptyBytes(t *testing.T) {
	got := formatValue([]byte{})
	want := "0x"
	if got != want {
		t.Errorf("formatValue([]byte{}): got %q, want %q", got, want)
	}
}

func TestFormatValue_FixedBytes4(t *testing.T) {
	b := [4]byte{0xde, 0xad, 0xbe, 0xef}
	got := formatValue(b)
	want := "0xdeadbeef"
	if got != want {
		t.Errorf("formatValue([4]byte): got %q, want %q", got, want)
	}
}

func TestFormatValue_FixedBytes32(t *testing.T) {
	var b [32]byte
	b[31] = 0x01
	got := formatValue(b)
	want := "0x" + strings.Repeat("00", 31) + "01"
	if got != want {
		t.Errorf("formatValue([32]byte): got %q, want %q", got, want)
	}
}

func TestFormatValue_Uint256(t *testing.T) {
	n := big.NewInt(12345)
	got := formatValue(n)
	want := "12345"
	if got != want {
		t.Errorf("formatValue(*big.Int): got %q, want %q", got, want)
	}
}

func TestFormatValue_Bool(t *testing.T) {
	if got := formatValue(true); got != "true" {
		t.Errorf("formatValue(true): got %q, want %q", got, "true")
	}
	if got := formatValue(false); got != "false" {
		t.Errorf("formatValue(false): got %q, want %q", got, "false")
	}
}

func TestFormatValue_SliceOfInts(t *testing.T) {
	s := []uint32{1, 2, 3}
	got := formatValue(s)
	want := "[1 2 3]"
	if got != want {
		t.Errorf("formatValue([]uint32): got %q, want %q", got, want)
	}
}

func TestFormatValue_StructWithBytes(t *testing.T) {
	type call struct {
		Target  string
		AllowFail bool
		CallData  []byte
	}
	v := call{
		Target:    "0xABCD",
		AllowFail: true,
		CallData:  []byte{0x01, 0x02},
	}
	got := formatValue(v)
	want := "{0xABCD true 0x0102}"
	if got != want {
		t.Errorf("formatValue(struct): got %q, want %q", got, want)
	}
}

func TestFormatValue_SliceOfStructsWithBytes(t *testing.T) {
	type call struct {
		Data []byte
	}
	s := []call{{Data: []byte{0xAA}}, {Data: []byte{0xBB}}}
	got := formatValue(s)
	want := "[{0xaa} {0xbb}]"
	if got != want {
		t.Errorf("formatValue([]struct): got %q, want %q", got, want)
	}
}

func TestFormatValue_Nil(t *testing.T) {
	got := formatValue(nil)
	want := "<nil>"
	if got != want {
		t.Errorf("formatValue(nil): got %q, want %q", got, want)
	}
}

// ---- normalizeForJSON ----

func TestNormalizeForJSON_Bytes(t *testing.T) {
	b := []byte{0x00, 0xe0}
	got := normalizeForJSON(b)
	want := "0x00e0"
	if got != want {
		t.Errorf("normalizeForJSON([]byte): got %v, want %v", got, want)
	}
}

func TestNormalizeForJSON_FixedBytes(t *testing.T) {
	b := [2]byte{0xca, 0xfe}
	got := normalizeForJSON(b)
	want := "0xcafe"
	if got != want {
		t.Errorf("normalizeForJSON([2]byte): got %v, want %v", got, want)
	}
}

func TestNormalizeForJSON_SliceOfBytes(t *testing.T) {
	s := [][]byte{{0x01}, {0x02}}
	got := normalizeForJSON(s)
	b, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	want := `["0x01","0x02"]`
	if string(b) != want {
		t.Errorf("normalizeForJSON([][]byte) JSON: got %s, want %s", b, want)
	}
}

func TestNormalizeForJSON_StructWithBytes(t *testing.T) {
	type row struct {
		Addr string
		Data []byte
	}
	v := row{Addr: "0xABCD", Data: []byte{0xde, 0xad}}
	got := normalizeForJSON(v)

	m, ok := got.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", got)
	}
	if m["Data"] != "0xdead" {
		t.Errorf("Data field: got %v, want 0xdead", m["Data"])
	}
	if m["Addr"] != "0xABCD" {
		t.Errorf("Addr field: got %v, want 0xABCD", m["Addr"])
	}
}

func TestNormalizeForJSON_Nil(t *testing.T) {
	if got := normalizeForJSON(nil); got != nil {
		t.Errorf("normalizeForJSON(nil): got %v, want nil", got)
	}
}

func TestNormalizeForJSON_BigInt(t *testing.T) {
	n := big.NewInt(999)
	got := normalizeForJSON(n)
	if got != n {
		t.Errorf("normalizeForJSON(*big.Int): got %v, want %v", got, n)
	}
}
