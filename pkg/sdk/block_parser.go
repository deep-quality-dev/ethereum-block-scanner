package sdk

import (
	"context"
	"fmt"
	"strings"

	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/blocks"
	"github.com/deep-quality-dev/ethereum-block-scanner/pkg/numbers"
)

// BlockParser implements SDK operations on the Ethereum blockchain.
type BlockParser struct {
	EthClient RPCClient
	TxStore   TransactionHistoryStore
	SubsStore SubscriptionsStore
}

var _ Parser = (*BlockParser)(nil)

// NewBlockParser is a constructor function for BlockParser.
func NewBlockParser(
	ethClient RPCClient,
	txStore TransactionHistoryStore,
	subsStore SubscriptionsStore,
) *BlockParser {
	return &BlockParser{
		EthClient: ethClient,
		TxStore:   txStore,
		SubsStore: subsStore,
	}
}

// GetCurrentBlock implements getting the latest parsed block of transactions.
func (p *BlockParser) GetCurrentBlock(ctx context.Context) (int, error) {
	resp, err := p.EthClient.Call(ctx, "eth_blockNumber")
	if err != nil {
		return -1, err
	}

	hexStr := fmt.Sprintf("%v", resp.Result)

	blockNumberInt, err := numbers.HexToInt(hexStr)
	if err != nil {
		return -1, err
	}

	return blockNumberInt, nil
}

// Subscribe implements adding an address to an observer.
func (p *BlockParser) Subscribe(address string) bool {
	p.SubsStore.InsertSubscriberAddress(strings.ToLower(address))

	return true
}

// GetBlockTransactions returns all transactions contained is a block.
func (p *BlockParser) GetBlockTransactions(ctx context.Context, blockNum int) ([]blocks.Transaction, error) {
	var block blocks.Block

	err := p.EthClient.CallFor(
		ctx, &block, "eth_getBlockByNumber", numbers.IntToHex(blockNum), true)
	if err != nil {
		return nil, err
	}

	return block.Transactions, nil
}

// GetTransactionsForBlockRange implements getting the transaction history for inbound and outbound transactions
// given an address.
func (p *BlockParser) GetTransactionsForBlockRange(
	ctx context.Context, address string, blockRange int) ([]blocks.Transaction, error) {
	latestBlockNum, err := p.GetCurrentBlock(ctx)
	if err != nil {
		return nil, err
	}

	currentBlockRange := latestBlockNum - blockRange

	blockTransactions, err := p.GetBlockTransactions(ctx, latestBlockNum)
	if err != nil {
		return nil, err
	}

	address = strings.ToLower(address)
	lastStoredBlockNum := p.TxStore.GetLatestBlockNumberPerAddress(address)

	if lastStoredBlockNum > 0 {
		currentBlockRange = lastStoredBlockNum
	}

	for i := latestBlockNum; i >= currentBlockRange; i-- {
		for _, tx := range blockTransactions {
			//log.Println(tx.To)
			//log.Println(tx.From)
			if address == strings.ToLower(tx.To) {
				p.TxStore.Insert(address, latestBlockNum, tx, true)
			} else if address == strings.ToLower(tx.From) {
				p.TxStore.Insert(address, latestBlockNum, tx, false)
			}
		}
	}

	return p.TxStore.GetAllTransactionsPerAddress(address), nil
}

// GetTransactionsPerSubscriber implements listing of observed transactions given a registered subscriber address.
func (p *BlockParser) GetTransactionsPerSubscriber(address string) []blocks.Transaction {
	address = strings.ToLower(address)

	return p.SubsStore.GetObservedTransactionsPerAddress(address)
}
