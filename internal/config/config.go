package config

import (
	"os"
	"strings"
	"time"

	validatorPkg "github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Backends   map[string]BackendConfig `yaml:"backends" mapstructure:"backends"`
	Prometheus PrometheusConfig         `yaml:"prometheus" mapstructure:"prometheus"`
}

type BackendConfig struct {
	Type           string            `yaml:"type" mapstructure:"type"`
	Deps           []string          `yaml:"deps" mapstructure:"deps"`
	URL            string            `yaml:"url" mapstructure:"url"`
	Timeout        time.Duration     `yaml:"timeout" mapstructure:"timeout"`
	Host           string            `yaml:"host" mapstructure:"host"`
	Port           int               `yaml:"port" mapstructure:"port"`
	Headers        map[string]string `yaml:"headers" mapstructure:"headers"`
	MetricsQueries []string          `yaml:"metrics_queries" mapstructure:"metrics_queries"`
}

type PrometheusConfig struct {
	URL     string            `yaml:"url" mapstructure:"url"`
	Timeout time.Duration     `yaml:"timeout" mapstructure:"timeout"`
	Headers map[string]string `yaml:"headers" mapstructure:"headers"`
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
