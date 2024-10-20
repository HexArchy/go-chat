package config

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/vault"
	"github.com/pkg/errors"
)

var k = koanf.New(".")

type Config struct {
	Engines          EnginesConfig  `koanf:"engines"`
	Auth             AuthConfig     `koanf:"auth"`
	Logging          LoggingConfig  `koanf:"logging"`
	Vault            VaultConfig    `koanf:"vault"`
	Handlers         HandlersConfig `koanf:"handlers"`
	GracefulShutdown time.Duration  `koanf:"graceful_shutdown"`
}

type EnginesConfig struct {
	Storage StorageConfig `koanf:"storage"`
}

type StorageConfig struct {
	URL             string        `koanf:"url"`
	MaxOpenConns    int           `koanf:"max_open_conns"`
	MaxIdleConns    int           `koanf:"max_idle_conns"`
	ConnMaxLifetime time.Duration `koanf:"conn_max_lifetime"`
}

type VaultConfig struct {
	Address string        `koanf:"address"`
	Token   string        `koanf:"token"`
	Path    string        `koanf:"path"`
	Timeout time.Duration `koanf:"timeout"`
}

type AuthConfig struct {
	JWT JWTConfig `koanf:"jwt"`
}

type JWTConfig struct {
	AccessSecret       string        `koanf:"access_secret"`
	RefreshSecret      string        `koanf:"refresh_secret"`
	AccessExpiryHours  time.Duration `koanf:"access_expiry_hours"`
	RefreshExpiryHours time.Duration `koanf:"refresh_expiry_hours"`
}

type HandlersConfig struct {
	HTTP HTTPConfig `koanf:"http"`
	GRPC GRPCConfig `koanf:"grpc"`
}

type HTTPConfig struct {
	ReadTimeout  time.Duration `koanf:"read_timeout"`
	WriteTimeout time.Duration `koanf:"write_timeout"`
	Address      string        `koanf:"address"`
	Port         string        `koanf:"port"`
}

type GRPCConfig struct {
	Address string `koanf:"address"`
	Port    string `koanf:"port"`
}

type LoggingConfig struct {
	Level string `koanf:"level"`
}

func LoadConfig(configPath string) (*Config, error) {
	if err := loadDefaults(); err != nil {
		return nil, errors.Wrap(err, "load defaults")
	}

	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		log.Printf("Error loading from YAML file: %v", err)
	}

	if err := k.Load(env.Provider("AUTH_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "AUTH_")), "_", ".", -1)
	}), nil); err != nil {
		return nil, errors.Wrap(err, "loading environment variables")
	}

	if k.Exists("vault.address") && k.Exists("vault.token") {
		vaultProvider := vault.Provider(
			vault.Config{
				Address: k.String("vault.address"),
				Token:   k.String("vault.token"),
				Path:    k.String("vault.path"),
				Timeout: k.Duration("vault.timeout"),
			},
		)
		if err := k.Load(vaultProvider, nil); err != nil {
			log.Printf("Error loading secrets from Vault: %v", err)
		}

		if accessSecret := k.String("vault.data.access_secret"); accessSecret != "" {
			k.Set("auth.jwt.access_secret", accessSecret)
		}
		if refreshSecret := k.String("vault.data.refresh_secret"); refreshSecret != "" {
			k.Set("auth.jwt.refresh_secret", refreshSecret)
		}
	}

	var config Config
	if err := k.Unmarshal("", &config); err != nil {
		return nil, errors.Wrap(err, "unmarshal config")
	}

	return &config, nil
}

func loadDefaults() error {
	defaults := map[string]interface{}{
		"engines.storage.max_open_conns":    10,
		"engines.storage.max_idle_conns":    5,
		"engines.storage.conn_max_lifetime": time.Hour,
		"vault.timeout":                     5 * time.Minute,
		"auth.jwt.access_expiry_hours":      1,
		"auth.jwt.refresh_expiry_hours":     168,
		"logging.level":                     "info",
		"handlers.http.read_timeout":        10 * time.Second,
		"handlers.http.write_timeout":       10 * time.Second,
		"handlers.http.address":             "localhost",
		"handlers.http.port":                "8080",
		"handlers.grpc.address":             "localhost",
		"handlers.grpc.port":                "9090",
		"graceful_shutdown":                 15 * time.Second,
	}

	return k.Load(confmap.Provider(defaults, "."), nil)
}

func (h *HTTPConfig) FullAddress() string {
	return net.JoinHostPort(h.Address, h.Port)
}

func (g *GRPCConfig) FullAddress() string {
	return net.JoinHostPort(g.Address, g.Port)
}
