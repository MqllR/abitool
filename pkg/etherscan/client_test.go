// Copyright 2025 MqllR. All rights reserved.
// SPDX-License-Identifier: MIT

package etherscan

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// ---- FromInt / ChainID ----

func TestFromInt(t *testing.T) {
	cases := []struct {
		id   int
		want string
	}{
		{1, "1"},
		{137, "137"},
		{11155111, "11155111"},
		{0, "0"},
	}
	for _, tc := range cases {
		got := string(FromInt(tc.id))
		if got != tc.want {
			t.Errorf("FromInt(%d): got %q, want %q", tc.id, got, tc.want)
		}
	}
}

// ---- buildURL ----

func TestBuildURL_ContainsRequiredParams(t *testing.T) {
	c := NewClient("myapikey", FromInt(1))
	rawURL := c.buildURL("contract", "getabi", "0x1234")

	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("url.Parse: %v", err)
	}

	q := u.Query()
	checks := map[string]string{
		"chainid": "1",
		"apikey":  "myapikey",
		"module":  "contract",
		"action":  "getabi",
		"address": "0x1234",
	}
	for key, want := range checks {
		if got := q.Get(key); got != want {
			t.Errorf("query param %q: got %q, want %q", key, got, want)
		}
	}
}

func TestBuildURL_ChainID137(t *testing.T) {
	c := NewClient("key", FromInt(137))
	rawURL := c.buildURL("contract", "getabi", "0xabc")

	u, _ := url.Parse(rawURL)
	if got := u.Query().Get("chainid"); got != "137" {
		t.Errorf("chainid: got %q, want %q", got, "137")
	}
}

// ---- call (via httptest) ----

func TestCall_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"1","message":"OK","result":"some_abi"}`)
	}))
	defer ts.Close()

	c := NewClient("testkey", FromInt(1))
	resp, err := c.call(context.Background(), ts.URL)
	if err != nil {
		t.Fatalf("call: unexpected error: %v", err)
	}
	if resp.Message != "OK" {
		t.Errorf("message: got %q, want %q", resp.Message, "OK")
	}
}

func TestCall_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"0","message":"NOTOK","result":"Contract source code not verified"}`)
	}))
	defer ts.Close()

	c := NewClient("testkey", FromInt(1))
	_, err := c.call(context.Background(), ts.URL)
	if err == nil {
		t.Fatal("call: expected error for API error, got nil")
	}
	if !strings.Contains(err.Error(), "etherscan error") {
		t.Errorf("error message: got %q, want it to contain 'etherscan error'", err.Error())
	}
}

func TestCall_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := NewClient("testkey", FromInt(1))
	_, err := c.call(context.Background(), ts.URL)
	if err == nil {
		t.Fatal("call: expected error for HTTP 500, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error message: got %q, want it to mention status 500", err.Error())
	}
}

func TestCall_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `not valid json at all!!!`)
	}))
	defer ts.Close()

	c := NewClient("testkey", FromInt(1))
	_, err := c.call(context.Background(), ts.URL)
	if err == nil {
		t.Fatal("call: expected error for invalid JSON, got nil")
	}
}

// ---- GetABI (via httptest) ----

func TestGetABI_Success(t *testing.T) {
	const abiPayload = `[{"type":"function","name":"transfer"}]`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"1","message":"OK","result":%q}`, abiPayload)
	}))
	defer ts.Close()

	// Override the base URL so GetABI routes to our test server.
	orig := etherscanBaseURL
	etherscanBaseURL = ts.URL
	defer func() { etherscanBaseURL = orig }()

	c := NewClient("testkey", FromInt(1))
	got, err := c.GetABI(context.Background(), "0xabc")
	if err != nil {
		t.Fatalf("GetABI: %v", err)
	}
	if got != abiPayload {
		t.Errorf("GetABI: got %q, want %q", got, abiPayload)
	}
}

func TestGetABI_NonStringResult(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"1","message":"OK","result":42}`)
	}))
	defer ts.Close()

	orig := etherscanBaseURL
	etherscanBaseURL = ts.URL
	defer func() { etherscanBaseURL = orig }()

	c := NewClient("testkey", FromInt(1))
	_, err := c.GetABI(context.Background(), "0xabc")
	if err == nil {
		t.Fatal("GetABI with non-string result: expected error, got nil")
	}
}

// ---- GetSourceCode (via httptest) ----

func TestGetSourceCode_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"1","message":"OK","result":[{"ABI":"[{}]","ContractName":"MyToken"}]}`)
	}))
	defer ts.Close()

	orig := etherscanBaseURL
	etherscanBaseURL = ts.URL
	defer func() { etherscanBaseURL = orig }()

	c := NewClient("testkey", FromInt(1))
	got, err := c.GetSourceCode(context.Background(), "0xabc")
	if err != nil {
		t.Fatalf("GetSourceCode: %v", err)
	}
	if got.ContractName != "MyToken" {
		t.Errorf("ContractName: got %q, want %q", got.ContractName, "MyToken")
	}
	if got.ABI != "[{}]" {
		t.Errorf("ABI: got %q, want %q", got.ABI, "[{}]")
	}
}

func TestGetSourceCode_EmptyArray(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"1","message":"OK","result":[]}`)
	}))
	defer ts.Close()

	orig := etherscanBaseURL
	etherscanBaseURL = ts.URL
	defer func() { etherscanBaseURL = orig }()

	c := NewClient("testkey", FromInt(1))
	_, err := c.GetSourceCode(context.Background(), "0xabc")
	if err == nil {
		t.Fatal("GetSourceCode with empty array: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no result found") {
		t.Errorf("error: got %q, want it to contain 'no result found'", err.Error())
	}
}
