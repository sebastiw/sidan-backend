package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/viper"

	c "github.com/sebastiw/sidan-backend/src/config"
	r "github.com/sebastiw/sidan-backend/src/router"
)

func main() {
	var configuration c.Configurations

	c.ReadConfig(&configuration)

	// sql.connect(connect_config)

	address := fmt.Sprintf(":%v", configuration.Server.Port)
	log.Printf("Starting backend service at %v", address)

	log.Fatal(http.ListenAndServe(address, r.Mux()))
}
