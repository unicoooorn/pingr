package config

import (
	"fmt"
	"gopkg.in/go-playground/validator.v9"
)

func ValidateConfig(config *Config) error {
	validate := validator.New()
	validate.RegisterStructValidation(func(sl validator.StructLevel) {
		cfg := sl.Current().Interface().(BackendConfig)
		switch cfg.Type {
		case "http":
			if cfg.URL == "" {
				sl.ReportError(cfg.URL, "URL", "url", "required_for_http", "")
			}
		case "grpc":
			if cfg.Host == "" || cfg.Port == 0 {
				sl.ReportError(cfg.Host, "Host", "host", "required_for_grpc", "")
				sl.ReportError(cfg.Port, "Port", "port", "required_for_grpc", "")
			}
		case "icmp":
			if cfg.Host == "" {
				sl.ReportError(cfg.Host, "Host", "host", "required_for_icmp", "")
			}
		case "tcp", "redis":
			if cfg.Host == "" || cfg.Port == 0 {
				sl.ReportError(cfg.Host, "Host", "host", "required_for_"+cfg.Type, "")
				sl.ReportError(cfg.Port, "Port", "port", "required_for_"+cfg.Type, "")
			}
		case "postgres":
			if cfg.URL == "" {
				sl.ReportError(cfg.URL, "URL", "url", "required_for_postgres", "")
			}
		}
	}, BackendConfig{})

	for name, backend := range config.Backends {
		if err := validate.Struct(backend); err != nil {
			return fmt.Errorf("invalid config in '%s': %w", name, err)
		}
	}
	return nil
}
