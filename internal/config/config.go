package config

import (
	"fmt"
	"os"
	"strconv"

	commonconfig "github.com/mysunshines/gocommon/config"
	"gopkg.in/yaml.v3"
)

// 共享类型别名 — 直接复用 common/config，零映射开销
type (
	AppConfig       = commonconfig.AppConfig
	DatabaseConfig  = commonconfig.DatabaseConfig
	RedisConfig     = commonconfig.RedisConfig
	JWTConfig       = commonconfig.JWTConfig
	GRPCConfig      = commonconfig.GRPCConfig
	HTTPConfig      = commonconfig.HTTPConfig
	ConsulConfig    = commonconfig.ConsulConfig
	MetricsConfig   = commonconfig.MetricsConfig
	RateLimitConfig = commonconfig.RateLimitConfig
)

// Config 文章服务配置
type Config struct {
	App       AppConfig       `yaml:"app"`
	HTTP      HTTPConfig      `yaml:"http"`
	GRPC      GRPCConfig      `yaml:"grpc"`
	Database  DatabaseConfig  `yaml:"database"`
	Redis     RedisConfig     `yaml:"redis"`
	JWT       JWTConfig       `yaml:"jwt"`
	Consul    ConsulConfig    `yaml:"consul"`
	Metrics   MetricsConfig   `yaml:"metrics"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	CORS      CORSConfig      `yaml:"cors"`
}

// CORSConfig 跨域配置（article-service 独有）
type CORSConfig struct {
	Enabled          bool     `yaml:"enabled"`
	AllowOrigins     []string `yaml:"allow_origins"`
	AllowMethods     []string `yaml:"allow_methods"`
	AllowHeaders     []string `yaml:"allow_headers"`
	ExposeHeaders    []string `yaml:"expose_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

var globalConfig *Config

// Load 加载配置文件
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	setDefaults(&c)
	applyEnvOverrides(&c)
	globalConfig = &c
	return &c, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

func setDefaults(c *Config) {
	if c.App.LogLevel == "" {
		c.App.LogLevel = "info"
	}
	if c.App.Host == "" {
		c.App.Host = "0.0.0.0"
	}
	if c.HTTP.Port == 0 {
		c.HTTP.Port = 8082
	}
	if c.GRPC.Port == 0 {
		c.GRPC.Port = 9002
	}
	if c.Metrics.Port == 0 {
		c.Metrics.Port = 9092
	}
	if c.Metrics.Path == "" {
		c.Metrics.Path = "/metrics"
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 100
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 10
	}
	if c.Database.ConnMaxLifetime == 0 {
		c.Database.ConnMaxLifetime = 3600
	}
	if c.Redis.PoolSize == 0 {
		c.Redis.PoolSize = 100
	}
	if c.Consul.CheckInterval == 0 {
		c.Consul.CheckInterval = 10
	}
	if c.Consul.DeregisterCritical == 0 {
		c.Consul.DeregisterCritical = 30
	}
	if c.RateLimit.QPS == 0 {
		c.RateLimit.QPS = 1000
	}
	if c.RateLimit.Burst == 0 {
		c.RateLimit.Burst = 2000
	}
	if c.CORS.MaxAge == 0 {
		c.CORS.MaxAge = 86400
	}
}

func applyEnvOverrides(c *Config) {
	if v := os.Getenv("DB_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Database.Port = port
		}
	}
	if v := os.Getenv("DB_USER"); v != "" {
		c.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		c.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		c.Database.Name = v
	}
	if v := os.Getenv("REDIS_HOST"); v != "" {
		c.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Redis.Port = port
		}
	}
	if v := os.Getenv("CONSUL_ADDRESS"); v != "" {
		c.Consul.Address = v
	}
	if v := os.Getenv("GRPC_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.GRPC.Port = port
		}
	}
}
