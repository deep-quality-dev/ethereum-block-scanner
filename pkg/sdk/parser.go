package sdk

import (
	"context"
	"fmt"
	"strings"

	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/blocks"
	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/numbers"
	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/storage/memory"
	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/transport/client/jsonrpc"
)

const _maxBlockRange = 100000

// Parser defines SDK operations on the Ethereum blockchain.
type Parser interface {
	// GetCurrentBlock last parsed block.
	GetCurrentBlock(ctx context.Context) (int, error)

	// Subscribe add address to observer
	Subscribe(address string) bool

	// GetTransactions list of inbound or outbound transactions for an address
	GetTransactions(ctx context.Context, address string) ([]blocks.Transaction, error)
}

// BlockParser implements SDK operations on the Ethereum blockchain.
type BlockParser struct {
	EthClient *jsonrpc.RPCClient
	TxStore   *memory.TransactionsRepository
}

var _ Parser = (*BlockParser)(nil)

// NewBlockParser is a constructor function for BlockParser.
func NewBlockParser(ethClient *jsonrpc.RPCClient, txStore *memory.TransactionsRepository) *BlockParser {
	return &BlockParser{
		EthClient: ethClient,
		TxStore:   txStore,
	}
}

// GetCurrentBlock implements getting the latest parsed block of transactions.
func (p *BlockParser) GetCurrentBlock(ctx context.Context) (int, error) {
	resp, err := p.EthClient.Call(ctx, "eth_blockNumber")
	if err != nil {
		return -1, err
	}

	hexStr := fmt.Sprintf("%v", resp.Result)

	blockNumberInt, err := numbers.ParseInt(hexStr)
	if err != nil {
		return -1, err
	}

	return blockNumberInt, nil
}

// Subscribe implements adding an address to an observer.
func (p *BlockParser) Subscribe(address string) bool {
	return false
}

// GetTransactions implements getting the transaction history for inbound and outbound transactions given an address.
// NOTE: Checked blocks are limited to _maxBlockRange from the latest block due to time constraints.
func (p *BlockParser) GetTransactions(ctx context.Context, address string) ([]blocks.Transaction, error) {
	latestBlockNum, err := p.GetCurrentBlock(ctx)
	if err != nil {
		return nil, err
	}

	currentBlockRange := latestBlockNum - _maxBlockRange

	var block blocks.Block

	err = p.EthClient.CallFor(
		ctx, &block, "eth_getBlockByNumber", numbers.IntToHex(latestBlockNum), true)
	if err != nil {
		return nil, err
	}

	address = strings.ToLower(address)
	lastStoredBlockNum := p.TxStore.GetLatestBlockNumberPerAddress(address)

	if lastStoredBlockNum > 0 {
		currentBlockRange = lastStoredBlockNum
	}

	for i := latestBlockNum; i >= currentBlockRange; i-- {
		for _, tx := range block.Transactions {
			if address == strings.ToLower(tx.To) {
				p.TxStore.Insert(address, latestBlockNum, tx, true)
			} else if address == strings.ToLower(tx.From) {
				p.TxStore.Insert(address, latestBlockNum, tx, false)
			}
		}
	}

	return p.TxStore.GetAllTransactionsPerAddress(address), nil
}