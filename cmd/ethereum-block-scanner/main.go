package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"

	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/configs"
	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/handlers"
	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/sdk"
	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/storage/memory"
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

	txStore := memory.NewTransactionsRepository()
	subsStore := memory.NewSubscriptionsRepository()

	blockParser := sdk.NewBlockParser(client, txStore, subsStore)
	blockListener := sdk.NewBlockObserver(blockParser, subsStore)

	errServerCh := make(chan error)
	errListenerCh := make(chan error)

	// Start block observer to track new transactions that have occurred in the latest block
	// involving subscribed addresses.
	go blockListener.ListenForNewTransactions(ctx, errListenerCh)

	router := mux.NewRouter()
	router = handlers.InitializeHandlers(conf, router, blockParser)
	s := server.NewServer(conf, router)

	// Start HTTP server.
	go s.Start(ctx, errServerCh)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sigs:
			if err = s.Stop(ctx); err != nil {
				log.Fatal("error stopping HTTP server:", errors.WithStack(err))
			}

			os.Exit(0)
		case err = <-errServerCh:
			log.Fatal("HTTP server error: ", errors.WithStack(err))
		case err = <-errListenerCh:
			log.Println(err)
		}
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
