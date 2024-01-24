package util

import (
	"context"
	"testing"

	vault "github.com/hashicorp/vault/api"
)

// bootstrap vault server for testing
func TestBootstrap(test *testing.T) {
	// check if we should skip bootstrap
	auths, _ := VaultClient.Sys().ListAuth()
	if _, ok := auths["auth/aws/"]; ok {
		test.Skip("Vault server already bootstrapped; skipping")
	}

	// enable auth: aws
	VaultClient.Sys().EnableAuthWithOptions("auth/aws", &vault.EnableAuthOptions{Type: "aws"})
	// enable secrets: database, aws, kv1 (kv2 enabled by default with dev server)
	VaultClient.Sys().Mount("aws/", &vault.MountInput{Type: "aws"})
	VaultClient.Sys().Mount("database/", &vault.MountInput{Type: "database"})
	VaultClient.Sys().Mount(KV1Mount, &vault.MountInput{Type: "kv"})
	// modify new kv secrets engine to be version 1
	VaultClient.Sys().TuneMount(KV1Mount, vault.MountConfigInput{PluginVersion: "1"})
	// put kv1 and kv2 secrets
	VaultClient.KVv1(KV1Mount).Put(
		context.Background(),
		KVPath,
		map[string]interface{}{KVKey: KVValue},
	)
	VaultClient.KVv2(KV2Mount).Put(
		context.Background(),
		KVPath,
		map[string]interface{}{KVKey: KVValue},
	)
	// for full "in" test
	VaultClient.KVv2(KV2Mount).Put(
		context.Background(),
		"bar/baz",
		map[string]interface{}{KVKey: KVValue},
	)
}
