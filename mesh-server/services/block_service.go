package services

import (
    "context"

    "github.com/coinbase/rosetta-sdk-go/types"
)

// BlockAPIService implements the Block API interface
type BlockAPIService struct {
	network *types.NetworkIdentifier
}

// NewBlockAPIService creates a new BlockAPIService
func NewBlockAPIService(network *types.NetworkIdentifier) *BlockAPIService {
	return &BlockAPIService{
		network: network,
	}
}

// helper to get *string
func strPtr(s string) *string { return &s }

// Block implements the /block endpoint
func (s *BlockAPIService) Block(
	ctx context.Context,
	request *types.BlockRequest,
) (*types.BlockResponse, *types.Error) {
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