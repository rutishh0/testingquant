package utils

// During the Mesh migration, Coinbase CDP authentication is removed.
// Mesh endpoints do not require CDP JWT headers. This helper returns the
// minimal JSON header set so existing client code can continue to call it.
func GenerateAuthHeaders(method, path string) (map[string]string, error) {
    return map[string]string{
        "Content-Type": "application/json",
    }, nil
}
