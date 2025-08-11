package clients

import (
	"context"
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
	client exchclient.RestClient
}

// NewExchangeClient initialises a RestClient using credentials from the
// environment. Preferred source is EXCHANGE_CREDENTIALS – a JSON blob with
// apiKey, passphrase and signingKey – matching Coinbase's quick-start guide.
// If that variable is absent or malformed, we fall back to individual vars.
// Only the API key and signing secret are mandatory – Coinbase’s newer UI no
// longer provides a passphrase, so we treat an empty passphrase as acceptable.
//
//	COINBASE_API_KEY, COINBASE_API_PASSPHRASE (optional), COINBASE_API_SECRET
func NewExchangeClient() (*ExchangeClient, error) {
	creds, err := credentials.ReadEnvCredentials("EXCHANGE_CREDENTIALS")
	if err != nil {
		// Attempt fallback
		apiKey := os.Getenv("COINBASE_API_KEY")
		pass := os.Getenv("COINBASE_API_PASSPHRASE") // may be blank
		secret := os.Getenv("COINBASE_API_SECRET")

		// API key and secret are mandatory; passphrase may be empty.
		if apiKey == "" || secret == "" {
			return nil, fmt.Errorf("missing exchange credentials: %v", err)
		}

		creds = &credentials.Credentials{
			ApiKey:     apiKey,
			Passphrase: pass,
			SigningKey: secret,
		}
	}

	httpCli, err := core.DefaultHttpClient()
	if err != nil {
		return nil, err
	}

	c := exchclient.NewRestClient(creds, httpCli)
	return &ExchangeClient{client: c}, nil
}

// ListAccounts returns the authenticated accounts for the user associated with
// the provided credentials.
func (e *ExchangeClient) ListAccounts(ctx context.Context) (*accounts.ListAccountsResponse, error) {
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
