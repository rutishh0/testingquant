package services

import (
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "math/big"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"
)

// Default hardcoded Infura Sepolia endpoint (replace YOUR_INFURA_PROJECT_ID or set ENV vars)
const defaultInfuraSepoliaURL = "https://sepolia.infura.io/v3/YOUR_INFURA_PROJECT_ID"

type rpcError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

type rpcRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      int         `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params"`
}

type rpcResponse struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      int             `json:"id"`
    Result  json.RawMessage `json:"result"`
    Error   *rpcError       `json:"error,omitempty"`
}

type EthRPCClient struct {
    URL        string
    httpClient *http.Client
}

func NewEthRPCFromEnv() (*EthRPCClient, error) {
    url := os.Getenv("INFURA_RPC_URL")
    if url == "" {
        url = os.Getenv("ETH_RPC_URL")
    }
    // If not provided via ENV, fall back to a hardcoded Sepolia endpoint placeholder
    if url == "" {
        url = defaultInfuraSepoliaURL
        // Note: This is a placeholder. Replace YOUR_INFURA_PROJECT_ID or set ENV vars.
    }
    return &EthRPCClient{
        URL: url,
        httpClient: &http.Client{Timeout: 20 * time.Second},
    }, nil
}

func (c *EthRPCClient) call(method string, params interface{}, out interface{}) error {
    if c == nil {
        return errors.New("nil EthRPCClient")
    }
    reqBody := rpcRequest{JSONRPC: "2.0", ID: 1, Method: method, Params: params}
    b, err := json.Marshal(reqBody)
    if err != nil {
        return err
    }
    httpReq, err := http.NewRequest("POST", c.URL, bytes.NewReader(b))
    if err != nil {
        return err
    }
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return fmt.Errorf("rpc http status %d", resp.StatusCode)
    }

    var r rpcResponse
    if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
        return err
    }
    if r.Error != nil {
        return fmt.Errorf("rpc error %d: %s", r.Error.Code, r.Error.Message)
    }
    if out != nil {
        if err := json.Unmarshal(r.Result, out); err != nil {
            return err
        }
    }
    return nil
}

// Helpers
func hexToInt64(h string) (int64, error) {
    if h == "" {
        return 0, errors.New("empty hex string")
    }
    // support 0x prefix
    n, err := strconv.ParseInt(strings.TrimPrefix(h, "0x"), 16, 64)
    if err != nil {
        return 0, err
    }
    return n, nil
}

func hexToBigInt(h string) (*big.Int, error) {
    if h == "" {
        return big.NewInt(0), errors.New("empty hex string")
    }
    x := new(big.Int)
    _, ok := x.SetString(strings.TrimPrefix(h, "0x"), 16)
    if !ok {
        return big.NewInt(0), fmt.Errorf("invalid hex: %s", h)
    }
    return x, nil
}

func int64ToHex(n int64) string { return fmt.Sprintf("0x%x", n) }

func bigMul(a, b *big.Int) *big.Int { z := new(big.Int).Mul(a, b); return z }

func bigToStringWei(x *big.Int) string { if x == nil { return "0" }; return x.String() }

// Types for decoding JSON-RPC responses

type rpcBlock struct {
    Number           string          `json:"number"`
    Hash             string          `json:"hash"`
    ParentHash       string          `json:"parentHash"`
    Timestamp        string          `json:"timestamp"`
    Transactions     json.RawMessage `json:"transactions"` // can be []tx or []hash
}

type rpcTx struct {
    Hash      string `json:"hash"`
    From      string `json:"from"`
    To        string `json:"to"`
    Value     string `json:"value"`
    Gas       string `json:"gas"`
    GasPrice  string `json:"gasPrice"`
    Nonce     string `json:"nonce"`
}

type rpcReceipt struct {
    Status             string `json:"status"`
    GasUsed            string `json:"gasUsed"`
    EffectiveGasPrice  string `json:"effectiveGasPrice"`
    BlockNumber        string `json:"blockNumber"`
    BlockHash          string `json:"blockHash"`
}