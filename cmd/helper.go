package helper

import (
	"encoding/json"
	"log"
	"os"

	"github.com/mschuchard/concourse-vault-resource/concourse"
	"github.com/mschuchard/concourse-vault-resource/vault"
)

// writes inResponse.Metadata marshalled to json to file at /opt/resource/vault.json
func SecretsToJsonFile(filePath string, secretValues concourse.SecretValues) error {
	// marshal secretValues into json data
	secretsData, err := json.Marshal(secretValues)
	if err != nil {
		log.Print("unable to marshal SecretValues struct to json data")
		return err
	}
	// write secrets to file at /opt/resource/vault.json
	secretsFile := filePath + "/vault.json"
	if err = os.WriteFile(secretsFile, secretsData, 0o600); err != nil {
		log.Printf("error writing secrets to destination file at %s", secretsFile)
		return err
	}

	return nil
}

// converts Vault secret metadata information to Concourse metadata
func VaultToConcourseMetadata(prefix string, secretMetadata vault.Metadata) []concourse.MetadataEntry {
	// initialize metadata entries for raw secret
	metadataEntries := make([]concourse.MetadataEntry, 3)

	// assign lease id, lease duration, and renewable converted to string to the entries
	metadataEntries[0] = concourse.MetadataEntry{
		Name:  prefix + "-LeaseID",
		Value: secretMetadata.LeaseID,
	}
	metadataEntries[1] = concourse.MetadataEntry{
		Name:  prefix + "-LeaseDuration",
		Value: secretMetadata.LeaseDuration,
	}
	metadataEntries[2] = concourse.MetadataEntry{
		Name:  prefix + "-Renewable",
		Value: secretMetadata.Renewable,
	}

	return metadataEntries
}
