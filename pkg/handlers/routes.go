package handlers

import (
	"fmt"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/deep-quality-dev/ethereum-block-scanner/docs"
	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/configs"
)

func registerHTTPRoutes(
	config *configs.Config, muxer *mux.Router, handler *BlockHandler) *mux.Router {
	muxer.HandleFunc(
		"/api/v1/block/current",
		handler.GetCurrentBlock()).Methods("GET")
	muxer.HandleFunc(
		"/api/v1/address/{address}/transactions",
		handler.GetBlockTransactionsPerAddress()).Methods("GET")
	muxer.HandleFunc(
		"/api/v1/subscription/{address}/transactions",
		handler.GetTransactionsPerSubscriber()).Methods("GET")
	muxer.HandleFunc(
		"/api/v1/address/subscribe",
		handler.SubscribeAddress()).Methods("POST")

	swaggerJsonURL := fmt.Sprintf("http://%s:%d/swagger/doc.json", config.Host, config.Port)

	muxer.PathPrefix("/swagger/").
		Handler(httpSwagger.Handler(httpSwagger.URL(swaggerJsonURL))) // The url pointing to API definition

	return muxer
}
