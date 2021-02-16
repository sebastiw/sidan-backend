package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/viper"

	c "github.com/sebastiw/sidan-backend/src/config"
	r "github.com/sebastiw/sidan-backend/src/router"
)

func read_config(configuration *c.Configurations) {
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
	viper.SetDefault("database.dbname", "test_db")

	err := viper.Unmarshal(configuration)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}
}

func main() {
	var configuration c.Configurations

	read_config(&configuration)

	// sql.connect(connect_config)

	address := fmt.Sprintf(":%v", configuration.Server.Port)
	log.Printf("Starting backend service at %v", address)

	log.Fatal(http.ListenAndServe(address, r.Mux()))
}
