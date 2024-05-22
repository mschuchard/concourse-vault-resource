package vault

import (
	"testing"

	vault "github.com/hashicorp/vault/api"

	"github.com/mschuchard/concourse-vault-resource/vault/util"
)

// test secret generate credential
func TestGenerateCredentials(test *testing.T) {}

// test secret key value secret
func TestRetrieveKVSecret(test *testing.T) {
	kv1VaultSecret, err := NewVaultSecret("kv1", "", util.KVPath)
	if err != nil {
		test.Error("kv secret failed to construct")
		test.Error(err)
	}

	kv1Value, version, secretMetadata, err := kv1VaultSecret.retrieveKVSecret(util.VaultClient, "")

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
	if kv1Value[util.KVKey] != util.KVValue {
		test.Error("the retrieved kv1 secret value was incorrect")
		test.Errorf("secret map value: %v", kv1Value)
	}

	kv2VaultSecret, err := NewVaultSecret("kv2", util.KV2Mount, util.KVPath)
	if err != nil {
		test.Error("kv secret failed to construct")
		test.Error(err)
	}

	kv2Value, version, secretMetadata, err := kv2VaultSecret.retrieveKVSecret(util.VaultClient, "")

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
	if kv2Value[util.KVKey] != util.KVValue {
		test.Error("the retrieved kv2 secret value was incorrect")
		test.Errorf("secret map value: %v", kv2Value)
	}
}

// test populate kv1 secret
func TestPopulateKV1Secret(test *testing.T) {
	kv1VaultSecret, err := NewVaultSecret("kv1", "", util.KVPath)
	if err != nil {
		test.Error("kv secret failed to construct")
		test.Error(err)
	}

	version, secretMetadata, err := kv1VaultSecret.populateKV1Secret(
		util.VaultClient,
		map[string]interface{}{util.KVKey: util.KVValue},
	)
	if err != nil {
		test.Error("the kv1 secret was not successfully put")
		test.Error(err)
	}
	if secretMetadata == (Metadata{}) {
		test.Error("the kv1 secret retrieval returned empty metadata")
	}
	if version != "0" {
		test.Errorf("the kv1 secret put returned non-zero version: %s", version)
	}
}

// test populate kv2 secret
func TestPopulateKV2Secret(test *testing.T) {
	kv2VaultSecret, err := NewVaultSecret("kv2", "", util.KVPath)
	if err != nil {
		test.Error("kv secret failed to construct")
		test.Error(err)
	}

	version, secretMetadata, err := kv2VaultSecret.populateKV2Secret(
		util.VaultClient,
		map[string]interface{}{util.KVKey: util.KVValue},
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
	version, secretMetadata, err = kv2VaultSecret.populateKV2Secret(
		util.VaultClient,
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
	expectedMetadata := Metadata{LeaseID: rawSecret.LeaseID, LeaseDuration: "65535", Renewable: "false"}

	if metadata != expectedMetadata {
		test.Error("the converted metadata returned unexpected values")
		test.Errorf("expected values: %v", expectedMetadata)
		test.Errorf("actual values: %v", metadata)
	}
}
