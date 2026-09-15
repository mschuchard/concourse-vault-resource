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
	if _, ok := auths["kubernetes/"]; ok {
		test.Skip("Vault server already bootstrapped; skipping")
	}

	// enable auth: approle, aws, kubernetes (token enabled by default with dev server)
	if err := VaultClient.Sys().EnableAuthWithOptions("approle", &vault.EnableAuthOptions{Type: "approle"}); err != nil {
		test.Fatalf("failed to enable approle auth: %s", err)
	}
	if _, err := VaultClient.Logical().Write("auth/approle/role/"+AppRole, map[string]any{
		"token_policies": "default",
		"token_ttl":      "1h",
		"token_max_ttl":  "4h",
	}); err != nil {
		test.Fatalf("failed to configure approle auth: %s", err)
	}
	if err := VaultClient.Sys().EnableAuthWithOptions("aws", &vault.EnableAuthOptions{Type: "aws"}); err != nil {
		test.Fatalf("failed to enable aws auth: %s", err)
	}
	if err := VaultClient.Sys().EnableAuthWithOptions("kubernetes", &vault.EnableAuthOptions{Type: "kubernetes"}); err != nil {
		test.Fatalf("failed to enable kubernetes auth: %s", err)
	}

	// enable secrets: database, aws, kv1 (kv2 enabled by default with dev server)
	VaultClient.Sys().Mount("aws/", &vault.MountInput{Type: "aws"})
	VaultClient.Sys().Mount("database/", &vault.MountInput{Type: "database"})
	VaultClient.Sys().Mount(KV1Mount, &vault.MountInput{Type: "kv"})

	// modify new kv secrets engine to be version 1
	pluginVersion := "1"
	VaultClient.Sys().TuneMountAllowNil(KV1Mount, vault.TuneMountConfigInput{PluginVersion: &pluginVersion})

	// put kv1 and kv2 secrets
	VaultClient.KVv1(KV1Mount).Put(
		context.Background(),
		KVPath,
		map[string]any{KVKey: KVValue},
	)
	VaultClient.KVv2(KV2Mount).Put(
		context.Background(),
		KVPath,
		map[string]any{KVKey: KVValue, "other_password": "ultrasecret"},
	)
	// for full "in" test
	VaultClient.KVv2(KV2Mount).Put(
		context.Background(),
		"bar/baz",
		map[string]any{KVKey: KVValue},
	)
}
