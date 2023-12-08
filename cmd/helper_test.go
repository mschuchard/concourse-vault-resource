package helper

import (
	"os"
	"testing"

	"github.com/mitodl/concourse-vault-resource/concourse"
	"github.com/mitodl/concourse-vault-resource/vault"
)

// minimum coverage testing for helper functions
func TestVaultClientFromSource(test *testing.T) {
	source := concourse.Source{Token: "abcdefghijklmnopqrstuvwxyz09"}
	VaultClientFromSource(source)
}

func TestSecretsToJsonFile(test *testing.T) {
	secretValues := concourse.SecretValues{}
	SecretsToJsonFile(".", secretValues)
	defer os.Remove("./vault.json")
}

func TestVaultToConcourseMetadata(test *testing.T) {
	secretMetadata := vault.Metadata{
		LeaseID:       "abcdefg12345",
		LeaseDuration: "65535",
		Renewable:     "false",
	}

	metadata := VaultToConcourseMetadata("secret-foo/bar", secretMetadata)
	if len(metadata) != 3 {
		test.Error("metadata did not contain the expected number (three) entries per raw secret")
	}
	if metadata[0].Name != "secret-foo/bar-LeaseID" || metadata[0].Value != secretMetadata.LeaseID {
		test.Error("first metadata entry is inaccurate")
		test.Errorf("expected name: secret-foo/bar-LeaseID, actual: %s", metadata[0].Name)
		test.Errorf("expected value: %s, actual: %s", secretMetadata.LeaseID, metadata[0].Value)
	}
	if metadata[1].Name != "secret-foo/bar-LeaseDuration" || metadata[1].Value != secretMetadata.LeaseDuration {
		test.Error("first metadata entry is inaccurate")
		test.Errorf("expected name: secret-foo/bar-LeaseDuration, actual: %s", metadata[1].Name)
		test.Errorf("expected value: %s, actual: %s", secretMetadata.LeaseDuration, metadata[1].Value)
	}
	if metadata[2].Name != "secret-foo/bar-Renewable" || metadata[2].Value != secretMetadata.Renewable {
		test.Error("first metadata entry is inaccurate")
		test.Errorf("expected name: secret-foo/bar-Renewable, actual: %s", metadata[2].Name)
		test.Errorf("expected value: %s, actual: %s", secretMetadata.Renewable, metadata[2].Value)
	}
}
