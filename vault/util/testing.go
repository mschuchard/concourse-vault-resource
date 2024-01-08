package util

import vault "github.com/hashicorp/vault/api"

// global test helpers
const (
	VaultAddress = "http://127.0.0.1:8200"
	VaultToken   = "abcdefghijklmnopqrstuvwxyz09"
)

var (
	VaultClient = basicVaultClient()
)

// helper for basic vault client
func basicVaultClient() *vault.Client {
	vaultConfig := &vault.Config{Address: VaultAddress}
	vaultConfig.ConfigureTLS(&vault.TLSConfig{Insecure: true})
	client, _ := vault.NewClient(vaultConfig)
	client.SetToken(VaultToken)

	return client
}
