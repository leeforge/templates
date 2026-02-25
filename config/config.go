package config

import (
	"fmt"
	"net"
	"strings"
)

// Config is the example-host runtime configuration surface aligned to backend sections.
type Config struct {
	Server        *ServerConfig        `mapstructure:"server"`
	Database      *DatabaseConfig      `mapstructure:"database"`
	Cache         *CacheConfig         `mapstructure:"cache"`
	Log           *LogConfig           `mapstructure:"log"`
	Tracing       *TracingConfig       `mapstructure:"tracing"`
	Metrics       *MetricsConfig       `mapstructure:"metrics"`
	Security      *SecurityConfig      `mapstructure:"security"`
	AccessControl *AccessControlConfig `mapstructure:"access_control"`
	Plugins       *PluginsConfig       `mapstructure:"plugins"`
	Init          *InitConfig          `mapstructure:"init"`
	Frontend      *FrontendConfig      `mapstructure:"frontend"`
	Captcha       *CaptchaConfig       `mapstructure:"captcha"`
}

type ServerConfig struct {
	Port string     `mapstructure:"port"`
	Mode string     `mapstructure:"mode"`
	CORS CORSConfig `mapstructure:"cors"`
}

type CORSConfig struct {
	Enabled          bool     `mapstructure:"enabled"`
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`

	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
	Params   string `mapstructure:"params"`

	AutoMigrate  bool `mapstructure:"auto_migrate"`
	MaxOpenConns int  `mapstructure:"max_open_conns"`
	MaxIdleConns int  `mapstructure:"max_idle_conns"`
}

func (c DatabaseConfig) DSN() string {
	if c.URL != "" {
		return c.URL
	}

	port := c.Port
	if port == "" {
		port = "5432"
	}

	sslmode := c.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}

	hostPort := c.Host
	if port != "" {
		hostPort = net.JoinHostPort(c.Host, port)
	}

	auth := ""
	if c.Username != "" {
		if c.Password != "" {
			auth = fmt.Sprintf("%s:%s@", c.Username, c.Password)
		} else {
			auth = fmt.Sprintf("%s@", c.Username)
		}
	}

	query := fmt.Sprintf("sslmode=%s", sslmode)
	if c.Params != "" {
		params := strings.TrimSpace(c.Params)
		params = strings.TrimPrefix(params, "?")
		params = strings.TrimPrefix(params, "&")
		if params != "" {
			query = fmt.Sprintf("%s&%s", query, params)
		}
	}

	return fmt.Sprintf("postgres://%s%s/%s?%s", auth, hostPort, c.Name, query)
}

func (c DatabaseConfig) Validate() error {
	if c.URL != "" {
		return nil
	}
	if c.Name == "" {
		return fmt.Errorf("database.name cannot be empty")
	}
	if c.Host == "" {
		return fmt.Errorf("database.host cannot be empty")
	}
	if c.Port == "" {
		return fmt.Errorf("database.port cannot be empty")
	}
	return nil
}

type CacheConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (c *CacheConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type LogConfig struct {
	Director       string `mapstructure:"director"`
	MessageKey     string `mapstructure:"message-key"`
	LevelKey       string `mapstructure:"level-key"`
	TimeKey        string `mapstructure:"time-key"`
	NameKey        string `mapstructure:"name-key"`
	CallerKey      string `mapstructure:"caller-key"`
	LineEnding     string `mapstructure:"line-ending"`
	StacktraceKey  string `mapstructure:"stacktrace-key"`
	Level          string `mapstructure:"level"`
	EncodeLevel    string `mapstructure:"encode-level"`
	Prefix         string `mapstructure:"prefix"`
	TimeFormat     string `mapstructure:"time-format"`
	Format         string `mapstructure:"format"`
	LogInTerminal  bool   `mapstructure:"log-in-terminal"`
	MaxAge         int    `mapstructure:"max-age"`
	MaxSize        int    `mapstructure:"max-size"`
	MaxBackups     int    `mapstructure:"max-backups"`
	Compress       bool   `mapstructure:"compress"`
	ShowLineNumber bool   `mapstructure:"show-line-number"`
}

type TracingConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
}

type MetricsConfig struct {
	Port string `mapstructure:"port"`
}

type SecurityConfig struct {
	JWTSecret       string       `mapstructure:"jwt_secret"`
	TokenExpiry     int          `mapstructure:"token_expiry"`
	RefreshExpiry   int          `mapstructure:"refresh_expiry"`
	PasswordCost    int          `mapstructure:"password_cost"`
	EnableRateLimit bool         `mapstructure:"enable_rate_limit"`
	RateLimit       int          `mapstructure:"rate_limit"`
	Cookie          CookieConfig `mapstructure:"cookie"`
}

type CookieConfig struct {
	Secure   bool   `mapstructure:"secure"`
	SameSite string `mapstructure:"same_site"`
	Domain   string `mapstructure:"domain"`
	Path     string `mapstructure:"path"`
}

type AccessControlConfig struct {
	MultiTenancy MultiTenancyConfig `mapstructure:"multi_tenancy"`
	Project      ProjectConfig      `mapstructure:"project"`
	Domain       DomainConfig       `mapstructure:"domain"`
	ABAC         FeatureToggle      `mapstructure:"abac"`
	Share        FeatureToggle      `mapstructure:"share"`
	Quota        FeatureToggle      `mapstructure:"quota"`
	Audit        FeatureToggle      `mapstructure:"audit"`
}

type DomainConfig struct {
	Mode            string `mapstructure:"mode"`
	DefaultDomainID string `mapstructure:"default_domain_id"`
}

type MultiTenancyConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	DefaultTenantID string `mapstructure:"default_tenant_id"`
}

type ProjectConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	DefaultProjectID string `mapstructure:"default_project_id"`
	RequireProject   bool   `mapstructure:"require_project"`
}

type PluginsConfig struct {
	Tenant FeatureToggle `mapstructure:"tenant"`
	OU     FeatureToggle `mapstructure:"ou"`
}

type FeatureToggle struct {
	Enabled bool `mapstructure:"enabled"`
}

type InitConfig struct {
	SecretKey string `mapstructure:"secret_key"`
}

type FrontendConfig struct {
	URL string `mapstructure:"url"`
}

type CaptchaConfig struct {
	Enabled        bool       `mapstructure:"enabled"`
	TTL            string     `mapstructure:"ttl"`
	GenerateLimit  int        `mapstructure:"generate_limit"`
	GenerateWindow string     `mapstructure:"generate_window"`
	MaxAttempts    int        `mapstructure:"max_attempts"`
	AttemptWindow  string     `mapstructure:"attempt_window"`
	Math           MathConfig `mapstructure:"math"`
}

type MathConfig struct {
	Width           int `mapstructure:"width"`
	Height          int `mapstructure:"height"`
	NoiseCount      int `mapstructure:"noise_count"`
	ShowLineOptions int `mapstructure:"show_line_options"`
}

func Default() *Config {
	return &Config{
		Server: &ServerConfig{
			Port: "8080",
			Mode: "release",
			CORS: CORSConfig{
				Enabled:          true,
				AllowedOrigins:   []string{"http://localhost:3000"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
				ExposedHeaders:   []string{"Link"},
				AllowCredentials: true,
				MaxAge:           300,
			},
		},
		Database: &DatabaseConfig{
			Host:         "localhost",
			Port:         "15436",
			Username:     "",
			Password:     "",
			Name:         "postgres",
			SSLMode:      "disable",
			AutoMigrate:  true,
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		},
		Cache: &CacheConfig{
			Host:     "127.0.0.1",
			Port:     "16379",
			Password: "",
			DB:       0,
		},
		Log: &LogConfig{
			Director:       "logs",
			Level:          "info",
			Format:         "console",
			LogInTerminal:  true,
			ShowLineNumber: true,
			MaxAge:         7,
			MaxSize:        100,
			MaxBackups:     10,
			Compress:       true,
			TimeFormat:     "2006/01/02 - 15:04:05",
			EncodeLevel:    "LowercaseColorLevelEncoder",
		},
		Tracing: &TracingConfig{
			Enabled:  false,
			Endpoint: "localhost:4317",
		},
		Metrics: &MetricsConfig{
			Port: "9090",
		},
		Security: &SecurityConfig{
			JWTSecret:       "",
			TokenExpiry:     24,
			RefreshExpiry:   72,
			PasswordCost:    12,
			EnableRateLimit: true,
			RateLimit:       60,
			Cookie: CookieConfig{
				Secure:   false,
				SameSite: "lax",
				Domain:   "",
				Path:     "/api/v1/auth",
			},
		},
		AccessControl: &AccessControlConfig{
			MultiTenancy: MultiTenancyConfig{Enabled: true},
			Project:      ProjectConfig{Enabled: true},
			Domain:       DomainConfig{Mode: "domain"},
			ABAC:         FeatureToggle{Enabled: false},
			Share:        FeatureToggle{Enabled: false},
			Quota:        FeatureToggle{Enabled: false},
			Audit:        FeatureToggle{Enabled: false},
		},
		Plugins: &PluginsConfig{
			Tenant: FeatureToggle{Enabled: true},
			OU:     FeatureToggle{Enabled: true},
		},
		Init: &InitConfig{
			SecretKey: "",
		},
		Frontend: &FrontendConfig{
			URL: "http://localhost:3000",
		},
		Captcha: &CaptchaConfig{
			Enabled:        false,
			TTL:            "5m",
			GenerateLimit:  10,
			GenerateWindow: "1m",
			MaxAttempts:    5,
			AttemptWindow:  "5m",
			Math: MathConfig{
				Width:           240,
				Height:          80,
				NoiseCount:      5,
				ShowLineOptions: 2,
			},
		},
	}
}
