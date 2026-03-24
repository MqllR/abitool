package abitool

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

var (
	once    sync.Once
	loadErr error
)

func Load() error {
	once.Do(func() {
		var c Config

		configPath := viper.GetString("config")
		cfgWithEnv := os.ExpandEnv(configPath)

		fh, err := os.Open(cfgWithEnv)
		if err != nil {
			loadErr = err
			return
		}
		defer func() { _ = fh.Close() }()

		viper.SetConfigType("yaml")

		if err = viper.ReadConfig(fh); err != nil {
			loadErr = err
			return
		}

		if err = viper.Unmarshal(&c); err != nil {
			loadErr = err
			return
		}

		cfg = c
	})

	if loadErr != nil {
		return fmt.Errorf("loading config: %w", loadErr)
	}

	return nil
}
