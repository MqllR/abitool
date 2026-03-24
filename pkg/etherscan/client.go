package etherscan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const etherscanBaseURL = "https://api.etherscan.io/v2/api"

type ChainID string

var Mainnet ChainID = "1"

// FromInt converts an Int into a ChainID
func FromInt(id int) ChainID {
	return ChainID(fmt.Sprintf("%d", id))
}

type response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  any    `json:"result"`
}

type Client struct {
	chainID    ChainID
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string, chainID ChainID) *Client {
	return &Client{
		chainID:    chainID,
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// buildURL constructs a safe Etherscan API URL using url.Values to prevent query-parameter injection.
func (c *Client) buildURL(module, action, address string) string {
	q := url.Values{}
	q.Set("chainid", string(c.chainID))
	q.Set("apikey", c.apiKey)
	q.Set("module", module)
	q.Set("action", action)
	q.Set("address", address)
	return etherscanBaseURL + "?" + q.Encode()
}

// call sends the request to Etherscan and returns the structured response.
// It validates both the HTTP status code and the API-level status field in the JSON body.
func (c *Client) call(ctx context.Context, rawURL string) (*response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %v", resp.StatusCode)
	}

	var jsonResp response
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if jsonResp.Status != "1" {
		return nil, fmt.Errorf("etherscan error: %s - %v", jsonResp.Message, jsonResp.Result)
	}

	return &jsonResp, nil
}
