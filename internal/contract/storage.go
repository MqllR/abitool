package contract

import (
	"encoding/json"
	"fmt"

	"github.com/MqllR/abitool/pkg/storage/contract"
)

type Metadata struct {
	ContractName string `json:"contract_name"`
	ABIPath      string `json:"abi_path"`
}

type Contract struct {
	Address  string   `json:"address"`
	Metadata Metadata `json:"metadata"`
}

func (a *ABIManager) GetContract(address string) (*Contract, error) {
	rawMeta, err := a.contractStore.Get(address)
	if err != nil {
		if err == contract.ErrNotFound {
			return nil, fmt.Errorf("contract not found: %w", err)
		}

		return nil, fmt.Errorf("getting contract: %w", err)
	}

	var meta Metadata
	if err := json.Unmarshal(rawMeta, &meta); err != nil {
		return nil, fmt.Errorf("unmarshaling contract metadata: %w", err)
	}

	return &Contract{
		Address:  address,
		Metadata: meta,
	}, nil
}

func (a *ABIManager) SaveContractWithABI(address string, meta *Metadata, abi string) error {
	_, err := a.contractStore.Get(address)
	if err != nil {
		if err != contract.ErrNotFound {
			return fmt.Errorf("getting contract: %w", err)
		}
	}

	if err := a.abiStore.Write(address, abi); err != nil {
		return fmt.Errorf("writing ABI to store: %w", err)
	}

	meta.ABIPath = a.abiStore.GetPath(address)

	rawMeta, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshaling contract metadata: %w", err)
	}

	if err := a.contractStore.Add(address, rawMeta); err != nil {
		return fmt.Errorf("adding contract to store: %w", err)
	}

	return nil
}
