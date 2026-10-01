// Copyright (c) 2026 TrustReady <hello@trustready.io>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package bootstrap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/DhruvWork/trustready-grc/pkg/trustreadydconfig"
	"sigs.k8s.io/yaml"
)

func TestWriteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "trustreadyd.yml")

	cfg := &trustreadydconfig.FullConfig{
		Unit: trustreadydconfig.UnitConfig{
			Metrics: trustreadydconfig.MetricsConfig{Addr: "localhost:9090"},
		},
		TrustReadyd: trustreadydconfig.Config{
			BaseURL:       "http://localhost:8080",
			EncryptionKey: "test-key",
		},
	}

	err := WriteConfig(cfg, configPath, FormatYAML)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var loaded trustreadydconfig.FullConfig

	err = yaml.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, cfg.Unit.Metrics.Addr, loaded.Unit.Metrics.Addr)
	assert.Equal(t, cfg.TrustReadyd.BaseURL, loaded.TrustReadyd.BaseURL)
	assert.Equal(t, cfg.TrustReadyd.EncryptionKey, loaded.TrustReadyd.EncryptionKey)
}

func TestWriteConfig_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nested", "dir", "trustreadyd.yml")

	cfg := &trustreadydconfig.FullConfig{
		TrustReadyd: trustreadydconfig.Config{BaseURL: "http://localhost:8080"},
	}

	err := WriteConfig(cfg, configPath, FormatYAML)
	require.NoError(t, err)

	_, err = os.Stat(configPath)
	require.NoError(t, err)
}

func TestWriteConfig_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "trustreadyd.yml")

	cfg := &trustreadydconfig.FullConfig{}

	err := WriteConfig(cfg, configPath, FormatYAML)
	require.NoError(t, err)

	info, err := os.Stat(configPath)
	require.NoError(t, err)

	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestWriteConfig_OmitsOptionalFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "trustreadyd.yml")

	cfg := &trustreadydconfig.FullConfig{
		Unit: trustreadydconfig.UnitConfig{
			Metrics: trustreadydconfig.MetricsConfig{Addr: "localhost:9090"},
			Tracing: trustreadydconfig.TracingConfig{Addr: ""},
		},
		TrustReadyd: trustreadydconfig.Config{
			BaseURL:      "http://localhost:8080",
			ChromeDPAddr: "",
			Pg: trustreadydconfig.PgConfig{
				Addr:     "localhost:5432",
				Username: "postgres",
				Password: "",
				Database: "",
				PoolSize: 100,
			},
		},
	}

	err := WriteConfig(cfg, configPath, FormatYAML)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var tree map[string]any

	err = yaml.Unmarshal(data, &tree)
	require.NoError(t, err)

	trustreadyd, ok := tree["trustreadyd"].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "http://localhost:8080", trustreadyd["base-url"])
	assert.NotContains(t, trustreadyd, "chrome-dp-addr")
	assert.NotContains(t, trustreadyd, "esign")

	pg, ok := trustreadyd["pg"].(map[string]any)
	require.True(t, ok)

	assert.Equal(t, "localhost:5432", pg["addr"])
	assert.Equal(t, "postgres", pg["username"])
	assert.NotContains(t, pg, "password")
	assert.NotContains(t, pg, "database")
	assert.Contains(t, pg, "pool-size")

	unit, ok := tree["unit"].(map[string]any)
	require.True(t, ok)

	metrics, ok := unit["metrics"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "localhost:9090", metrics["addr"])

	tracing, ok := unit["tracing"].(map[string]any)
	require.True(t, ok)
	assert.NotContains(t, tracing, "addr")
}

func TestWriteConfig_OmitsEmptyOptionalBlocks(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "trustreadyd.yml")

	cfg := &trustreadydconfig.FullConfig{
		TrustReadyd: trustreadydconfig.Config{
			BaseURL:       "http://localhost:8080",
			EncryptionKey: "test-key",
			Api: trustreadydconfig.APIConfig{
				Addr: ":8080",
			},
			Auth: trustreadydconfig.AuthConfig{
				Cookie: trustreadydconfig.CookieConfig{
					Name:   "SSID",
					Secret: "secret",
				},
				Password: trustreadydconfig.PasswordConfig{
					Pepper: "pepper",
				},
			},
			CompliancePortal: trustreadydconfig.CompliancePortalConfig{
				HTTPAddr: ":80",
			},
			CustomDomains: trustreadydconfig.CustomDomainsConfig{
				RenewalInterval: 3600,
			},
			Agents: trustreadydconfig.AgentsConfig{
				Default: trustreadydconfig.LLMAgentConfig{
					Provider:  "openai",
					ModelName: "gpt-4o",
				},
			},
		},
	}

	err := WriteConfig(cfg, configPath, FormatYAML)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var tree map[string]any

	err = yaml.Unmarshal(data, &tree)
	require.NoError(t, err)

	trustreadyd, ok := tree["trustreadyd"].(map[string]any)
	require.True(t, ok)

	assert.NotContains(t, trustreadyd, "esign")

	api, ok := trustreadyd["api"].(map[string]any)
	require.True(t, ok)
	assert.NotContains(t, api, "cors")
	assert.NotContains(t, api, "proxy-protocol")
	assert.NotContains(t, api, "extra-header-fields")

	auth, ok := trustreadyd["auth"].(map[string]any)
	require.True(t, ok)
	assert.NotContains(t, auth, "google")
	assert.NotContains(t, auth, "microsoft")

	customDomains, ok := trustreadyd["custom-domains"].(map[string]any)
	require.True(t, ok)
	assert.NotContains(t, customDomains, "acme")

	compliancePortal, ok := trustreadyd["trust-center"].(map[string]any)
	require.True(t, ok)
	assert.NotContains(t, compliancePortal, "proxy-protocol")

	llm, ok := trustreadyd["llm"].(map[string]any)
	require.True(t, ok)
	assert.NotContains(t, llm, "trustready")
	assert.NotContains(t, llm, "tools")
}

func TestWriteConfig_OmitsEmptyProxyProtocolAndCorsSlices(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "trustreadyd.yml")

	cfg := &trustreadydconfig.FullConfig{
		TrustReadyd: trustreadydconfig.Config{
			BaseURL: "http://localhost:8080",
			Api: trustreadydconfig.APIConfig{
				Addr: ":8080",
				ProxyProtocol: trustreadydconfig.ProxyProtocolConfig{
					TrustedProxies: []string{},
				},
				Cors: trustreadydconfig.CorsConfig{
					AllowedOrigins: []string{},
				},
			},
			CompliancePortal: trustreadydconfig.CompliancePortalConfig{
				HTTPAddr: ":10080",
				ProxyProtocol: trustreadydconfig.ProxyProtocolConfig{
					TrustedProxies: make([]string, 0),
				},
			},
		},
	}

	err := WriteConfig(cfg, configPath, FormatYAML)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var tree map[string]any

	err = yaml.Unmarshal(data, &tree)
	require.NoError(t, err)

	trustreadyd, ok := tree["trustreadyd"].(map[string]any)
	require.True(t, ok)

	api, ok := trustreadyd["api"].(map[string]any)
	require.True(t, ok)
	assert.NotContains(t, api, "proxy-protocol")
	assert.NotContains(t, api, "cors")

	compliancePortal, ok := trustreadyd["trust-center"].(map[string]any)
	require.True(t, ok)
	assert.NotContains(t, compliancePortal, "proxy-protocol")
}

func TestWriteConfig_OmitsEmptyExtraHeaderFieldsMap(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "trustreadyd.yml")

	cfg := &trustreadydconfig.FullConfig{
		TrustReadyd: trustreadydconfig.Config{
			BaseURL: "http://localhost:8080",
			Api: trustreadydconfig.APIConfig{
				Addr:              ":8080",
				ExtraHeaderFields: map[string]string{},
			},
		},
	}

	err := WriteConfig(cfg, configPath, FormatYAML)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var tree map[string]any

	err = yaml.Unmarshal(data, &tree)
	require.NoError(t, err)

	trustreadyd, ok := tree["trustreadyd"].(map[string]any)
	require.True(t, ok)

	api, ok := trustreadyd["api"].(map[string]any)
	require.True(t, ok)
	assert.NotContains(t, api, "extra-header-fields")
}

func TestWriteConfig_CompleteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "trustreadyd.yml")

	cfg := &trustreadydconfig.FullConfig{
		Unit: trustreadydconfig.UnitConfig{
			Metrics: trustreadydconfig.MetricsConfig{Addr: "localhost:8081"},
			Tracing: trustreadydconfig.TracingConfig{
				Addr:          "localhost:4317",
				MaxBatchSize:  512,
				BatchTimeout:  5,
				ExportTimeout: 30,
				MaxQueueSize:  2048,
			},
		},
		TrustReadyd: trustreadydconfig.Config{
			BaseURL:       "http://localhost:8080",
			EncryptionKey: "test-key",
			ChromeDPAddr:  "localhost:9222",
			Api: trustreadydconfig.APIConfig{
				Addr: ":8080",
				Cors: trustreadydconfig.CorsConfig{
					AllowedOrigins: []string{"http://localhost:8080"},
				},
			},
			Pg: trustreadydconfig.PgConfig{
				Addr:                   "localhost:5432",
				Username:               "postgres",
				Password:               "postgres",
				Database:               "trustreadyd",
				PoolSize:               100,
				MinPoolSize:            10,
				MaxConnIdleTimeSeconds: 1800,
				MaxConnLifetimeSeconds: 3600,
			},
			Connectors: []trustreadydconfig.ConnectorConfig{
				{
					Provider: "slack",
					Protocol: "oauth2",
					RawConfig: trustreadydconfig.ConnectorConfigOAuth2{
						ClientID:     "client-id",
						ClientSecret: "client-secret",
					},
					RawSettings: map[string]any{
						"signing-secret": "secret",
					},
				},
			},
		},
	}

	err := WriteConfig(cfg, configPath, FormatYAML)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var loaded trustreadydconfig.FullConfig

	err = yaml.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, cfg.Unit.Metrics.Addr, loaded.Unit.Metrics.Addr)
	assert.Equal(t, cfg.Unit.Tracing.MaxBatchSize, loaded.Unit.Tracing.MaxBatchSize)
	assert.Equal(t, cfg.TrustReadyd.Api.Cors.AllowedOrigins, loaded.TrustReadyd.Api.Cors.AllowedOrigins)
	assert.Equal(t, cfg.TrustReadyd.Pg.PoolSize, loaded.TrustReadyd.Pg.PoolSize)
	assert.Equal(t, cfg.TrustReadyd.Pg.MinPoolSize, loaded.TrustReadyd.Pg.MinPoolSize)
	assert.Equal(t, cfg.TrustReadyd.Pg.MaxConnIdleTimeSeconds, loaded.TrustReadyd.Pg.MaxConnIdleTimeSeconds)
	assert.Equal(t, cfg.TrustReadyd.Pg.MaxConnLifetimeSeconds, loaded.TrustReadyd.Pg.MaxConnLifetimeSeconds)
	require.Len(t, loaded.TrustReadyd.Connectors, 1)
	assert.Equal(t, "SLACK", loaded.TrustReadyd.Connectors[0].Provider)
}

func TestWriteConfig_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "trustreadyd.json")

	cfg := &trustreadydconfig.FullConfig{
		Unit: trustreadydconfig.UnitConfig{
			Metrics: trustreadydconfig.MetricsConfig{Addr: "localhost:9090"},
		},
		TrustReadyd: trustreadydconfig.Config{
			BaseURL:       "http://localhost:8080",
			EncryptionKey: "",
		},
	}

	err := WriteConfig(cfg, configPath, FormatJSON)
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var tree map[string]any

	err = json.Unmarshal(data, &tree)
	require.NoError(t, err)

	trustreadyd, ok := tree["trustreadyd"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "http://localhost:8080", trustreadyd["base-url"])
	assert.Equal(t, "", trustreadyd["encryption-key"])

	var loaded trustreadydconfig.FullConfig

	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, cfg.Unit.Metrics.Addr, loaded.Unit.Metrics.Addr)
	assert.Equal(t, cfg.TrustReadyd.BaseURL, loaded.TrustReadyd.BaseURL)
}

func TestWriteConfig_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "trustreadyd.txt")

	cfg := &trustreadydconfig.FullConfig{}

	err := WriteConfig(cfg, configPath, Format("toml"))
	require.Error(t, err)
}
