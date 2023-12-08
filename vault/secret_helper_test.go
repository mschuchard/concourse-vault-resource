package vault

import (
	"testing"

	vault "github.com/hashicorp/vault/api"
)

// test secret generate credential

// test secret key value secret
func TestRetrieveKVSecret(test *testing.T) {
	basicVaultClient := basicVaultConfig.AuthClient()

	kv1VaultSecret := NewVaultSecret("kv1", "", KVPath)
	kv1Value, version, secretMetadata, err := kv1VaultSecret.retrieveKVSecret(basicVaultClient, "")

	if err != nil {
		test.Error("kv1 secret retrieval failed")
		test.Error(err)
	}
	if secretMetadata == (Metadata{}) {
		test.Error("the kv2 secret retrieval returned empty metadata")
	}
	if version != "0" {
		test.Errorf("the kv1 secret retrieval returned non-zero version: %s", version)
	}
	if kv1Value[KVKey] != KVValue {
		test.Error("the retrieved kv1 secret value was incorrect")
		test.Errorf("secret map value: %v", kv1Value)
	}

	kv2VaultSecret := NewVaultSecret("kv2", KV2Mount, KVPath)
	kv2Value, version, secretMetadata, err := kv2VaultSecret.retrieveKVSecret(basicVaultClient, "")

	if err != nil {
		test.Error("kv2 secret retrieval failed")
		test.Error(err)
	}
	if secretMetadata == (Metadata{}) {
		test.Error("the kv2 secret retrieval returned empty metadata")
	}
	if version == "0" {
		test.Errorf("the kv2 secret retrieval returned an invalid version: %s", version)
	}
	if kv2Value[KVKey] != KVValue {
		test.Error("the retrieved kv2 secret value was incorrect")
		test.Errorf("secret map value: %v", kv2Value)
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

// test raw secret to metadata
func TestRawSecretToMetadata(test *testing.T) {
	rawSecret := &vault.Secret{
		LeaseID:       "abcdefg12345",
		LeaseDuration: 65535,
		Renewable:     false,
	}

	metadata := rawSecretToMetadata(rawSecret)

	if metadata.LeaseID != rawSecret.LeaseID || metadata.LeaseDuration != "65535" || metadata.Renewable != "false" {
		test.Error("the converted metadata returned unexpected values")
		test.Errorf("expected leaseid: %s, actual: %s", rawSecret.LeaseID, metadata.LeaseID)
		test.Errorf("expected leaseduration value and type: 65535 string, actual: %s %T", metadata.LeaseDuration, metadata.LeaseDuration)
		test.Errorf("expected renewable value and type: false string, actual: %s %T", metadata.Renewable, metadata.Renewable)
	}
}
