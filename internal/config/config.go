package config

import (
	"os"
	"strings"

	validatorPkg "github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Name string `mapstructure:"name"`
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
