package connector

import (
	"errors"
	"fmt"
	"github.com/rutishh0/testingquant/internal/mesh"
	"github.com/rutishh0/testingquant/internal/overledger"
)

// Service defines the connector service interface
type Service interface {
	Preprocess(req *PreprocessRequest) (*PreprocessResponse, error)
	Payloads(req *PayloadsRequest) (*PayloadsResponse, error)
	Combine(req *CombineRequest) (*CombineResponse, error)
	Submit(req *SubmitRequest) (*SubmitResponse, error)
	GetBalance(req *BalanceRequest) (*BalanceResponse, error)
	GetBlock(req *BlockRequest) (*BlockResponse, error)
	GetTransaction(req *TransactionRequest) (*TransactionResponse, error)
	// Overledger-specific methods
	GetOverledgerNetworks() (*overledger.NetworksResponse, error)
	GetOverledgerBalance(networkID, address string) (*overledger.BalanceResponse, error)
	CreateOverledgerTransaction(req *overledger.TransactionRequest) (*overledger.TransactionResponse, error)
	TestOverledgerConnection() error
}

// service implements the Service interface
type service struct {
	meshClient       *mesh.Client
	overledgerClient *overledger.Client
}

// NewService creates a new connector service
func NewService(meshClient *mesh.Client, overledgerClient *overledger.Client) Service {
	return &service{
		meshClient:       meshClient,
		overledgerClient: overledgerClient,
	}
}

// Preprocess handles the preprocess request
func (s *service) Preprocess(req *PreprocessRequest) (*PreprocessResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Map Overledger request to Mesh request
	meshReq := &mesh.ConstructionPreprocessRequest{
		NetworkIdentifier: mesh.NetworkIdentifier{
			Blockchain: req.DLT,
			Network:    req.Network,
		},
		Operations: mapOperations(req),
		Metadata:   req.Metadata,
	}

	// Call Mesh API
	meshResp, err := s.meshClient.ConstructionPreprocess(meshReq)
	if err != nil {
		return nil, fmt.Errorf("mesh preprocess failed: %w", err)
	}

	// Map Mesh response to Overledger response
	return &PreprocessResponse{
		Options:            meshResp.Options,
		RequiredSigners:    mapRequiredSigners(meshResp.RequiredPublicKeys),
		TransactionFee:     calculateTransactionFee(meshResp),
		GatewayFee:         "0",
		PreparedTransaction: generatePreparedTransaction(meshResp),
	}, nil
}

// Payloads handles the payloads request
func (s *service) Payloads(req *PayloadsRequest) (*PayloadsResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Map Overledger request to Mesh request
	meshReq := &mesh.ConstructionPayloadsRequest{
		NetworkIdentifier: mesh.NetworkIdentifier{
			Blockchain: req.DLT,
			Network:    req.Network,
		},
		Operations: mapPayloadOperations(req),
		Metadata:   req.Metadata,
		PublicKeys: mapPublicKeys(req.PublicKeys),
	}

	// Call Mesh API
	meshResp, err := s.meshClient.ConstructionPayloads(meshReq)
	if err != nil {
		return nil, fmt.Errorf("mesh payloads failed: %w", err)
	}

	// Map Mesh response to Overledger response
	return &PayloadsResponse{
		UnsignedTransaction: meshResp.UnsignedTransaction,
		Payloads:            mapSigningPayloads(meshResp.Payloads),
	}, nil
}

// Combine handles the combine request
func (s *service) Combine(req *CombineRequest) (*CombineResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Map Overledger request to Mesh request
	meshReq := &mesh.ConstructionCombineRequest{
		NetworkIdentifier: mesh.NetworkIdentifier{
			Blockchain: req.DLT,
			Network:    req.Network,
		},
		UnsignedTransaction: req.UnsignedTransaction,
		Signatures:          mapSignatures(req.Signatures),
	}

	// Call Mesh API
	meshResp, err := s.meshClient.ConstructionCombine(meshReq)
	if err != nil {
		return nil, fmt.Errorf("mesh combine failed: %w", err)
	}

	// Map Mesh response to Overledger response
	return &CombineResponse{
		SignedTransaction: meshResp.SignedTransaction,
	}, nil
}

// Submit handles the submit request
func (s *service) Submit(req *SubmitRequest) (*SubmitResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Map Overledger request to Mesh request
	meshReq := &mesh.ConstructionSubmitRequest{
		NetworkIdentifier: mesh.NetworkIdentifier{
			Blockchain: req.DLT,
			Network:    req.Network,
		},
		SignedTransaction: req.SignedTransaction,
	}

	// Call Mesh API
	meshResp, err := s.meshClient.ConstructionSubmit(meshReq)
	if err != nil {
		return nil, fmt.Errorf("mesh submit failed: %w", err)
	}

	// Map Mesh response to Overledger response
	return &SubmitResponse{
		TransactionID: meshResp.TransactionIdentifier.Hash,
		Status:        "PENDING",
		Metadata:      meshResp.Metadata,
	}, nil
}

// GetBalance handles the balance request
func (s *service) GetBalance(req *BalanceRequest) (*BalanceResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Map Overledger request to Mesh request
	meshReq := &mesh.AccountBalanceRequest{
		NetworkIdentifier: mesh.NetworkIdentifier{
			Blockchain: req.DLT,
			Network:    req.Network,
		},
		AccountIdentifier: mesh.AccountIdentifier{
			Address:  req.Address,
			Metadata: req.Metadata,
		},
	}

	// Call Mesh API
	meshResp, err := s.meshClient.AccountBalance(meshReq)
	if err != nil {
		return nil, fmt.Errorf("mesh balance failed: %w", err)
	}

	// Map Mesh response to Overledger response
	return &BalanceResponse{
		Address:  req.Address,
		Balances: mapBalances(meshResp.Balances),
		Block: BlockInfo{
			Number: meshResp.BlockIdentifier.Index,
			Hash:   meshResp.BlockIdentifier.Hash,
		},
		Metadata: meshResp.Metadata,
	}, nil
}

// GetBlock handles the block request
func (s *service) GetBlock(req *BlockRequest) (*BlockResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// Map Overledger request to Mesh request
	meshReq := &mesh.BlockRequest{
		NetworkIdentifier: mesh.NetworkIdentifier{
			Blockchain: req.DLT,
			Network:    req.Network,
		},
		BlockIdentifier: mesh.PartialBlockIdentifier{},
	}

	// Set block identifier based on request
	if req.BlockNumber != nil {
		index := int64(*req.BlockNumber)
		meshReq.BlockIdentifier.Index = &index
	}

	if req.BlockHash != "" {
		meshReq.BlockIdentifier.Hash = &req.BlockHash
	}

	// Call Mesh API
	meshResp, err := s.meshClient.Block(meshReq)
	if err != nil {
		return nil, fmt.Errorf("mesh block failed: %w", err)
	}

	if meshResp.Block == nil {
		return nil, errors.New("block not found")
	}

	// Map Mesh response to Overledger response
	return &BlockResponse{
		BlockID: meshResp.Block.BlockIdentifier.Hash,
		Number:  meshResp.Block.BlockIdentifier.Index,
		Transactions: mapTransactions(meshResp.Block.Transactions),
		Timestamp:    meshResp.Block.Timestamp,
		ParentHash:   meshResp.Block.ParentBlockIdentifier.Hash,
		Metadata:     meshResp.Block.Metadata,
	}, nil
}

// GetTransaction handles the transaction request
func (s *service) GetTransaction(req *TransactionRequest) (*TransactionResponse, error) {
	// This is a bit more complex as Mesh requires both block ID and transaction ID
	// We'll need to search for the transaction in recent blocks
	// For simplicity in this PoC, we'll return an error suggesting to use the block endpoint
	return nil, errors.New("direct transaction lookup not supported in this PoC; use block endpoint instead")
}

// Helper functions for mapping between Overledger and Mesh models

func mapOperations(req *PreprocessRequest) []mesh.Operation {
	// Simplified implementation for PoC
	// In a real implementation, this would map Overledger operations to Mesh operations
	operations := make([]mesh.Operation, 0)
	
	// For a simple transfer
	if req.Type == "TRANSFER" && len(req.Transfers) > 0 {
		transfer := req.Transfers[0]
		
		// Add sender operation (debit)
		operations = append(operations, mesh.Operation{
			OperationIdentifier: mesh.OperationIdentifier{Index: 0},
			Type:               "TRANSFER",
			Account: &mesh.AccountIdentifier{
				Address: transfer.From,
			},
			Amount: &mesh.Amount{
				Value: "-" + transfer.Amount,
				Currency: mesh.Currency{
					Symbol:   transfer.TokenSymbol,
					Decimals: 18, // Default for ETH, would be different for other tokens
				},
			},
		})
		
		// Add recipient operation (credit)
		operations = append(operations, mesh.Operation{
			OperationIdentifier: mesh.OperationIdentifier{Index: 1},
			RelatedOperations:   []mesh.OperationIdentifier{{Index: 0}},
			Type:               "TRANSFER",
			Account: &mesh.AccountIdentifier{
				Address: transfer.To,
			},
			Amount: &mesh.Amount{
				Value: transfer.Amount,
				Currency: mesh.Currency{
					Symbol:   transfer.TokenSymbol,
					Decimals: 18, // Default for ETH, would be different for other tokens
				},
			},
		})
	}
	
	return operations
}

func mapPayloadOperations(req *PayloadsRequest) []mesh.Operation {
	// Similar to mapOperations but for payload requests
	// For simplicity in this PoC, we'll use a similar implementation
	operations := make([]mesh.Operation, 0)
	
	// For a simple transfer
	if req.Type == "TRANSFER" && len(req.Transfers) > 0 {
		transfer := req.Transfers[0]
		
		// Add sender operation (debit)
		operations = append(operations, mesh.Operation{
			OperationIdentifier: mesh.OperationIdentifier{Index: 0},
			Type:               "TRANSFER",
			Account: &mesh.AccountIdentifier{
				Address: transfer.From,
			},
			Amount: &mesh.Amount{
				Value: "-" + transfer.Amount,
				Currency: mesh.Currency{
					Symbol:   transfer.TokenSymbol,
					Decimals: 18, // Default for ETH, would be different for other tokens
				},
			},
		})
		
		// Add recipient operation (credit)
		operations = append(operations, mesh.Operation{
			OperationIdentifier: mesh.OperationIdentifier{Index: 1},
			RelatedOperations:   []mesh.OperationIdentifier{{Index: 0}},
			Type:               "TRANSFER",
			Account: &mesh.AccountIdentifier{
				Address: transfer.To,
			},
			Amount: &mesh.Amount{
				Value: transfer.Amount,
				Currency: mesh.Currency{
					Symbol:   transfer.TokenSymbol,
					Decimals: 18, // Default for ETH, would be different for other tokens
				},
			},
		})
	}
	
	return operations
}

func mapRequiredSigners(accounts []mesh.AccountIdentifier) []string {
	signers := make([]string, len(accounts))
	for i, account := range accounts {
		signers[i] = account.Address
	}
	return signers
}

func calculateTransactionFee(resp *mesh.ConstructionPreprocessResponse) string {
	// In a real implementation, this would calculate the fee based on the response
	// For simplicity in this PoC, we'll return a fixed fee
	return "0.001"
}

func generatePreparedTransaction(resp *mesh.ConstructionPreprocessResponse) map[string]interface{} {
	// In a real implementation, this would generate a prepared transaction
	// For simplicity in this PoC, we'll return the options from the response
	return resp.Options
}

func mapPublicKeys(keys []PublicKey) []mesh.PublicKey {
	meshKeys := make([]mesh.PublicKey, len(keys))
	for i, key := range keys {
		meshKeys[i] = mesh.PublicKey{
			HexBytes:  key.HexBytes,
			CurveType: key.CurveType,
		}
	}
	return meshKeys
}

func mapSigningPayloads(payloads []mesh.SigningPayload) []SigningPayload {
	result := make([]SigningPayload, len(payloads))
	for i, payload := range payloads {
		var address string
		if payload.AccountIdentifier != nil {
			address = payload.AccountIdentifier.Address
		}
		
		var sigType string
		if payload.SignatureType != nil {
			sigType = *payload.SignatureType
		}
		
		result[i] = SigningPayload{
			Address:       address,
			HexBytes:      payload.HexBytes,
			SignatureType: sigType,
		}
	}
	return result
}

func mapSignatures(signatures []Signature) []mesh.Signature {
	meshSigs := make([]mesh.Signature, len(signatures))
	for i, sig := range signatures {
		meshSigs[i] = mesh.Signature{
			SigningPayload: mesh.SigningPayload{
				HexBytes: sig.HexBytes,
			},
			PublicKey: mesh.PublicKey{
				HexBytes:  sig.PublicKey.HexBytes,
				CurveType: sig.PublicKey.CurveType,
			},
			SignatureType: sig.SignatureType,
			HexBytes:     sig.SignatureBytes,
		}
	}
	return meshSigs
}

func mapBalances(amounts []mesh.Amount) []Balance {
	balances := make([]Balance, len(amounts))
	for i, amount := range amounts {
		balances[i] = Balance{
			Amount:      amount.Value,
			TokenSymbol: amount.Currency.Symbol,
			Decimals:    amount.Currency.Decimals,
			Metadata:    amount.Metadata,
		}
	}
	return balances
}

func mapTransactions(txs []mesh.Transaction) []TransactionInfo {
	result := make([]TransactionInfo, len(txs))
	for i, tx := range txs {
		result[i] = TransactionInfo{
			TxID:     tx.TransactionIdentifier.Hash,
			Metadata: tx.Metadata,
		}
	}
	return result
}

// Overledger-specific method implementations

// GetOverledgerNetworks retrieves available networks from Overledger
func (s *service) GetOverledgerNetworks() (*overledger.NetworksResponse, error) {
	return s.overledgerClient.GetNetworks()
}

// GetOverledgerBalance retrieves account balance from Overledger
func (s *service) GetOverledgerBalance(networkID, address string) (*overledger.BalanceResponse, error) {
	return s.overledgerClient.GetAccountBalance(networkID, address)
}

// CreateOverledgerTransaction creates a transaction via Overledger
func (s *service) CreateOverledgerTransaction(req *overledger.TransactionRequest) (*overledger.TransactionResponse, error) {
	return s.overledgerClient.CreateTransaction(req)
}

// TestOverledgerConnection tests the connection to Overledger API
func (s *service) TestOverledgerConnection() error {
	return s.overledgerClient.TestConnection()
}

func mapMeshOperations(ops []mesh.Operation) []mesh.Operation {
	return ops
}

func mapAccount(acc *mesh.AccountIdentifier) *mesh.AccountIdentifier {
	return acc
}

func mapAmount(amt *mesh.Amount) *mesh.Amount {
	return amt
}