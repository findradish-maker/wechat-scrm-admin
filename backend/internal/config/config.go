package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppPort string
	AppEnv  string
	MySQL   MySQLConfig
	Redis   RedisConfig
	Wechat  WechatConfig
	AI      AIConfig
	JWT     JWTConfig
	Seed    SeedConfig
}

type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type WechatConfig struct {
	BaseURL              string
	DefaultLoginPlatform string
}

type AIConfig struct {
	DeepSeek AIProviderConfig
	Claude   AIProviderConfig
	OpenAI   AIProviderConfig
	Relay    AIProviderConfig
}

type AIProviderConfig struct {
	APIKey       string
	BaseURL      string
	DefaultModel string
}

type JWTConfig struct {
	Secret      string
	ExpireHours int
}

type SeedConfig struct {
	AdminUsername string
	AdminPassword string
	AdminName     string
}

func Load() Config {
	return Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "127.0.0.1"),
			Port:     getEnvAsInt("MYSQL_PORT", 3306),
			User:     getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", "root123."),
			Database: getEnv("MYSQL_DATABASE", "wechat_enterprise"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
			Password: getEnv("REDIS_PASSWORD", "123456"),
			DB:       getEnvAsInt("REDIS_DB", 2),
		},
		Wechat: WechatConfig{
			BaseURL:              strings.TrimRight(getEnv("WECHAT_REAL_BASE_URL", "http://127.0.0.1:8062/api"), "/"),
			DefaultLoginPlatform: getEnv("WECHAT_DEFAULT_LOGIN_PLATFORM", "ipad"),
		},
		AI: AIConfig{
			DeepSeek: AIProviderConfig{
				APIKey:       getEnv("DEEPSEEK_API_KEY", "sk-c904a****38"),
				BaseURL:      strings.TrimRight(getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"), "/"),
				DefaultModel: getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
			},
			Claude: AIProviderConfig{
				APIKey:       getEnv("CLAUDE_API_KEY", ""),
				BaseURL:      strings.TrimRight(getEnv("CLAUDE_BASE_URL", "https://api.anthropic.com/v1"), "/"),
				DefaultModel: getEnv("CLAUDE_MODEL", "claude-sonnet-4-20250514"),
			},
			OpenAI: AIProviderConfig{
				APIKey:       getEnv("OPENAI_API_KEY", ""),
				BaseURL:      strings.TrimRight(getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"), "/"),
				DefaultModel: getEnv("OPENAI_MODEL", "gpt-5-mini"),
			},
			Relay: AIProviderConfig{
				APIKey:       getEnv("RELAY_API_KEY", ""),
				BaseURL:      strings.TrimRight(getEnv("RELAY_BASE_URL", ""), "/"),
				DefaultModel: getEnv("RELAY_MODEL", "gpt-4o-mini"),
			},
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "change-me-in-production"),
			ExpireHours: getEnvAsInt("JWT_EXPIRE_HOURS", 72),
		},
		Seed: SeedConfig{
			AdminUsername: getEnv("SEED_ADMIN_USERNAME", "admin"),
			AdminPassword: getEnv("SEED_ADMIN_PASSWORD", "admin123456"),
			AdminName:     getEnv("SEED_ADMIN_NAME", "系统管理员"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	raw := getEnv(key, "")
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
