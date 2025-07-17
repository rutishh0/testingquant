\# Technical Requirements Document: Quant-to-Coinbase Mesh Connector Document Version: 1.0 \(Final\) 

Date: 6/17/2025 

Status: Approved for Production Baseline 



--- 



\#\#\# 1. System Architecture 



The Quant-to-Coinbase Mesh Connector is a containerized middleware service designed to translate high-level, DLT-agnostic API calls from the Quant Overledger standard into low-level, DLT-specific requests compatible with the Coinbase Mesh specification. The architecture is modular, ensuring separation of concerns and facilitating horizontal scalability. 



\#\#\#\# 1.1. High-Level Component Diagram 



The system comprises three core internal services orchestrated behind an API Gateway, and it interacts with an external-facing Developer Portal. 



\!\[A diagram showing the modular architecture of the Quant-to-Coinbase Mesh Connector. An external Client makes an API call to the API Gateway. The Gateway routes the request through an Authentication Service, which validates the API Key. The request then goes to the Translation Service, which uses a Unified Data Model and a DLT 

Adapter \(e.g., for Ethereum or XRP\) to convert the request. The Translation Service interacts with a State Management Database to track transaction status. The final, translated request is sent to the target DLT Network via Coinbase Mesh. A separate Webhook Polling Service monitors the DLT for transaction updates and updates the State Management Database. A Developer Portal exists separately, providing API documentation and onboarding tools for the Client.\]\(https://i.imgur.com/gK6pD4K.png\) 



\#\#\#\# 1.2. Component Responsibilities 



| Component | Responsibility | Technical Implementation Notes | 

| :--- | :--- | :--- | 

| API Gateway | Acts as the single entry point for all incoming client requests. Manages request routing, rate limiting, and SSL termination. | Leverages a reverse proxy configuration to route traffic to appropriate downstream services. | 

| Authentication Service | A dedicated proxy service that intercepts all requests to validate API keys. Manages key issuance, revocation, and usage policies. | Enforces per-key rate limits and access controls before forwarding valid requests. | 

| Translation Service | The core logic engine. It ingests requests based on the Unified Data Model, applies mapping rules, and uses DLT-specific adapters to generate a valid Coinbase Mesh request. | Stateless service designed for horizontal scaling. | 

| State Management Database | A persistent datastore \(e.g., PostgreSQL, Redis\) used to track the lifecycle of each transaction. This was a critical addition from "Project Phoenix" to resolve state-related architectural flaws found in the PoC. | Stores transaction ID, status \(e.g., \`PENDING\`, \`CONFIRMED\`, \`FAILED\`\), DLT transaction hash, and associated metadata. | 

| Webhook / Polling Service | An asynchronous service responsible for monitoring the target DLT for transaction confirmations and finality. It updates the transaction status in the State Management Database. | Uses a combination of webhooks \(if supported by Mesh\) and periodic polling of the DLT. | 

| Developer Portal | A public-facing static site that provides developers with API documentation, tutorials, code samples, and a self-service interface for obtaining API keys. | Synchronized with the API specification to ensure documentation is always current. | 



--- 



\#\#\# 2. Data Models & Mapping 



To bridge the conceptual gap between Quant Overledger and Coinbase Mesh, the connector uses a canonical Unified Data Model. All incoming requests are first normalized to this model before being translated to the target Mesh format. 



\#\#\#\# 2.1. Unified Data Model Schema 



\`\`\`json 

\{ 

"transactionId": "string \(UUID\)", 

"dlt": "string \(e.g., 'ethereum', 'xrp-ledger'\)", 

"transactionType": "enum \('TRANSFER', 'CONTRACT\_INVOKE'\)", 

"origin": \[ 

\{ 

"address": "string", 

"privateKeyIdentifier": "string \(Reference to a secure vault\)" 

\} 

\], 

"destination": \[ 

\{ 

"address": "string", 

"amount": "string \(Decimal format\)", 

"currency": "string \(e.g., 'ETH', 'XRP'\)" 

\} 

\], 

"contractDetails": \{ 

"contractAddress": "string", 

"functionName": "string", 

"functionParameters": "object" 

\}, 

"options": \{ 

"feePrice": "string \(Decimal format, e.g., Gwei\)", 

"feeLimit": "string \(Units\)", 

"callbackUrl": "string \(URL for status updates\)" 

\} 

\} 

\`\`\` 



\#\#\#\# 2.2. Field Mapping: Unified Model to Coinbase Mesh \(Ethereum Example\) The Translation Service applies the following rules to map the Unified Model to a request for the Ethereum network via Mesh. 



| Unified Model Field | Logic / Transformation | Coinbase Mesh Field \(Ethereum\) | 

| :--- | :--- | :--- | 

| òrigin\[0\].address\` | Direct mapping. | \`from\` | 

| \`destination\[0\].address\` | Direct mapping. | \`tò | 

| \`destination\[0\].amount\` | Convert decimal ÈTH\` value tòweì \(integer\). | \`valuè | 

| \`contractDetails.contractAddress\` | Direct mapping \(if \`transactionTypeìs 

\`CONTRACT\_INVOKÈ\). | \`tò | 

| \`contractDetails.functionNamè \+ \`functionParameters\` | ABI-encode the function signature and parameters. | \`datà | 

| òptions.feePricè | Convert decimal \`Gweì value tòweì \(integer\). | \`gasPricè | 

| òptions.feeLimit\` | Direct mapping. | \`gasLimit\` | 

| \`dlt\` | Used internally to select the correct DLT adapter. | N/A \(Implicit in Mesh network selection\) | 



--- 



\#\#\# 3. API Specifications 



The connector exposes a RESTful API for submitting and querying transactions. 



\#\#\#\# 3.1. Endpoints 



Submit a Transaction 

- \`POST /v1/transactions\` 

- Description: Submits a new transaction to be processed and sent to the specified DLT. The request is validated synchronously, but processed asynchronously. 

- Request Body: Conforms to the \[Unified Data Model Schema\]\(\#21-unified-data-model-schema\). 

- Success Response \(202 Accepted\): 

\`\`\`json 

\{ 

"status": "PENDING", 

"transactionId": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8", 

"submittedAt": "2025-06-17T11:12:37Z", 

"message": "Transaction accepted and is being processed." 

\} 

\`\`\` 



Get Transaction Status 

- \`GET /v1/transactions/\{transactionId\}\` 

- Description: Retrieves the current status and details of a previously submitted transaction. 

- Success Response \(200 OK\): 

\`\`\`json 

\{ 

"transactionId": "a1b2c3d4-e5f6-7890-g1h2-i3j4k5l6m7n8", 

"status": "CONFIRMED", 

"dlt": "ethereum", 

"dltTransactionHash": "0xabc123...", 

"confirmations": 12, 

"submittedAt": "2025-06-17T11:12:37Z", 

"completedAt": "2025-06-17T11:15:02Z" 

\} 

\`\`\` 



\#\#\#\# 3.2. Authentication 



- Method: API Key 

- Mechanism: The client must include an \`X-API-Key\` header with every request. 

\`\`\` 

X-API-Key: <your-assigned-api-key> 

\`\`\` 

- Management: Keys are provisioned via the Developer Portal and managed by the Authentication Service, which enforces all access policies. 



\#\#\#\# 3.3. Error Codes 



A standardized error response format is used for all client-facing errors. 



\`\`\`json 

\{ 

"error": \{ 

"code": "VALIDATION\_ERROR", 

"message": "Invalid DLT specified.", 

"details": \[ 

\{ 

"field": "dlt", 

"issue": "Value 'bitcoiin' is not a supported DLT. Must be one of: \[ethereum, xrp-ledger\]." 

\} 

\] 

\} 

\} 

\`\`\` 



| HTTP Status | Error Code | Description | 

| :--- | :--- | :--- | 

| 400 Bad Request | \`VALIDATION\_ERROR\` | The request body fails schema validation \(e.g., incorrect data type, missing required field\). | 

| 401 Unauthorized | ÀUTH\_ERROR\` | ThèX-API-Key\` header is missing, invalid, or expired. | 

| 403 Forbidden | \`PERMISSION\_DENIED\` | The API key is valid but does not have permission to perform the requested action. | 

| 404 Not Found | \`NOT\_FOUND\` | The requested resource \(e.g., a specific 

\`transactionId\`\) does not exist. | 

| 429 Too Many Requests | \`RATE\_LIMIT\_EXCEEDED\` | The client has exceeded the request rate limit associated with their API key. | 

| 500 Internal Server Error | ÌNTERNAL\_ERROR\` | An unexpected server-side error occurred. The response will not contain sensitive details. | 



--- 



\#\#\# 4. Technical Implementation Details 



\#\#\#\# 4.1. Smart Contract Interaction Logic 

For \`CONTRACT\_INVOKÈ transactions, the Translation Service dynamically constructs the transaction \`datà payload. It uses an ABI \(Application Binary Interface\) encoder library to correctly serialize thèfunctionNameànd \`functionParameters\` from the Unified Model into a hex-encoded string as required by the Ethereum Virtual Machine \(EVM\). 



\#\#\#\# 4.2. Multi-DLT Adapter Design 

The system is built on a provider pattern to support multiple DLTs. 

- A common ÌDLTAdapterìnterface defines methods likècreateTransaction\(...\)\`, 

\`getFeeData\(...\)\`, and \`getConfirmationStatus\(...\)\`. 

- Concrete implementations \(ÈthereumAdapter\`, \`XRPAdapter\`, etc.\) encapsulate the specific logic for interacting with each DLT via the Mesh standard. 

- Thèdlt\` field in the incoming request determines which adapter is instantiated by a factory class. 



\#\#\#\# 4.3. Centralized Error Handling Middleware A global middleware layer is implemented in the API Gateway and backend services. Its responsibilities are: 

1. Catch all exceptions, whether application-level or system-level. 

2. Log the full, detailed error \(including stack trace\) to a secure, internal logging system \(e.g., ELK Stack\) for debugging. 

3. Generate a sanitized, user-friendly error response according to the \[Error Codes\]\(\#33-error-codes\) specification, ensuring no internal system details are leaked. 



\#\#\#\# 4.4. Currency Unit Conversion 

The system requires all monetary values in API requests to be in standard decimal format \(e.g., "1.5" ETH\). The selected DLT adapter is responsible for converting these values into the DLT's smallest native integer unit before building the transaction. 

- Ethereum: Converts ETH tòweì \(1 ETH = 10^18 wei\). 

- XRP Ledger: Converts XRP tòdrops\` \(1 XRP = 10^6 drops\). 



--- 



\#\#\# 5. Security Requirements 



| Control | Requirement | 

| :--- | :--- | 

| Input Validation | All incoming API requests MUST be strictly validated against the OpenAPI schema. This includes data types, formats \(e.g., \`0x\` prefix for hex\), and value ranges. Any deviation MUST result in à400 Bad Request\`. | 

| Authentication | All endpoints MUST be protected by the Authentication Service. 

Unauthenticated requests MUST be rejected with à401 Unauthorized\` status. | 

| Authorization | The Authentication Service MUST enforce access policies on a per-key basis \(e.g., allow/deny access to specific DLTs or transaction types\). Unauthorized actions MUST be rejected with à403 Forbidden\`. | 

| Information Leakage | Error responses sent to the client MUST NEVER contain sensitive system information, such as stack traces, file paths, or internal infrastructure details. | 

| Secret Management | All sensitive credentials, such as database connection strings or internal service keys, MUST be managed through a secure vault service \(e.g., HashiCorp Vault, AWS Secrets Manager\) and injected at runtime. They MUST NOT be stored in source code. | 

| Transport Security | All communication between the client and the API Gateway MUST 

be encrypted using TLS 1.2 or higher. | 



--- 



\#\#\# 6. Performance & Scalability \(Non-Functional Requirements\) Based on benchmark testing conducted in project Phase 1, the system must meet the following NFRs. 



| Metric | Requirement | 

| :--- | :--- | 

| API Response Time \(P95\) | < 200ms for synchronous validation requests \(\`POST 

/v1/transactions\`\). | 

| API Response Time \(P95\) | < 150ms for status retrieval requests \(\`GET 

/v1/transactions/\{id\}\`\). | 

| Transaction Throughput | The system must sustain a continuous load of 50 

transactions per second \(TPS\) without performance degradation or an increase in error rates. | 

| Scalability | The architecture must support horizontal scaling. Deploying additional container instances for the Translation Service must result in a proportional increase in transaction throughput. | 

| Availability | The system must maintain an uptime of 99.95%. | 



--- 



\#\#\# 7. Testing and Validation Strategy 



The following testing strategies, successfully used throughout the project, form the baseline for all future regression testing suites. Any code change must pass all relevant tests before being deployed to production. 



| Test Type | Scope & Purpose | 

| :--- | :--- | 

| Unit Testing | Focuses on individual functions and modules in isolation. Validates data transformations \(e.g., currency conversion\), ABI encoding, and DLT adapter logic using mocks. | 

| Integration Testing | Verifies the interaction between internal services. Tests the complete flow from the API Gateway through the Authentication and Translation services, mocking the final DLT network call. | 

| End-to-End \(E2E\) Testing | Tests the entire system workflow against a live DLT testnet. A test client submits a transaction via the API, and the test suite polls the status endpoint until \`CONFIRMED\`, then verifies the transaction on the testnet blockchain explorer. | 

| User Acceptance Testing \(UAT\) | Validates the system from a developer's perspective. 

Involves following tutorials on the Developer Portal, obtaining an API key, and successfully submitting a transaction using provided code samples. |



