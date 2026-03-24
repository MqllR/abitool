package abitool

var cfg Config

func ConfigInstance() *Config {
	return &cfg
}

type Config struct {
	EtherScan EtherScanConfig `mapstructure:"etherscan"`
	RPC       RPCConfig       `mapstructure:"rpc"`
}

type EtherScanConfig struct {
	APIKey string `mapstructure:"api_key"`
}

type RPCConfig struct {
	URL string `mapstructure:"url"`
}
