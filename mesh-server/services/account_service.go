package services

import (
    "context"

    "github.com/coinbase/rosetta-sdk-go/types"
)

// AccountAPIService implements the Account API interface
type AccountAPIService struct {
	network *types.NetworkIdentifier
}

// NewAccountAPIService creates a new AccountAPIService
func NewAccountAPIService(network *types.NetworkIdentifier) *AccountAPIService {
	return &AccountAPIService{
		network: network,
	}
}

// AccountBalance implements the /account/balance endpoint
func (s *AccountAPIService) AccountBalance(
	ctx context.Context,
	request *types.AccountBalanceRequest,
) (*types.AccountBalanceResponse, *types.Error) {
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