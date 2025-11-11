package config

import (
	"os"
	"strings"

	validatorPkg "github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Backends map[string]BackendConfig `yaml:"backends" mapstructure:"backends"`
}

type BackendConfig struct {
	Type     string                 `yaml:"type" mapstructure:"type"`
	Deps     []string               `yaml:"deps" mapstructure:"deps"`
	Settings map[string]interface{} `yaml:",inline" mapstructure:",remain"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("logger.instance", os.Getenv("HOSTNAME"))
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, validatorPkg.New().Struct(&config)
}
