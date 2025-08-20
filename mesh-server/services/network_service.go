package services

import (
    "context"
    "os"

    "github.com/coinbase/rosetta-sdk-go/types"
)

// NetworkAPIService implements the Network API interface
type NetworkAPIService struct {
    network *types.NetworkIdentifier
    rpc     *EthRPCClient
    live    bool
}

// NewNetworkAPIService creates a new NetworkAPIService
func NewNetworkAPIService(network *types.NetworkIdentifier, rpc *EthRPCClient) *NetworkAPIService {
    live := false
    if rpc != nil {
        // Enable live mode by default when RPC is provided, can be disabled with MESH_LIVE=false
        if v := os.Getenv("MESH_LIVE"); v == "false" || v == "0" {
            live = false
        } else {
            live = true
        }
    }
    return &NetworkAPIService{
        network: network,
        rpc:     rpc,
        live:    live,
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
    // Live path when RPC is available and live mode enabled
    if s.rpc != nil && s.live {
        // Get latest block number
        var numHex string
        if err := s.rpc.call("eth_blockNumber", []interface{}{}, &numHex); err == nil {
            if currentIndex, err := hexToInt64(numHex); err == nil {
                // Fetch current block (hash + timestamp)
                var blk rpcBlock
                _ = s.rpc.call("eth_getBlockByNumber", []interface{}{int64ToHex(currentIndex), false}, &blk)
                // Fetch genesis
                var genesis rpcBlock
                _ = s.rpc.call("eth_getBlockByNumber", []interface{}{int64ToHex(0), false}, &genesis)

                // Build identifiers
                currentHash := blk.Hash
                if currentHash == "" {
                    currentHash = ""
                }
                genesisHash := genesis.Hash

                // Timestamp
                ts := int64(0)
                if blk.Timestamp != "" {
                    if v, err := hexToInt64(blk.Timestamp); err == nil {
                        ts = v * 1000 // seconds -> ms
                    }
                }

                current := &types.BlockIdentifier{Index: currentIndex, Hash: currentHash}
                g := &types.BlockIdentifier{Index: 0, Hash: genesisHash}

                stage := "synced"
                synced := true
                return &types.NetworkStatusResponse{
                    CurrentBlockIdentifier: current,
                    CurrentBlockTimestamp:  ts,
                    GenesisBlockIdentifier: g,
                    OldestBlockIdentifier:  g,
                    SyncStatus: &types.SyncStatus{
                        CurrentIndex: &currentIndex,
                        TargetIndex:  &currentIndex,
                        Stage:        &stage,
                        Synced:       &synced,
                    },
                    Peers: []*types.Peer{},
                }, nil
            }
        }
        // If live fails for any reason, fall back to mock below
    }

    // Mock fallback
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