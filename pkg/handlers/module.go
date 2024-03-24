package handlers

import (
	"github.com/gorilla/mux"

	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/configs"
	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/sdk"
	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/storage/memory"
	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/transport/client/jsonrpc"
)

// InitializeHandlers registers HTTP routes and wires dependencies for HTTP handlers.
func InitializeHandlers(
	config *configs.Config, router *mux.Router, client *jsonrpc.RPCClient) *mux.Router {
	txStore := memory.NewTransactionsRepository()
	parser := sdk.NewBlockParser(client, txStore)
	handler := NewBlockHandler(parser)

	registerHTTPRoutes(config, router, handler)

	return router
}
