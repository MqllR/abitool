package abitool

var cfg Config

func ConfigInstance() *Config {
	return &cfg
}

type Config struct {
	EtherScan EtherScanConfig `mapstructure:"etherscan"`
}

type EtherScanConfig struct {
	APIKey string `mapstructure:"api_key"`
}
