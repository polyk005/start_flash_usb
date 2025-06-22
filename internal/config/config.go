package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Install  InstallConfig  `mapstructure:"install"`
	Database DatabaseConfig `mapstructure:"database"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
	ConfigPath  string `mapstructure:"config_path"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	File   string `mapstructure:"file"`
}

type InstallConfig struct {
	Timeout         time.Duration `mapstructure:"timeout"`
	MaxConcurrency  int           `mapstructure:"max_concurrency"`
	InstallBasePath string        `mapstructure:"install_base_path"`
	AppsConfigPath  string        `mapstructure:"apps_config_path"`
	AppsDirectory   string        `mapstructure:"apps_directory"`
}

type DatabaseConfig struct {
	Type string `mapstructure:"type"`
	Path string `mapstructure:"path"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	viper.SetDefault("app.name", "StartFlash")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.config_path", "./configs")

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "text")

	viper.SetDefault("install.timeout", "10m")
	viper.SetDefault("install.max_concurrency", 3)
	viper.SetDefault("install.apps_config_path", "./configs/apps.json")
	viper.SetDefault("install.apps_directory", "./apps")

	viper.SetDefault("database.type", "memory")
}

func validateConfig(config *Config) error {
	if config.App.Name == "" {
		return fmt.Errorf("app name is required")
	}

	if config.Install.Timeout <= 0 {
		return fmt.Errorf("install timeout must be positive")
	}

	if config.Install.MaxConcurrency <= 0 {
		return fmt.Errorf("max concurrency must be positive")
	}

	return nil
}

func (c *Config) GetString(key string) string {
	return viper.GetString(key)
}

func (c *Config) GetInt(key string) int {
	return viper.GetInt(key)
}

func (c *Config) GetDuration(key string) time.Duration {
	return viper.GetDuration(key)
}

func (c *Config) GetBool(key string) bool {
	return viper.GetBool(key)
} 
