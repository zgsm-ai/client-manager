package internal

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	ListenAddr string
	NoRedis    bool
	ConfigPath string
}

// AppConfig holds the global application configuration
var AppConfig = &Config{}

// InitConfig initializes the configuration
func InitConfig(rootCmd *cobra.Command) error {
	// Add command line flags
	rootCmd.Flags().StringVarP(&AppConfig.ListenAddr, "listen", "l", "", "Server listen address (e.g. :8080)")
	rootCmd.Flags().BoolVar(&AppConfig.NoRedis, "no-redis", false, "Disable Redis cache")
	rootCmd.Flags().StringVarP(&AppConfig.ConfigPath, "config", "c", "", "Configuration file path")

	return nil
}

// LoadConfig loads configuration from file and environment variables
// @returns {error} Error if configuration loading fails
// @description
// - Loads configuration from config.yaml file
// - Merges environment variables
// - Sets default values for missing configurations
// @throws
// - Configuration file not found error
// - Configuration parsing error
func LoadConfig(configPath string) error {
	// If custom config path is provided, use it
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./config")
		viper.AddConfigPath(".")
	}

	// Set default values
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("database.dsn", "client-manager.db")
	viper.SetDefault("redis.enabled", true)
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("log.level", "info")

	// Enable environment variable override
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; create default config
			return CreateDefaultConfig()
		}
		return err
	}

	return nil
}

// CreateDefaultConfig creates default configuration file
// @returns {error} Error if config file creation fails
// @description
// - Creates a default config.yaml file
// - Sets default values for all configuration options
// @throws
// - File creation error
// - File write error
func CreateDefaultConfig() error {
	config := `server:
  port: "8080"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 60s

database:
  dsn: "client-manager.db"
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600s

redis:
  enabled: true
  addr: "localhost:6379"
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5

log:
  level: "info"
  format: "json"
  output: "stdout"

cache:
  default_ttl: 300s
  cleanup_interval: 600s

metrics:
  enabled: true
  path: "/metrics"

swagger:
  enabled: true
  path: "/swagger"
`

	// Create config directory if it doesn't exist
	if err := os.MkdirAll("config", 0755); err != nil {
		return err
	}

	// Write default config file
	return os.WriteFile("config/config.yaml", []byte(config), 0644)
}

// ApplyConfig applies command line overrides to the configuration
func ApplyConfig(logger *logrus.Logger) {
	// Override listen address from command line if provided
	if AppConfig.ListenAddr != "" {
		viper.Set("server.port", AppConfig.ListenAddr)
	}

	// Override Redis settings if no-redis flag is set
	if AppConfig.NoRedis {
		viper.Set("redis.enabled", false)
		logger.Info("Redis is disabled by command line flag")
	}
}

// GetServerPort returns the server port from configuration
func GetServerPort() string {
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}
	return port
}

// IsRedisEnabled returns whether Redis is enabled in configuration
func IsRedisEnabled() bool {
	return !AppConfig.NoRedis && viper.GetBool("redis.enabled")
}
