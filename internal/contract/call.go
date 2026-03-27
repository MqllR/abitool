package contract

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strconv"

	"github.com/spf13/viper"

	"github.com/MqllR/abitool/internal/abitool"
	"github.com/MqllR/abitool/internal/ui"
	"github.com/MqllR/abitool/pkg/abicodec"
	"github.com/MqllR/abitool/pkg/abiparser"
	"github.com/MqllR/abitool/pkg/ethclient"
	abistore "github.com/MqllR/abitool/pkg/storage/abi"
	contractstore "github.com/MqllR/abitool/pkg/storage/contract"
)

// CallManager executes read-only contract calls (eth_call) using locally stored ABIs.
type CallManager struct {
	log           *log.Logger
	contractStore *contractstore.Local
	abiStore      *abistore.Local
	rpcURL        string
}

// NewCallManager creates a CallManager. The RPC URL is resolved from (in order of precedence):
//  1. the --rpc-url flag (bound via viper key "rpc-url")
//  2. the rpc.url field in the config file
func NewCallManager(logger *log.Logger) (*CallManager, error) {
	cfg := abitool.ConfigInstance()

	rpcURL := viper.GetString("rpc-url")
	if rpcURL == "" {
		rpcURL = cfg.RPC.URL
	}
	if rpcURL == "" {
		return nil, errors.New("RPC URL is not set: use --rpc-url flag or set rpc.url in config")
	}

	storePath := viper.GetString("abi-store")
	chainID := viper.GetInt("chainid")

	basePath := filepath.Join(storePath, strconv.Itoa(chainID))

	cs, err := contractstore.NewLocal(basePath)
	if err != nil {
		return nil, err
	}

	as, err := abistore.NewLocal(basePath)
	if err != nil {
		return nil, err
	}

	return &CallManager{
		log:           logger,
		contractStore: cs,
		abiStore:      as,
		rpcURL:        rpcURL,
	}, nil
}

// CallOptions controls how CallContract executes.
type CallOptions struct {
	// Interactive forces the bubbletea input form even when CLI args are provided.
	Interactive bool
	// Block is the block tag or hex number for eth_call (default: "latest").
	Block string
	// OutputJSON formats the decoded result as JSON instead of plain text.
	OutputJSON bool
}

// CallContract looks up the stored ABI for address, resolves functionName,
// gathers inputs (interactively or from args), sends eth_call, and writes the
// decoded result to out.
func (m *CallManager) CallContract(ctx context.Context, address, functionName string, args []string, opts CallOptions, out io.Writer) error {
	if opts.Block == "" {
		opts.Block = "latest"
	}

	element, err := m.loadFunctionElement(address, functionName)
	if err != nil {
		return err
	}

	method, err := abicodec.ParseMethod(*element)
	if err != nil {
		return fmt.Errorf("parsing method: %w", err)
	}

	// Gather inputs.
	if opts.Interactive || (len(element.Inputs) > 0 && len(args) == 0) {
		fields := make([]ui.FormField, len(element.Inputs))
		for i, inp := range element.Inputs {
			fields[i] = ui.FormField{Name: inp.Name, Type: inp.Type}
		}

		args, err = ui.RunForm(fields)
		if err != nil {
			return fmt.Errorf("collecting inputs: %w", err)
		}
	}

	// Encode calldata.
	calldata, err := abicodec.EncodeInput(method, args)
	if err != nil {
		return fmt.Errorf("encoding calldata: %w", err)
	}

	// Dial + call.
	client, err := ethclient.Dial(ctx, m.rpcURL)
	if err != nil {
		return err
	}
	defer client.Close()

	raw, err := client.CallContract(ctx, address, calldata, opts.Block)
	if err != nil {
		return err
	}

	// Decode output.
	values, err := abicodec.DecodeOutput(method, raw)
	if err != nil {
		return fmt.Errorf("decoding output: %w", err)
	}

	return writeResult(out, values, opts.OutputJSON)
}

// loadFunctionElement reads the stored ABI and returns the Element matching functionName.
func (m *CallManager) loadFunctionElement(address, functionName string) (*abiparser.Element, error) {
	rawMeta, err := m.contractStore.Get(address)
	if err != nil {
		return nil, fmt.Errorf("contract %s not found in store: %w", address, err)
	}

	var meta Metadata
	if err := json.Unmarshal(rawMeta, &meta); err != nil {
		return nil, fmt.Errorf("unmarshaling contract metadata: %w", err)
	}

	rawABI, err := m.abiStore.Read(address)
	if err != nil {
		return nil, fmt.Errorf("reading ABI for %s: %w", address, err)
	}

	parsedABI, err := abiparser.ParseABI(rawABI)
	if err != nil {
		return nil, fmt.Errorf("parsing ABI: %w", err)
	}

	for el := range parsedABI.All() {
		if el.IsFunction() && el.Name == functionName {
			return &el, nil
		}
	}

	return nil, fmt.Errorf("function %q not found in ABI for contract %s", functionName, address)
}

// writeResult formats and writes the decoded output values to out.
func writeResult(out io.Writer, values []interface{}, asJSON bool) error {
	if len(values) == 0 {
		_, err := fmt.Fprintln(out, "(no return value)")
		return err
	}

	if asJSON {
		b, err := json.MarshalIndent(values, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling result to JSON: %w", err)
		}
		_, err = fmt.Fprintln(out, string(b))
		return err
	}

	for i, v := range values {
		_, err := fmt.Fprintf(out, "[%d] %v\n", i, v)
		if err != nil {
			return err
		}
	}

	return nil
}
