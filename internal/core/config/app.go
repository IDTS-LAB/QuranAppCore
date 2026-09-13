package config

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// Config reflects the nested structure of your config.toml
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	S3       S3Config       `mapstructure:"s3"`
}

type AppConfig struct {
	Port string `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type DatabaseConfig struct {
	URL          string `mapstructure:"url"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// JWTConfig holds JWT settings. Durations are Go duration strings
// (e.g. "15m", "168h"). Every field can be overridden by environment:
// JWT_SECRET, JWT_ISSUER, JWT_ACCESS_TTL, JWT_REFRESH_TTL.
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	Issuer     string `mapstructure:"issuer"`
	AccessTTL  string `mapstructure:"access_ttl"`
	RefreshTTL string `mapstructure:"refresh_ttl"`
}

// AccessDuration parses AccessTTL ("15m") for access tokens.
func (c JWTConfig) AccessDuration() (time.Duration, error) {
	return time.ParseDuration(c.AccessTTL)
}

// RefreshDuration parses RefreshTTL ("168h") for refresh tokens.
func (c JWTConfig) RefreshDuration() (time.Duration, error) {
	return time.ParseDuration(c.RefreshTTL)
}

// S3Config holds S3-compatible object storage settings. Works against AWS S3
// itself or an S3-compatible server like the dev MinIO in
// docker-compose.dev.yml (path-style addressing + custom endpoint).
// Every field can be overridden by environment:
// S3_ENDPOINT, S3_REGION, S3_ACCESS_KEY, S3_SECRET_KEY, S3_BUCKET,
// S3_USE_PATH_STYLE, S3_PRESIGN_TTL.
type S3Config struct {
	Endpoint     string `mapstructure:"endpoint"`
	Region       string `mapstructure:"region"`
	AccessKey    string `mapstructure:"access_key"`
	SecretKey    string `mapstructure:"secret_key"`
	Bucket       string `mapstructure:"bucket"`
	UsePathStyle bool   `mapstructure:"use_path_style"`
	PresignTTL   string `mapstructure:"presign_ttl"`
}

// PresignDuration parses PresignTTL ("15m") for presigned download URLs.
func (c S3Config) PresignDuration() (time.Duration, error) {
	return time.ParseDuration(c.PresignTTL)
}

var (
	instance *Config
	once     sync.Once
)

// GetConfig initializes and returns the Singleton instance of the configuration
func GetConfig() *Config {
	once.Do(func() {
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
		viper.AddConfigPath(".") // Search in the root directory

		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Failed to read config.toml via Viper: %v", err)
		}

		instance = &Config{}
		if err := viper.Unmarshal(instance); err != nil {
			log.Fatalf("Failed to unmarshal TOML configuration: %v", err)
		}

		log.Println("TOML Configuration successfully loaded into memory as a Singleton!")
	})

	return instance
}
