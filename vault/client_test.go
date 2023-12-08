package vault

import (
	"context"
	"testing"

	vault "github.com/hashicorp/vault/api"
)

// global test helpers
const (
	testVaultAddress = "http://127.0.0.1:8200"
	testVaultToken   = "abcdefghijklmnopqrstuvwxyz09"
)

var basicVaultConfig = &VaultConfig{
	Address:  testVaultAddress,
	Token:    testVaultToken,
	Insecure: true,
}

// test config constructor
func TestNewVaultConfig(test *testing.T) {
	basicVaultConfig.New()

	if basicVaultConfig.Engine != token || basicVaultConfig.Address != testVaultAddress || len(basicVaultConfig.AWSMountPath) != 0 || len(basicVaultConfig.AWSRole) != 0 || basicVaultConfig.Token != testVaultToken || !basicVaultConfig.Insecure {
		test.Error("the Vault config constructor returned unexpected values.")
		test.Errorf("expected Auth Engine: %s, actual: %s", token, basicVaultConfig.Engine)
		test.Errorf("expected Vault Address: %s, actual: %s", testVaultAddress, basicVaultConfig.Address)
		test.Errorf("expected AWS Mount Path: (empty), actual: %s", basicVaultConfig.AWSMountPath)
		test.Errorf("expected AWS IAM Role: (empty), actual: %s", basicVaultConfig.AWSRole)
		test.Errorf("expected Vault Token: %s, actual: %s", testVaultToken, basicVaultConfig.Token)
		test.Errorf("expected Vault Insecure: %t, actual: %t", basicVaultConfig.Insecure, basicVaultConfig.Insecure)
	}
}

// test client error messages

// test client token authentication
func TestAuthClient(test *testing.T) {
	basicVaultClient := basicVaultConfig.AuthClient()

	if basicVaultClient.Token() != testVaultToken {
		test.Error("the authenticated Vault client return failed basic validation")
		test.Errorf("expected Vault token: %s, actual: %s", testVaultToken, basicVaultClient.Token())
	}
}

// bootstrap vault server for testing
func TestBootstrap(test *testing.T) {
	// instantiate client for bootstrapping
	basicVaultConfig.New()
	client := basicVaultConfig.AuthClient()

	// check if we should skip bootstrap
	auths, _ := client.Sys().ListAuth()
	if _, ok := auths["auth/aws/"]; ok {
		test.Skip("Vault server already bootstrapped; skipping")
	}

	// enable auth: aws
	client.Sys().EnableAuthWithOptions("auth/aws", &vault.EnableAuthOptions{Type: "aws"})
	// enable secrets: database, aws, kv1 (kv2 enabled by default with dev server)
	client.Sys().Mount("aws/", &vault.MountInput{Type: "aws"})
	client.Sys().Mount("database/", &vault.MountInput{Type: "database"})
	client.Sys().Mount(KV1Mount, &vault.MountInput{Type: "kv"})
	// modify new kv secrets engine to be version 1
	client.Sys().TuneMount(KV1Mount, vault.MountConfigInput{PluginVersion: "1"})
	// put kv1 and kv2 secrets
	client.KVv1(KV1Mount).Put(
		context.Background(),
		KVPath,
		map[string]interface{}{KVKey: KVValue},
	)
	client.KVv2(KV2Mount).Put(
		context.Background(),
		KVPath,
		map[string]interface{}{KVKey: KVValue},
	)
	// for full "in" test
	client.KVv2(KV2Mount).Put(
		context.Background(),
		"bar/baz",
		map[string]interface{}{KVKey: KVValue},
	)
}
