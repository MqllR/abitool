package etherscan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	baseURL    string
	httpClient *http.Client
}

func NewClient(apiKey string, chainID ChainID) *Client {
	baseURL := etherscanBaseURL
	baseURL += "?chainid=" + string(chainID)
	baseURL += "&apikey=" + apiKey

	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// call sends the requests to etherscan and return the structured response
func (c *Client) call(ctx context.Context, url string) (*response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status ok is not ok: %v", resp.StatusCode)
	}

	var jsonResp response
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &jsonResp, nil
}
