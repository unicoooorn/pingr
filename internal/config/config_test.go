package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unicoooorn/pingr/internal/config"
)

func TestLoadConfig(t *testing.T) {
	// Подготовим окружение
	_ = os.Setenv("BACKENDS_MYAPI_HOST", "example.com")
	_ = os.Setenv("BACKENDS_MYAPI_URL", "https://api.example.com")
	_ = os.Setenv("TIMEOUT", "12")

	// Тестовый yaml
	yaml := `
backends:
  myapi:
    type: http
    deps: []
    url: ${API_URL}
    timeout: 15s
    host: "example.com"
    port: 8080
    headers:
      Authorization: "Bearer mytoken"
      X-Feature: "test"
  local:
    type: local
    deps: ["myapi"]
    url: "file:///tmp/local"
    timeout: 0
    host: "localhost"
    port: 0
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(yaml), 0644))

	cfg, err := config.Load(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Contains(t, cfg.Backends, "myapi")
	assert.Contains(t, cfg.Backends, "local")

	api := cfg.Backends["myapi"]
	assert.Equal(t, "http", api.Type)
	assert.Equal(t, []string{}, api.Deps)
	assert.Equal(t, "https://api.example.com", api.URL)
	assert.Equal(t, 15 * time.Second, api.Timeout) // Проверяет, что YAML имеет больший приоритет чем переменная окружения
	assert.Equal(t, "example.com", api.Host)
	assert.Equal(t, 8080, api.Port)
	assert.Equal(t, map[string]string{
		"authorization": "Bearer mytoken",
		"x-feature":     "test",
	}, api.Headers)

	local := cfg.Backends["local"]
	assert.Equal(t, "local", local.Type)
	assert.Equal(t, []string{"myapi"}, local.Deps)
	assert.Equal(t, "file:///tmp/local", local.URL)
	assert.Equal(t, time.Duration(0), local.Timeout)
	assert.Equal(t, "localhost", local.Host)
	assert.Equal(t, 0, local.Port)
	assert.Nil(t, local.Headers)
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := config.Load("/path/does/not/exist.yaml")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "no such file"))

}