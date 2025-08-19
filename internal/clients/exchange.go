package clients

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/coinbase-samples/core-go"
	"github.com/coinbase-samples/exchange-sdk-go/accounts"
	exchclient "github.com/coinbase-samples/exchange-sdk-go/client"
	"github.com/coinbase-samples/exchange-sdk-go/credentials"
	"github.com/coinbase-samples/exchange-sdk-go/products"
)

// ExchangeClient provides a minimal wrapper around the Coinbase Exchange Go SDK
// exposing only the endpoints we currently need for the Quant Connector backend.
// Additional endpoints can be added easily as the project grows.
type ExchangeClient struct {
	client       exchclient.RestClient
	noAuthMode   bool // indicates if client is in no-auth mode for read-only operations
	configStatus string // tracks configuration status for better error reporting
}

// Configuration error types for clearer distinction
var (
	// ErrExchangeNotConfigured indicates that no exchange credentials are set at all.
	ErrExchangeNotConfigured = errors.New("exchange credentials not configured")

	// ErrExchangeMisconfigured indicates that exchange credentials are provided but invalid/malformed.
	ErrExchangeMisconfigured = errors.New("exchange credentials are misconfigured")

	// ErrExchangePartialConfig indicates partial credentials are provided (useful for debugging).
	ErrExchangePartialConfig = errors.New("exchange credentials partially configured")

	// ErrExchangeNoAuthUnsupported indicates that no-auth mode doesn't support the requested operation.
	ErrExchangeNoAuthUnsupported = errors.New("operation requires authentication but client is in no-auth mode")
)

// NewExchangeClient initialises a RestClient using credentials from the
// environment. Preferred source is EXCHANGE_CREDENTIALS – a JSON blob with
// apiKey, passphrase and signingKey – matching Coinbase's quick-start guide.
// If that variable is absent or malformed, we fall back to individual vars.
// Only the API key and signing secret are mandatory – Coinbase's newer UI no
// longer provides a passphrase, so we treat an empty passphrase as acceptable.
//
//	COINBASE_API_KEY, COINBASE_API_PASSPHRASE (optional), COINBASE_API_SECRET
//
// Supports an optional NO_AUTH path for read-only operations when
// EXCHANGE_NO_AUTH=true is set. In no-auth mode, authenticated endpoints will
// return ErrExchangeNoAuthUnsupported, while public endpoints (e.g., products)
// may still function if the SDK supports unauthenticated access.
//
// Returns ErrExchangeNotConfigured if no credentials are found and no-auth is disabled.
// Returns ErrExchangeMisconfigured if credentials are provided but invalid.
func NewExchangeClient() (*ExchangeClient, error) {
	noAuth := os.Getenv("EXCHANGE_NO_AUTH") == "true"

	creds, err := credentials.ReadEnvCredentials("EXCHANGE_CREDENTIALS")
	if err != nil {
		// Attempt fallback
		apiKey := os.Getenv("COINBASE_API_KEY")
		pass := os.Getenv("COINBASE_API_PASSPHRASE") // may be blank
		secret := os.Getenv("COINBASE_API_SECRET")

		// Check if any exchange credentials are set at all
		if apiKey == "" && secret == "" {
			if noAuth {
				// Continue without creds; some endpoints will be unavailable
				httpCli, err := core.DefaultHttpClient()
				if err != nil {
					return nil, fmt.Errorf("%w: failed to create HTTP client: %v", ErrExchangeMisconfigured, err)
				}
				c := exchclient.NewRestClient(nil, httpCli)
				return &ExchangeClient{client: c, noAuthMode: true, configStatus: "no_auth"}, nil
			}
			return nil, ErrExchangeNotConfigured
		}

		// If some credentials are set but incomplete, it's misconfigured
		if apiKey == "" || secret == "" {
			return nil, fmt.Errorf("%w: API key and secret are required", ErrExchangeMisconfigured)
		}

		creds = &credentials.Credentials{
			ApiKey:     apiKey,
			Passphrase: pass,
			SigningKey: secret,
		}
	}

	httpCli, err := core.DefaultHttpClient()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create HTTP client: %v", ErrExchangeMisconfigured, err)
	}

	c := exchclient.NewRestClient(creds, httpCli)
	return &ExchangeClient{client: c, noAuthMode: false, configStatus: "auth"}, nil
}

// ListAccounts returns the authenticated accounts for the user associated with
// the provided credentials.
func (e *ExchangeClient) ListAccounts(ctx context.Context) (*accounts.ListAccountsResponse, error) {
	if e.noAuthMode {
		return nil, ErrExchangeNoAuthUnsupported
	}
	svc := accounts.NewAccountsService(e.client)
	return svc.ListAccounts(ctx, &accounts.ListAccountsRequest{})
}

// ListProducts returns all tradeable Exchange products (e.g. BTC-USD).
func (e *ExchangeClient) ListProducts(ctx context.Context) (*products.ListProductsResponse, error) {
	svc := products.NewProductsService(e.client)
	return svc.ListProducts(ctx, &products.ListProductsRequest{})
}

// GetProduct returns details for a single product.
func (e *ExchangeClient) GetProduct(ctx context.Context, productID string) (*products.GetProductResponse, error) {
	svc := products.NewProductsService(e.client)
	return svc.GetProduct(ctx, &products.GetProductRequest{ProductId: productID})
}
