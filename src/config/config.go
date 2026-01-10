package config

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/gorilla/securecookie"
	"github.com/spf13/viper"
)

type Configuration struct {
	Server       ServerConfiguration
	Database     DatabaseConfiguration
	Mail         MailConfiguration
	JWT          JWTConfiguration
	OAuth2       map[string]OAuth2Configuration
}

type ServerConfiguration struct {
	Port int
	StaticPath string
}

type DatabaseConfiguration struct {
	Type     string
	Host     string
	Port     int
	Schema   string
	User     string
	Password string
}

type MailConfiguration struct {
	Host     string
	Port     int
	User     string
	Password string
}

type JWTConfiguration struct {
	Secret      string
	ExpiryHours int
}

type OAuth2Configuration struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

var (
       cfg          Configuration
       once         sync.Once
)

func Init() {
       once.Do(load)
}

func load() {
	ReadConfig(&cfg)
}

func ReadConfig(configuration *Configuration) {
	// Set the path to look for the configurations file
	viper.AddConfigPath("./config")

	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()

	// Config file to read
	viper.SetConfigType("yaml")
	viper.SetDefault("CONFIG_FILE", "local")
	file := viper.GetString("CONFIG_FILE")
	slog.Info("Config reading", slog.String("file", file))
	viper.SetConfigName(file)

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// Set undefined variables
	viper.SetDefault("database.type", "mysql")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "3306")
	viper.SetDefault("mail.host", "localhost")
	viper.SetDefault("mail.port", "25")
	viper.SetDefault("server.staticpath", "./static")
	viper.SetDefault("jwt.expiryhours", 8)

	err := viper.Unmarshal(configuration)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	os.Setenv("SESSION_KEY", hex.EncodeToString(securecookie.GenerateRandomKey(32)))
	slog.Debug("Config", slog.String("SESSION_KEY", os.Getenv("SESSION_KEY")))
}

func Get() *Configuration {
	return &cfg
}

func GetDatabase() *DatabaseConfiguration {
	return &cfg.Database
}

func GetServer() *ServerConfiguration {
	return &cfg.Server
}

func GetMail() *MailConfiguration {
	return &cfg.Mail
}

func GetJWT() *JWTConfiguration {
	return &cfg.JWT
}

// GetJWTSecret returns JWT secret from env or config
func GetJWTSecret() []byte {
	// Check environment variable first
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return []byte(secret)
	}
	
	// Fall back to config
	if cfg.JWT.Secret != "" && cfg.JWT.Secret != "${JWT_SECRET}" {
		return []byte(cfg.JWT.Secret)
	}
	
	// Generate and warn about default key (dev only)
	slog.Warn("Using default JWT secret - SET JWT_SECRET environment variable in production!")
	defaultKey := securecookie.GenerateRandomKey(32)
	return defaultKey
}

/*
func GetOAuth2() *map[string]OAuth2Configuration {
	return &cfg.OAuth2
}
*/
