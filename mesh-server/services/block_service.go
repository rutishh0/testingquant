package services

import (
    "context"
    "encoding/json"
    "os"

    "github.com/coinbase/rosetta-sdk-go/types"
)

// BlockAPIService implements the Block API interface
type BlockAPIService struct {
	network *types.NetworkIdentifier
    rpc     *EthRPCClient
    live    bool
}

// NewBlockAPIService creates a new BlockAPIService
func NewBlockAPIService(network *types.NetworkIdentifier, rpc *EthRPCClient) *BlockAPIService {
    live := false
    if rpc != nil {
        if v := os.Getenv("MESH_LIVE"); v == "false" || v == "0" {
            live = false
        } else {
            live = true
        }
    }
	return &BlockAPIService{
		network: network,
        rpc:     rpc,
        live:    live,
	}
}

// helper to get *string
func strPtr(s string) *string { return &s }

// Block implements the /block endpoint
func (s *BlockAPIService) Block(
	ctx context.Context,
	request *types.BlockRequest,
) (*types.BlockResponse, *types.Error) {
    // Live path
    if s.rpc != nil && s.live {
        var blk rpcBlock
        // Resolve by index or hash
        if request.BlockIdentifier != nil {
            if request.BlockIdentifier.Index != nil {
                if err := s.rpc.call("eth_getBlockByNumber", []interface{}{int64ToHex(*request.BlockIdentifier.Index), true}, &blk); err != nil {
                    // fall through to mock
                } else {
                    return s.blockToRosetta(&blk)
                }
            } else if request.BlockIdentifier.Hash != nil {
                if err := s.rpc.call("eth_getBlockByHash", []interface{}{*request.BlockIdentifier.Hash, true}, &blk); err != nil {
                    // fall through to mock
                } else {
                    return s.blockToRosetta(&blk)
                }
            }
        }
        // If no identifier provided, fetch latest
        if err := s.rpc.call("eth_getBlockByNumber", []interface{}{"latest", true}, &blk); err == nil {
            return s.blockToRosetta(&blk)
        }
        // fall back to mock below on any error
    }

	// Mock block data
	blockIndex := int64(1000000)
	if request.BlockIdentifier.Index != nil {
		blockIndex = *request.BlockIdentifier.Index
	}

	blockHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	if request.BlockIdentifier.Hash != nil {
		blockHash = *request.BlockIdentifier.Hash
	}

	block := &types.Block{
		BlockIdentifier: &types.BlockIdentifier{
			Index: blockIndex,
			Hash:  blockHash,
		},
		ParentBlockIdentifier: &types.BlockIdentifier{
			Index: blockIndex - 1,
			Hash:  "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		},
		Timestamp: 1640995200000, // Mock timestamp
		Transactions: []*types.Transaction{
			{
				TransactionIdentifier: &types.TransactionIdentifier{
					Hash: "0xtx1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				},
				Operations: []*types.Operation{
					{
						OperationIdentifier: &types.OperationIdentifier{
							Index: 0,
						},
                        Type:   "Transfer",
                        Status: strPtr("SUCCESS"),
						Account: &types.AccountIdentifier{
							Address: "0x1234567890abcdef1234567890abcdef1234567890",
						},
						Amount: &types.Amount{
							Value:    "1000000000000000000", // 1 ETH in wei
							Currency: &types.Currency{Symbol: "ETH", Decimals: 18},
						},
					},
					{
						OperationIdentifier: &types.OperationIdentifier{
							Index: 1,
						},
                        Type:   "Transfer",
                        Status: strPtr("SUCCESS"),
						Account: &types.AccountIdentifier{
							Address: "0xabcdef1234567890abcdef1234567890abcdef1234",
						},
						Amount: &types.Amount{
							Value:    "-1000000000000000000", // -1 ETH in wei
							Currency: &types.Currency{Symbol: "ETH", Decimals: 18},
						},
					},
				},
			},
		},
		Metadata: map[string]interface{}{
			"gas_used":     "21000",
			"gas_limit":    "21000",
			"base_fee_per_gas": "20000000000",
		},
	}

	return &types.BlockResponse{
		Block: block,
		OtherTransactions: []*types.TransactionIdentifier{
			{
				Hash: "0xtxabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			},
		},
	}, nil
}

// BlockTransaction implements the /block/transaction endpoint
func (s *BlockAPIService) BlockTransaction(
	ctx context.Context,
	request *types.BlockTransactionRequest,
) (*types.BlockTransactionResponse, *types.Error) {
    if s.rpc != nil && s.live {
        txHash := request.TransactionIdentifier.Hash
        if txHash == "" {
            return nil, &types.Error{Code: 1, Message: "Transaction hash is required", Retriable: false}
        }
        var tx rpcTx
        if err := s.rpc.call("eth_getTransactionByHash", []interface{}{txHash}, &tx); err == nil && tx.Hash != "" {
            // fetch receipt for fee data
            var rc rpcReceipt
            _ = s.rpc.call("eth_getTransactionReceipt", []interface{}{txHash}, &rc)
            // Build rosetta transaction
            ops := []*types.Operation{}
            // value transfer if present
            if tx.Value != "" {
                // from (debit)
                ops = append(ops, &types.Operation{
                    OperationIdentifier: &types.OperationIdentifier{Index: 0},
                    Type:   "Transfer",
                    Status: strPtr("SUCCESS"),
                    Account: &types.AccountIdentifier{Address: tx.From},
                    Amount: &types.Amount{Value: "-" + hexToBigIntMust(tx.Value), Currency: &types.Currency{Symbol: "ETH", Decimals: 18}},
                })
                // to (credit)
                ops = append(ops, &types.Operation{
                    OperationIdentifier: &types.OperationIdentifier{Index: 1},
                    Type:   "Transfer",
                    Status: strPtr("SUCCESS"),
                    Account: &types.AccountIdentifier{Address: tx.To},
                    Amount: &types.Amount{Value: hexToBigIntMust(tx.Value), Currency: &types.Currency{Symbol: "ETH", Decimals: 18}},
                })
            }
            // fee if available
            if rc.GasUsed != "" && rc.EffectiveGasPrice != "" {
                fee := mulHexBigToString(rc.GasUsed, rc.EffectiveGasPrice)
                ops = append(ops, &types.Operation{
                    OperationIdentifier: &types.OperationIdentifier{Index: int64(len(ops))},
                    Type:   "Fee",
                    Status: strPtr("SUCCESS"),
                    Account: &types.AccountIdentifier{Address: tx.From},
                    Amount: &types.Amount{Value: "-" + fee, Currency: &types.Currency{Symbol: "ETH", Decimals: 18}},
                })
            }
            return &types.BlockTransactionResponse{
                Transaction: &types.Transaction{
                    TransactionIdentifier: &types.TransactionIdentifier{Hash: tx.Hash},
                    Operations:           ops,
                },
            }, nil
        }
        // fall back to mock on error
    }

	// Mock transaction data
	txHash := request.TransactionIdentifier.Hash
	if txHash == "" {
		return nil, &types.Error{
			Code:      1,
			Message:   "Transaction hash is required",
			Retriable: false,
		}
	}

	transaction := &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: txHash,
		},
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{
					Index: 0,
				},
                Type:   "Transfer",
                Status: strPtr("SUCCESS"),
				Account: &types.AccountIdentifier{
					Address: "0x1234567890abcdef1234567890abcdef1234567890",
				},
				Amount: &types.Amount{
					Value:    "1000000000000000000", // 1 ETH in wei
					Currency: &types.Currency{Symbol: "ETH", Decimals: 18},
				},
			},
			{
				OperationIdentifier: &types.OperationIdentifier{
					Index: 1,
				},
                Type:   "Transfer",
                Status: strPtr("SUCCESS"),
				Account: &types.AccountIdentifier{
					Address: "0xabcdef1234567890abcdef1234567890abcdef1234",
				},
				Amount: &types.Amount{
					Value:    "-1000000000000000000", // -1 ETH in wei
					Currency: &types.Currency{Symbol: "ETH", Decimals: 18},
				},
			},
			{
				OperationIdentifier: &types.OperationIdentifier{
					Index: 2,
				},
                Type:   "Fee",
                Status: strPtr("SUCCESS"),
				Account: &types.AccountIdentifier{
					Address: "0x1234567890abcdef1234567890abcdef1234567890",
				},
				Amount: &types.Amount{
					Value:    "-21000000000000000", // 0.021 ETH fee
					Currency: &types.Currency{Symbol: "ETH", Decimals: 18},
				},
			},
		},
		Metadata: map[string]interface{}{
			"gas_used":     "21000",
			"gas_price":    "20000000000",
			"gas_limit":    "21000",
			"nonce":        "42",
		},
	}

	return &types.BlockTransactionResponse{
		Transaction: transaction,
	}, nil
}

// blockToRosetta converts rpcBlock to Rosetta BlockResponse
func (s *BlockAPIService) blockToRosetta(blk *rpcBlock) (*types.BlockResponse, *types.Error) {
    // parse index
    idx, _ := hexToInt64(blk.Number)
    // timestamp ms
    ts := int64(0)
    if blk.Timestamp != "" {
        if v, err := hexToInt64(blk.Timestamp); err == nil { ts = v * 1000 }
    }
    // parse transactions
    txs := []*types.Transaction{}
    if len(blk.Transactions) > 0 {
        var full []rpcTx
        if err := json.Unmarshal(blk.Transactions, &full); err == nil {
            for i := range full {
                tx := full[i]
                ops := []*types.Operation{}
                if tx.Value != "" {
                    ops = append(ops, &types.Operation{
                        OperationIdentifier: &types.OperationIdentifier{Index: 0},
                        Type:   "Transfer",
                        Status: strPtr("SUCCESS"),
                        Account: &types.AccountIdentifier{Address: tx.From},
                        Amount: &types.Amount{Value: "-" + hexToBigIntMust(tx.Value), Currency: &types.Currency{Symbol: "ETH", Decimals: 18}},
                    })
                    ops = append(ops, &types.Operation{
                        OperationIdentifier: &types.OperationIdentifier{Index: 1},
                        Type:   "Transfer",
                        Status: strPtr("SUCCESS"),
                        Account: &types.AccountIdentifier{Address: tx.To},
                        Amount: &types.Amount{Value: hexToBigIntMust(tx.Value), Currency: &types.Currency{Symbol: "ETH", Decimals: 18}},
                    })
                }
                txs = append(txs, &types.Transaction{
                    TransactionIdentifier: &types.TransactionIdentifier{Hash: tx.Hash},
                    Operations: ops,
                })
            }
        }
    }

    block := &types.Block{
        BlockIdentifier: &types.BlockIdentifier{Index: idx, Hash: blk.Hash},
        ParentBlockIdentifier: &types.BlockIdentifier{Index: idx - 1, Hash: blk.ParentHash},
        Timestamp: ts,
        Transactions: txs,
    }
    return &types.BlockResponse{Block: block}, nil
}

// helpers for math
func hexToBigIntMust(h string) string {
    bi, err := hexToBigInt(h)
    if err != nil { return "0" }
    return bi.String()
}

func mulHexBigToString(aHex, bHex string) string {
    a, err1 := hexToBigInt(aHex)
    b, err2 := hexToBigInt(bHex)
    if err1 != nil || err2 != nil { return "0" }
    return bigMul(a, b).String()
}