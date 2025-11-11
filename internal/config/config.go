package config

import (
	"fmt"
	"os"
	"strings"
	"regexp"

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
	// Читаем файл как текст
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Раскрываем переменные окружения в тексте
	expandedContent, err := expandEnvVarsInText(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to expand environment variables: %w", err)
	}

	// Загружаем конфигурацию из обработанной строки
	viper.SetConfigType("yaml")
	if err := viper.ReadConfig(strings.NewReader(expandedContent)); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetDefault("logger.instance", os.Getenv("HOSTNAME"))

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, validatorPkg.New().Struct(&config)
}

// envVarRegex для поиска паттернов ${VAR_NAME} или ${VAR_NAME:default_value}
var envVarRegex = regexp.MustCompile(`\$\{([^}:]+)(?::([^}]*))?\}`)

// expandEnvVarsInText заменяет все вхождения переменных окружения в тексте.
// Поддерживает два формата:
//   - ${VAR_NAME} - обязательная переменная, вызовет ошибку если не установлена
//   - ${VAR_NAME:default_value} - переменная со значением по умолчанию
func expandEnvVarsInText(text string) (string, error) {
	var missingVars []string

	result := envVarRegex.ReplaceAllStringFunc(text, func(match string) string {
		matches := envVarRegex.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}

		varName := matches[1]
		defaultValue := ""
		if len(matches) > 2 && matches[2] != "" {
			defaultValue = matches[2]
		}

		value := os.Getenv(varName)
		if value == "" {
			if defaultValue != "" {
				return defaultValue
			}
			missingVars = append(missingVars, varName)
			return match
		}
		return value
	})

	if len(missingVars) > 0 {
		return "", fmt.Errorf("missing required environment variables: %s", strings.Join(missingVars, ", "))
	}

	return result, nil
}