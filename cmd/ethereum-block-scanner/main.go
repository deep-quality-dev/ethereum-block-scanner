package main

import (
	"context"
	"log"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/configs"
	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/handlers"
	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/transport/client/jsonrpc"
	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/transport/server"
)

// @title Ethereum Block Scanner API
// @version 1.0
// @description API for exploring Ethereum blocks.
// @termsOfService http://swagger.io/terms/

// @host 0.0.0.0:8080
// @BasePath /
func main() {
	ctx := context.Background()

	var err error

	setEnvironment()

	conf := configs.InitializeConfig()
	client := jsonrpc.NewDefaultClient(conf.EthereumHost)

	router := mux.NewRouter()
	router = handlers.InitializeHandlers(conf, router, client)
	s := server.NewServer(conf, router)

	if err = s.Run(ctx); err != nil {
		log.Fatal("error starting the HTTP server: ", err)
	}
}

func setEnvironment() {
	_, foundHost := os.LookupEnv("SERVER_HOST")
	_, foundPort := os.LookupEnv("SERVER_PORT")

	if !foundHost && !foundPort {
		err := godotenv.Load(".env.dist")
		if err != nil {
			panic(err)
		}
	}
}
