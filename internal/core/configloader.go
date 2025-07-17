package core

import (
    "os"

    "gopkg.in/yaml.v3"
)

// LoadConnectorConfigs reads a YAML file where each top-level key is a connector
// ID and the value is a free-form map passed to Connector.Init().
// Example YAML:
// mesh:
//   base_url: "http://localhost:8081"
// overledger:
//   base_url: "https://api.overledger.dev"
//   client_id: "abc"
//   client_secret: "def"
//   auth_url: "https://auth.overledger.dev/oauth2/token"
func LoadConnectorConfigs(path string) (map[string]map[string]any, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var out map[string]map[string]any
    if err := yaml.Unmarshal(data, &out); err != nil {
        return nil, err
    }
    return out, nil
}
