package abitool

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/spf13/viper"
)

var once sync.Once

func Load() error {
	var err error

	once.Do(func() {
		var c Config

		configPath := viper.GetString("config")
		cfgWithEnv := os.ExpandEnv(configPath)

		var fh io.ReadCloser
		fh, err = os.Open(cfgWithEnv)
		if err != nil {
			return
		}
		defer fh.Close()

		viper.SetConfigType("yaml")

		err = viper.ReadConfig(fh)
		if err != nil {
			return
		}

		err = viper.Unmarshal(&c)
		if err != nil {
			return
		}
		cfg = c
	})

	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	return nil
}
