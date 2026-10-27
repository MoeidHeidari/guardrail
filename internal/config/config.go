package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func Load(cfgFile string) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		viper.AddConfigPath(filepath.Join(home))
		viper.SetConfigName(".guardrail")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("GUARDRAIL")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}
	return nil
}

func GetString(key string) string { return viper.GetString(key) }
func GetBool(key string) bool     { return viper.GetBool(key) }
