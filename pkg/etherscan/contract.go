package etherscan

import (
	"context"
	"fmt"
)

func (c *Client) GetABI(ctx context.Context, address string) (string, error) {
	url := c.baseURL + "&module=contract&action=getabi&address=" + address

	resp, err := c.call(ctx, url)
	if err != nil {
		return "", fmt.Errorf("calling etherscan: %w", err)
	}

	return resp.Result, nil
}
