package helper

import (
	"encoding/json"
	"log"
	"os"

	"github.com/mschuchard/concourse-vault-resource/concourse"
	"github.com/mschuchard/concourse-vault-resource/vault"
)

// writes inResponse.Metadata marshalled to json to file at /opt/resource/vault.json
func SecretsToJSONFile(filePath string, secretValues concourse.SecretValues) error {
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
	// return vault metadata lease id, lease duration, and renewable as concourse metadata entries
	return []concourse.MetadataEntry{
		{
			Name:  prefix + "-LeaseID",
			Value: secretMetadata.LeaseID,
		},
		{
			Name:  prefix + "-LeaseDuration",
			Value: secretMetadata.LeaseDuration,
		},
		{
			Name:  prefix + "-Renewable",
			Value: secretMetadata.Renewable,
		},
	}
}
