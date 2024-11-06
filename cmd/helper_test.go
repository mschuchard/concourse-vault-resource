package helper

import (
	"os"
	"slices"
	"testing"

	"github.com/mschuchard/concourse-vault-resource/concourse"
	"github.com/mschuchard/concourse-vault-resource/vault"
	"github.com/mschuchard/concourse-vault-resource/vault/util"
)

// minimum coverage testing for helper functions
func TestVaultClientFromSource(test *testing.T) {
	source := concourse.Source{Token: util.VaultToken}
	client, err := VaultClientFromSource(source)
	if err != nil {
		test.Error("failed to initialize vault client from concourse source")
		test.Error(err)
	} else {
		if client.Token() != util.VaultToken {
			test.Error("vault client configured with parameters from concourse source was not authenticated with expected token from source parameters")
			test.Errorf("actual token: %s, expected token: %s", client.Token(), util.VaultToken)
		}
	}
}

func TestSecretsToJsonFile(test *testing.T) {
	secretValues := concourse.SecretValues{"secretValue": {"key": "value"}}
	if err := SecretsToJsonFile(".", secretValues); err != nil {
		test.Error(err)
	}
	defer os.Remove("./vault.json")
}

func TestVaultToConcourseMetadata(test *testing.T) {
	secretMetadata := vault.Metadata{
		LeaseID:       "abcdefg12345",
		LeaseDuration: "65535",
		Renewable:     "false",
	}
	secretPath := "secret-foo/bar"

	concourseMetadata := VaultToConcourseMetadata(secretPath, secretMetadata)
	expectedConcourseMetadata := []concourse.MetadataEntry{
		{
			Name:  secretPath + "-LeaseID",
			Value: secretMetadata.LeaseID,
		},
		{
			Name:  secretPath + "-LeaseDuration",
			Value: secretMetadata.LeaseDuration,
		},
		{
			Name:  secretPath + "-Renewable",
			Value: secretMetadata.Renewable,
		},
	}

	if !slices.Equal(expectedConcourseMetadata, concourseMetadata) {
		test.Error("vault to concourse metadata conversion returned unexpected value")
		test.Errorf("expected value: %v", expectedConcourseMetadata)
		test.Errorf("actual value: %v", concourseMetadata)
	}
}
