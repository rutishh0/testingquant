# Overledger API Reference

This document provides a comprehensive overview and reference for the Overledger API, detailing its various endpoints and functionalities.

## GET STARTED

### How to Access Overledger APIs
- **Base URL**: `https://api.overledger.dev` (for both testnet and mainnet).
- **API Versioning**:
    - `v2`: Included in the URL path (e.g., `/v2/...`).
    - `v3+`: Specified via the `Api-Version` header.
- **Key Patterns**:
    - **Prepare-Sign-Execute**: For state-changing DLT interactions. A `prepare` call returns data to be signed, and an `execute` call submits the signed transaction.
    - **Autoexecute**: For read-only queries, simplifying the process into a single call.
- **Location Object**: Used to specify the target blockchain `technology` and `network`.

### OAuth2 Token API
- **Endpoint**: `POST /oauth2/token`
- **Description**: Authenticates and retrieves an OAuth2 bearer token required for all other API calls.
- **Authentication**: Uses Basic Authentication with `clientId` and `clientSecret`.
- **Usage**: The retrieved token must be included in the `Authorization` header of subsequent requests as a `Bearer` token.

## TRANSACT API
- **Description**: Endpoints for creating and executing DLT transactions.
- **Endpoints**:
    - `POST /v2/preparation/transaction`: Prepares a standard DLT transaction for signing.
    - `POST /v2/execution/transaction`: Executes a signed transaction.
    - `POST /v2/preparation/nativetransaction`: Prepares a native DLT transaction for advanced users.
    - `POST /v2/execution/nativetransaction`: Executes a signed native transaction.

## SEARCH API
- **Description**: Endpoints for querying data from DLTs.
- **Endpoints**:
    - `POST /v2/autoexecution/search/transaction`: Auto-executes a search for a transaction.
    - `POST /v2/autoexecution/search/address`: Auto-executes a search for an address balance and sequence numbers.
    - `GET /v2/search/block/{blockId}`: Searches for a block by its ID or hash.
    - `POST /v2/autoexecution/search/utxo`: Auto-executes a search for unspent transaction outputs (UTXOs) at a given address.

## WEBHOOKS API
- **Description**: Manages webhooks for receiving asynchronous event notifications.
- **Endpoints**:
    - `POST /v2/webhook/subscription`: Creates a new webhook subscription.
    - `PATCH /v2/webhook/subscription/{webhookId}`: Updates an existing webhook subscription.
    - `GET /v2/webhook/subscription`: Lists all active webhook subscriptions.
    - `DELETE /v2/webhook/subscription/{webhookId}`: Deletes a webhook subscription.

## SMART CONTRACT API
- **Description**: Endpoints for interacting with smart contracts.
- **Endpoints**:
    - `POST /v2/autoexecution/search/smartcontract`: Auto-executes a read query on a smart contract.
    - `POST /v2/preparation/smartcontract`: Prepares a smart contract function call for signing.
    - `POST /v2/execution/smartcontract`: Executes a signed smart contract transaction.

## TOKEN API
- **Description**: Endpoints for managing and querying fungible and non-fungible tokens.
- **Endpoints**:
    - `GET /v2/tokens`: Retrieves a list of supported fungible tokens (e.g., ERC20, QRC20).
    - `GET /v2/tokens/nfts`: Retrieves a list of supported non-fungible tokens (e.g., ERC721, QRC721).

## AUTHORISE API
- **Description**: Endpoints for managing permissions and authorising accounts for specific actions.
- **Endpoints**:
    - `POST /v2/preparation/authorise`: Prepares an authorisation transaction for signing.
    - `POST /v2/execution/authorise`: Executes a signed authorisation transaction.

## BRIDGE API
- **Description**: Facilitates the transfer of assets between different DLTs.
- **Endpoints**:
    - `POST /v2/preparation/assettransfer`: Prepares an asset transfer for signing.
    - `POST /v2/execution/assettransfer`: Executes a signed asset transfer.
    - `GET /v2/assettransfer/{assetTransferId}`: Retrieves details of a specific asset transfer.

## Overledger Preview Sandboxes

### Digital Currency Sandbox
- **Description**: A sandbox for experimenting with digital currencies and bonds.
- **Endpoints**: Reading token/bond balances, issuing, purchasing, redeeming bonds, and transferring tokens.

### SATP (Secure Asset Transfer Protocol) Sandbox
- **Description**: A multi-step workflow for secure asset transfers.
- **Workflow**: Involves initiation, proposal, commencement, lock assertion, commitment, and completion.

### SAEP (Secure Asset Exchange Protocol) Sandbox
- **Description**: A multi-step workflow for secure, atomic asset exchanges (swaps).
- **Workflow**: Follows a similar multi-step process to SATP for proposing, locking, and committing the exchange.

## Overledger Preview APIs

### Connector: Hyperledger Fabric
- **Description**: Interacts with Hyperledger Fabric networks.
- **Functionality**: Prepare/execute smart contract transactions, read data, create accounts, and search blocks/transactions.

### App: Advanced Access Control
- **Description**: Adds a validation layer to smart contract interactions.
- **Functionality**: Create, manage, and enforce validation schemas for smart contract writes and deployments.

### App: Automated Smart Contract Audit
- **Description**: Triggers an automated audit of a smart contract.
- **Endpoint**: `POST /v2/automated-audit`

### Connector: Solana
- **Description**: Interacts with the Solana blockchain.
- **Functionality**: Read account balances and prepare/execute transactions.

### Connectors: Combined Calls
- **Description**: Unified endpoints for reading data across different DLTs.
- **Endpoints**: `GET /v2/read/transaction` and `GET /v2/read/block`.

### Overwallet
- **Description**: A service for managing Hierarchical Deterministic (HD) wallets.
- **Functionality**: Create HD accounts, generate keys, sign transactions, and retrieve asset information. Note: Some documentation links for retrieving addresses and public keys were found to be broken.
