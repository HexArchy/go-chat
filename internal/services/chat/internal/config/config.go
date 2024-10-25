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
	Engines          EnginesConfig     `koanf:"engines"`
	Logging          LoggingConfig     `koanf:"logging"`
	Handlers         HandlersConfig    `koanf:"handlers"`
	AuthService      AuthServiceConfig `koanf:"auth_service"`
	WebsiteService   ServiceConfig     `koanf:"website_service"`
	Vault            VaultConfig       `koanf:"vault"`
	GracefulShutdown time.Duration     `koanf:"graceful_shutdown"`
	WebSocket        WebSocketConfig   `koanf:"websocket"`
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

type LoggingConfig struct {
	Level string `koanf:"level"`
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

type AuthServiceConfig struct {
	Address      string `koanf:"address"`
	ServiceToken string `koanf:"service_token"`
}

type ServiceConfig struct {
	Address      string `koanf:"address"`
	ServiceToken string `koanf:"service_token"`
}

type WebSocketConfig struct {
	PingInterval     time.Duration `koanf:"ping_interval"`
	PongWait         time.Duration `koanf:"pong_wait"`
	MaxMessageSize   int64         `koanf:"max_message_size"`
	WriteWait        time.Duration `koanf:"write_wait"`
	MessageQueueSize int           `koanf:"message_queue_size"`
}

type VaultConfig struct {
	Address string        `koanf:"address"`
	Token   string        `koanf:"token"`
	Path    string        `koanf:"path"`
	Timeout time.Duration `koanf:"timeout"`
}

func LoadConfig(configPath string) (*Config, error) {
	if err := loadDefaults(); err != nil {
		return nil, errors.Wrap(err, "load defaults")
	}

	log.Printf("Loading config from file: %s", configPath)
	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		log.Printf("Error loading from YAML file: %v", err)
	}

	if err := k.Load(env.Provider("CHAT_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "CHAT_")), "_", ".", -1)
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

		if serviceToken := k.String("vault.data.service_token"); serviceToken != "" {
			k.Set("auth_service.service_token", serviceToken)
			k.Set("website_service.service_token", serviceToken)
		}
	}

	var config Config
	if err := k.Unmarshal("", &config); err != nil {
		return nil, errors.Wrap(err, "unmarshal config")
	}

	// Проверяем обязательные значения
	if config.AuthService.ServiceToken == "" {
		return nil, errors.New("auth service token is not set after loading")
	}
	if config.WebsiteService.ServiceToken == "" {
		return nil, errors.New("website service token is not set after loading")
	}

	return &config, nil
}

func loadDefaults() error {
	defaults := map[string]interface{}{
		"engines.storage.max_open_conns":    10,
		"engines.storage.max_idle_conns":    5,
		"engines.storage.conn_max_lifetime": time.Hour,
		"logging.level":                     "info",
		"handlers.http.read_timeout":        10 * time.Second,
		"handlers.http.write_timeout":       10 * time.Second,
		"handlers.http.address":             "localhost",
		"handlers.http.port":                "8082",
		"handlers.grpc.address":             "localhost",
		"handlers.grpc.port":                "9092",
		"auth_service.address":              "localhost:9090",
		"website_service.address":           "localhost:9091",
		"websocket.ping_interval":           30 * time.Second,
		"websocket.pong_wait":               60 * time.Second,
		"websocket.max_message_size":        4096, // 4KB
		"websocket.write_wait":              10 * time.Second,
		"websocket.message_queue_size":      256,
		"vault.timeout":                     5 * time.Minute,
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
