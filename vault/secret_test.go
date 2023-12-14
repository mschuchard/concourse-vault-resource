package vault

import "testing"

// globals for vault package testing
const (
	KVPath   = "foo/bar"
	KVKey    = "password"
	KVValue  = "supersecret"
	KV1Mount = "kv"
	KV2Mount = "secret"
)

// test secret constructor
func TestNewVaultSecret(test *testing.T) {
	dbVaultSecret := NewVaultSecret("database", "", KVPath)
	if dbVaultSecret.engine != database || dbVaultSecret.path != KVPath || dbVaultSecret.mount != "database" || dbVaultSecret.dynamic != true {
		test.Error("the database Vault secret constructor returned unexpected values")
		test.Errorf("expected engine: %s, actual: %s", dbVaultSecret.engine, database)
		test.Errorf("expected path: %s, actual: %s", dbVaultSecret.path, KVPath)
		test.Errorf("expected mount: %s, actual: %s", dbVaultSecret.mount, "database")
		test.Errorf("expected dynamic to be true, actual: %t", dbVaultSecret.dynamic)
	}

	awsVaultSecret := NewVaultSecret("aws", "gcp", KVPath)
	if awsVaultSecret.engine != aws || awsVaultSecret.path != KVPath || awsVaultSecret.mount != "gcp" || dbVaultSecret.dynamic != true {
		test.Error("the AWS Vault secret constructor returned unexpected values")
		test.Errorf("expected engine: %s, actual: %s", awsVaultSecret.engine, aws)
		test.Errorf("expected path: %s, actual: %s", awsVaultSecret.path, KVPath)
		test.Errorf("expected mount: gcp, actual: %s", awsVaultSecret.mount)
		test.Errorf("expected dynamic to be true, actual: %t", dbVaultSecret.dynamic)
	}
}

// test populate secret
func TestPopulateKVSecret(test *testing.T) {
	basicVaultClient := basicVaultConfig.AuthClient()

	kv1VaultSecret := NewVaultSecret("kv1", "", KVPath)
	version, secretMetadata, err := kv1VaultSecret.PopulateKVSecret(
		basicVaultClient,
		map[string]interface{}{KVKey: KVValue},
		false,
	)
	if err != nil {
		test.Error("the kv1 secret was not successfully put")
		test.Error(err)
	}
	if secretMetadata == (Metadata{}) {
		test.Error("the kv2 secret retrieval returned empty metadata")
	}
	if version != "0" {
		test.Errorf("the kv1 secret put returned non-zero version: %s", version)
	}

	kv2VaultSecret := NewVaultSecret("kv2", "", KVPath)
	version, secretMetadata, err = kv2VaultSecret.PopulateKVSecret(
		basicVaultClient,
		map[string]interface{}{KVKey: KVValue},
		false,
	)
	if err != nil {
		test.Error("the kv2 secret was not successfully put")
		test.Error(err)
	}
	if secretMetadata == (Metadata{}) {
		test.Error("the kv2 secret put returned empty metadata")
	}
	if version == "0" {
		test.Errorf("the kv2 secret put returned an invalid version: %s", version)
	}
	version, secretMetadata, err = kv2VaultSecret.PopulateKVSecret(
		basicVaultClient,
		map[string]interface{}{"other_password": "ultrasecret"},
		true,
	)
	if err != nil {
		test.Error("the kv2 secret was not successfully patched")
		test.Error(err)
	}
	if secretMetadata == (Metadata{}) {
		test.Error("the kv2 secret patch returned empty metadata")
	}
	if version == "0" {
		test.Errorf("the kv2 secret patch returned an invalid version: %s", version)
	}
}

// test secret renew
func TestRenew(test *testing.T) {}
