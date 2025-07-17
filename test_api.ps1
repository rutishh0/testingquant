# Test script for Quant-Mesh Connector API

Write-Host "Testing Quant-Mesh Connector API..."
Write-Host "================================="

# Test Health Endpoint
Write-Host "\n1. Testing Health Endpoint:"
try {
    $healthResponse = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET
    Write-Host "✓ Health Check: $($healthResponse | ConvertTo-Json)"
} catch {
    Write-Host "✗ Health Check Failed: $($_.Exception.Message)"
}

# Test Status Endpoint
Write-Host "\n2. Testing Status Endpoint:"
try {
    $statusResponse = Invoke-RestMethod -Uri "http://localhost:8080/status" -Method GET
    Write-Host "✓ Status Check: $($statusResponse | ConvertTo-Json)"
} catch {
    Write-Host "✗ Status Check Failed: $($_.Exception.Message)"
}

# Test API with API Key
Write-Host "\n3. Testing API Endpoint with API Key:"
$headers = @{
    "X-API-Key" = "test-api-key-1234567890"
    "Content-Type" = "application/json"
}

$testPayload = @{
    "network_identifier" = @{
        "blockchain" = "ethereum"
        "network" = "mainnet"
    }
    "account_identifier" = @{
        "address" = "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b"
    }
} | ConvertTo-Json -Depth 3

try {
    $balanceResponse = Invoke-RestMethod -Uri "http://localhost:8080/v1/account/balance" -Method POST -Headers $headers -Body $testPayload
    Write-Host "✓ Balance API: $($balanceResponse | ConvertTo-Json)"
} catch {
    Write-Host "✗ Balance API Failed: $($_.Exception.Message)"
}

Write-Host "\n================================="
Write-Host "API Testing Complete!"