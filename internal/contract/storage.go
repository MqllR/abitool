package contract

import (
	"encoding/json"
	"fmt"

	"github.com/MqllR/abitool/pkg/storage/contract"
)

func (a *ABIManager) getContract(address string) (*Contract, error) {
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

func (a *ABIManager) listContracts() ([]*Contract, error) {
	contractIter, err := a.contractStore.List()
	if err != nil {
		return nil, fmt.Errorf("listing contracts: %w", err)
	}

	var contracts []*Contract

	for address := range contractIter {
		c, err := a.getContract(address)
		if err != nil {
			return nil, fmt.Errorf("getting contract %s: %w", address, err)
		}
		contracts = append(contracts, c)
	}

	return contracts, nil
}

func (a *ABIManager) saveContractWithABI(address string, meta *Metadata, abi string) error {
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
