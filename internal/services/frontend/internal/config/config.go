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
	"github.com/pkg/errors"
)

var k = koanf.New(".")

type Config struct {
	Engines          EnginesConfig  `koanf:"engines"`
	Logging          LoggingConfig  `koanf:"logging"`
	Handlers         HandlersConfig `koanf:"handlers"`
	AuthService      ServiceConfig  `koanf:"auth_service"`
	WebsiteService   ServiceConfig  `koanf:"website_service"`
	ChatService      ServiceConfig  `koanf:"chat_service"`
	Session          SessionConfig  `koanf:"session"`
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

type LoggingConfig struct {
	Level string `koanf:"level"`
}

type HandlersConfig struct {
	HTTP HTTPConfig `koanf:"http"`
}

type HTTPConfig struct {
	ReadTimeout   time.Duration `koanf:"read_timeout"`
	WriteTimeout  time.Duration `koanf:"write_timeout"`
	Address       string        `koanf:"address"`
	Port          string        `koanf:"port"`
	TemplatesPath string        `koanf:"templates_path"`
}

type ServiceConfig struct {
	Address      string `koanf:"address"`
	ServiceToken string `koanf:"service_token"`
}

type SessionConfig struct {
	Secret string        `koanf:"secret"`
	MaxAge time.Duration `koanf:"max_age"`
}

func LoadConfig(configPath string) (*Config, error) {
	if err := loadDefaults(); err != nil {
		return nil, errors.Wrap(err, "load defaults")
	}

	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		log.Printf("Error loading from YAML file: %v", err)
	}

	if err := k.Load(env.Provider("FRONTEND_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "FRONTEND_")), "_", ".", -1)
	}), nil); err != nil {
		return nil, errors.Wrap(err, "loading environment variables")
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
		"logging.level":                     "info",
		"handlers.http.read_timeout":        10 * time.Second,
		"handlers.http.write_timeout":       10 * time.Second,
		"handlers.http.address":             "localhost",
		"handlers.http.port":                "8084",
		"handlers.http.templates_path":      "templates",
		"handlers.http.static_path":         "static",
		"auth_service.address":              "localhost:9090",
		"website_service.address":           "localhost:9091",
		"chat_service.address":              "localhost:9092",
		"session.max_age":                   24 * time.Hour,
		"vault.timeout":                     5 * time.Minute,
		"graceful_shutdown":                 15 * time.Second,
	}

	return k.Load(confmap.Provider(defaults, "."), nil)
}

func (h *HTTPConfig) FullAddress() string {
	return net.JoinHostPort(h.Address, h.Port)
}
