package util

import (
	"log"

	vault "github.com/hashicorp/vault/api"
)

// global test helpers
const (
	VaultAddress = "http://127.0.0.1:8200"
	VaultToken   = "abcdefghijklmnopqrstuvwxyz09"
	KVPath       = "foo/bar"
	KVKey        = "password"
	KVValue      = "supersecret"
	KV1Mount     = "kv"
	KV2Mount     = "secret"
)

var (
	VaultClient         = basicVaultClient()
	RoleID, SecretID, _ = approleAttrs()
)

// helper for basic vault client
func basicVaultClient() *vault.Client {
	vaultConfig := &vault.Config{Address: VaultAddress}
	vaultConfig.ConfigureTLS(&vault.TLSConfig{Insecure: true})
	client, _ := vault.NewClient(vaultConfig)
	client.SetToken(VaultToken)

	return client
}

// helper for approle auth
func approleAttrs() (string, string, error) {
	// retrieve role id and secret id for testing approle auth in "push" mode
	roleID, err := VaultClient.Logical().Read("auth/approle/role/myAppRole/role-id")
	if err != nil {
		log.Print("failed to retrieve role ID for approle auth")
		return "", "", err
	}
	secretID, err := VaultClient.Logical().Write("auth/approle/role/myAppRole/secret-id", nil)
	if err != nil {
		log.Print("failed to retrieve secret ID for approle auth")
		return "", "", err
	}

	return roleID.Data["role_id"].(string), secretID.Data["secret_id"].(string), nil
}
