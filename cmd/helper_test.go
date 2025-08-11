package helper

import (
	"os"
	"slices"
	"testing"
	"time"

	"github.com/mschuchard/concourse-vault-resource/concourse"
	"github.com/mschuchard/concourse-vault-resource/vault"
)

// minimum coverage testing for helper functions
func TestSecretsToJsonFile(test *testing.T) {
	secretValues := concourse.SecretValues{"secretValue": {"key": "value"}}
	if err := SecretsToJsonFile(".", secretValues); err != nil {
		test.Error(err)
	}
	defer os.Remove("./vault.json")
}

func TestVaultToConcourseMetadata(test *testing.T) {
	duration, _ := time.ParseDuration("65535s")
	secretMetadata := vault.Metadata{
		LeaseID:       "abcdefg12345",
		LeaseDuration: duration,
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
			Value: secretMetadata.LeaseDuration.String(),
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
