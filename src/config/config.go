package config

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/spf13/viper"
)

type Configuration struct {
	Server       ServerConfiguration
	Database     DatabaseConfiguration
	Mail         MailConfiguration
	OAuth2       map[string]OAuth2Configuration
}

type ServerConfiguration struct {
	Port int
	StaticPath string
}

type DatabaseConfiguration struct {
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

type OAuth2Configuration struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
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
	log.Print("Reading config file: ", file)
	viper.SetConfigName(file)

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// Set undefined variables
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "3306")
	viper.SetDefault("mail.host", "localhost")
	viper.SetDefault("mail.port", "25")
	viper.SetDefault("server.staticpath", "./static")

	err := viper.Unmarshal(configuration)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}

	os.Setenv("SESSION_KEY", hex.EncodeToString(securecookie.GenerateRandomKey(32)))
	log.Println("SESSION_KEY", os.Getenv("SESSION_KEY"))
}
