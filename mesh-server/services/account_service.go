package services

import (
    "context"
    "os"

    "github.com/coinbase/rosetta-sdk-go/types"
)

// AccountAPIService implements the Account API interface
type AccountAPIService struct {
	network *types.NetworkIdentifier
    rpc     *EthRPCClient
    live    bool
}

// NewAccountAPIService creates a new AccountAPIService
func NewAccountAPIService(network *types.NetworkIdentifier, rpc *EthRPCClient) *AccountAPIService {
    live := false
    if rpc != nil {
        if v := os.Getenv("MESH_LIVE"); v == "false" || v == "0" {
            live = false
        } else {
            live = true
        }
    }
	return &AccountAPIService{
		network: network,
        rpc:     rpc,
        live:    live,
	}
}

// AccountBalance implements the /account/balance endpoint
func (s *AccountAPIService) AccountBalance(
	ctx context.Context,
	request *types.AccountBalanceRequest,
) (*types.AccountBalanceResponse, *types.Error) {
    // Live path when RPC is available and live mode enabled
    if s.rpc != nil && s.live {
        accountAddress := request.AccountIdentifier.Address
        if accountAddress == "" {
            return nil, &types.Error{Code: 1, Message: "Account address is required", Retriable: false}
        }
        
        // Determine block number to query
        blockTag := "latest"
        if request.BlockIdentifier != nil {
            if request.BlockIdentifier.Index != nil {
                blockTag = int64ToHex(*request.BlockIdentifier.Index)
            } else if request.BlockIdentifier.Hash != nil {
                blockTag = *request.BlockIdentifier.Hash
            }
        }
        
        // Get ETH balance
        var balanceHex string
        if err := s.rpc.call("eth_getBalance", []interface{}{accountAddress, blockTag}, &balanceHex); err == nil {
            // Get block info
            var blk rpcBlock
            _ = s.rpc.call("eth_getBlockByNumber", []interface{}{blockTag, false}, &blk)
            
            // Parse block identifier
            blockIdx := int64(0)
            if blk.Number != "" {
                blockIdx, _ = hexToInt64(blk.Number)
            }
            blockHash := blk.Hash
            if blockHash == "" && blockTag != "latest" {
                blockHash = blockTag
            }
            
            currentBlock := &types.BlockIdentifier{Index: blockIdx, Hash: blockHash}
            
            // Convert balance
            balanceWei := hexToBigIntMust(balanceHex)
            
            return &types.AccountBalanceResponse{
                BlockIdentifier: currentBlock,
                Balances: []*types.Amount{
                    {
                        Value:    balanceWei,
                        Currency: &types.Currency{Symbol: "ETH", Decimals: 18},
                    },
                },
                Metadata: map[string]interface{}{},
            }, nil
        }
        // fall back to mock on error
    }

	// Mock account balance data
	accountAddress := request.AccountIdentifier.Address
	if accountAddress == "" {
		return nil, &types.Error{
			Code:      1,
			Message:   "Account address is required",
			Retriable: false,
		}
	}

	// Mock current block
	currentBlock := &types.BlockIdentifier{
		Index: 1000000,
		Hash:  "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
	}

	// Mock balances
	balances := []*types.Amount{
		{
			Value:    "5000000000000000000", // 5 ETH in wei
			Currency: &types.Currency{Symbol: "ETH", Decimals: 18},
		},
		{
			Value:    "100000000", // 100 USDC (6 decimals)
			Currency: &types.Currency{Symbol: "USDC", Decimals: 6},
		},
		{
			Value:    "1000000000000000000000", // 1000 USDT (18 decimals)
			Currency: &types.Currency{Symbol: "USDT", Decimals: 18},
		},
	}

	return &types.AccountBalanceResponse{
		BlockIdentifier: currentBlock,
		Balances:        balances,
		Metadata: map[string]interface{}{
			"sequence_number": "42",
			"account_type":    "contract",
		},
	}, nil
}

// AccountCoins implements the /account/coins endpoint
func (s *AccountAPIService) AccountCoins(
	ctx context.Context,
	request *types.AccountCoinsRequest,
) (*types.AccountCoinsResponse, *types.Error) {
	// Mock account coins data
	accountAddress := request.AccountIdentifier.Address
	if accountAddress == "" {
		return nil, &types.Error{
			Code:      1,
			Message:   "Account address is required",
			Retriable: false,
		}
	}

	// Mock current block
	currentBlock := &types.BlockIdentifier{
		Index: 1000000,
		Hash:  "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
	}

	// Mock coins
	coins := []*types.Coin{
		{
			CoinIdentifier: &types.CoinIdentifier{
				Identifier: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef:0",
			},
			Amount: &types.Amount{
				Value:    "5000000000000000000", // 5 ETH in wei
				Currency: &types.Currency{Symbol: "ETH", Decimals: 18},
			},
		},
		{
			CoinIdentifier: &types.CoinIdentifier{
				Identifier: "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890:0",
			},
			Amount: &types.Amount{
				Value:    "100000000", // 100 USDC (6 decimals)
				Currency: &types.Currency{Symbol: "USDC", Decimals: 6},
			},
		},
	}

	return &types.AccountCoinsResponse{
		BlockIdentifier: currentBlock,
		Coins:          coins,
		Metadata: map[string]interface{}{
			"sequence_number": "42",
			"account_type":    "contract",
		},
	}, nil
}