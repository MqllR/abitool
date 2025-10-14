package etherscan

import (
	"context"
	"fmt"
)

// GetABI retrieves the ABI for a given contract address.
// https://docs.etherscan.io/api-endpoints/contracts#get-contract-abi-for-verified-contract-source-codes
func (c *Client) GetABI(ctx context.Context, address string) (string, error) {
	url := c.baseURL + "&module=contract&action=getabi&address=" + address

	resp, err := c.call(ctx, url)
	if err != nil {
		return "", fmt.Errorf("calling etherscan: %w", err)
	}

	resultStr, ok := resp.Result.(string)
	if !ok {
		return "", fmt.Errorf("unexpected result type: %T", resp.Result)
	}

	return resultStr, nil
}

// GetSourceCode retrieves the source code for a given contract address.
// https://docs.etherscan.io/api-endpoints/contracts#get-contract-source-code-for-verified-contract-source-codes
func (c *Client) GetSourceCode(ctx context.Context, address string) (*ContractSourceCodeResponse, error) {
	url := c.baseURL + "&module=contract&action=getsourcecode&address=" + address

	resp, err := c.call(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("calling etherscan: %w", err)
	}

	responseSlice, ok := resp.Result.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", resp.Result)
	}

	if len(responseSlice) == 0 {
		return nil, fmt.Errorf("no result found for address: %s", address)
	}

	responseMap, ok := responseSlice[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected result item type: %T", responseSlice[0])
	}

	var contract ContractSourceCodeResponse

	if resultABI, ok := responseMap["ABI"].(string); ok {
		contract.ABI = resultABI
	}

	if resultContractName, ok := responseMap["ContractName"].(string); ok {
		contract.ContractName = resultContractName
	}

	return &contract, nil
}
