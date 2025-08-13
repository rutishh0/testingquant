package services

import (
    "context"

    "github.com/coinbase/rosetta-sdk-go/types"
)

// NetworkAPIService implements the Network API interface
type NetworkAPIService struct {
	network *types.NetworkIdentifier
}

// NewNetworkAPIService creates a new NetworkAPIService
func NewNetworkAPIService(network *types.NetworkIdentifier) *NetworkAPIService {
	return &NetworkAPIService{
		network: network,
	}
}

// NetworkList implements the /network/list endpoint
func (s *NetworkAPIService) NetworkList(
	ctx context.Context,
	request *types.MetadataRequest,
) (*types.NetworkListResponse, *types.Error) {
	return &types.NetworkListResponse{
		NetworkIdentifiers: []*types.NetworkIdentifier{
			s.network,
		},
	}, nil
}

// NetworkOptions implements the /network/options endpoint
func (s *NetworkAPIService) NetworkOptions(
	ctx context.Context,
	request *types.NetworkRequest,
) (*types.NetworkOptionsResponse, *types.Error) {
	return &types.NetworkOptionsResponse{
		Version: &types.Version{
			RosettaVersion: "1.5.1",
			NodeVersion:    "1.0.0",
		},
		Allow: &types.Allow{
			OperationStatuses: []*types.OperationStatus{
				{
					Status:     "SUCCESS",
					Successful: true,
				},
				{
					Status:     "FAILURE",
					Successful: false,
				},
			},
			OperationTypes: []string{
				"Transfer",
				"Reward",
				"Fee",
			},
			Errors: []*types.Error{
				{
					Code:      1,
					Message:   "Invalid request",
					Retriable: false,
				},
				{
					Code:      2,
					Message:   "Network error",
					Retriable: true,
				},
			},
			HistoricalBalanceLookup: true,
			CallMethods:             []string{},
			BalanceExemptions:       []*types.BalanceExemption{},
			MempoolCoins:            false,
		},
	}, nil
}

// NetworkStatus implements the /network/status endpoint
func (s *NetworkAPIService) NetworkStatus(
	ctx context.Context,
	request *types.NetworkRequest,
) (*types.NetworkStatusResponse, *types.Error) {
	// Mock current block
	currentBlock := &types.BlockIdentifier{
		Index: 1000000,
		Hash:  "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
	}

	// Mock genesis block
	genesisBlock := &types.BlockIdentifier{
		Index: 0,
		Hash:  "0x0000000000000000000000000000000000000000000000000000000000000000",
	}

	// Mock peers
	peers := []*types.Peer{
		{
			PeerID:   "peer1",
			Metadata: map[string]interface{}{"address": "127.0.0.1:8080"},
		},
		{
			PeerID:   "peer2",
			Metadata: map[string]interface{}{"address": "127.0.0.1:8081"},
		},
	}

    stage := "synced"
    synced := true
    return &types.NetworkStatusResponse{
		CurrentBlockIdentifier: currentBlock,
		CurrentBlockTimestamp:  1640995200000, // Mock timestamp
		GenesisBlockIdentifier: genesisBlock,
		OldestBlockIdentifier:  genesisBlock,
        SyncStatus: &types.SyncStatus{
            CurrentIndex: &currentBlock.Index,
            TargetIndex:  &currentBlock.Index,
            Stage:        &stage,
            Synced:       &synced,
        },
		Peers: peers,
	}, nil
} 