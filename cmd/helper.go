package helper

import (
	"encoding/json"
	"log"
	"os"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/mschuchard/concourse-vault-resource/concourse"
	"github.com/mschuchard/concourse-vault-resource/vault"
)

// instantiates vault client from concourse source
func VaultClientFromSource(source concourse.Source) (*vaultapi.Client, error) {
	// initialize vault config and client
	vaultConfig := &vault.VaultConfig{
		Engine:       vault.AuthEngine(source.AuthEngine),
		Address:      source.Address,
		AWSMountPath: source.AWSMountPath,
		AWSRole:      source.AWSVaultRole,
		Token:        source.Token,
		Insecure:     source.Insecure,
	}
	if err := vaultConfig.New(); err != nil {
		log.Print("error initializing Vault client config from Concourse source")
		return nil, err
	}

	// transparently return client or error up the stack
	return vaultConfig.AuthClient()
}

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
	var metadataEntries []concourse.MetadataEntry

	// append lease id, lease duration, and renewable converted to string to the entries
	metadataEntries = append(metadataEntries, concourse.MetadataEntry{
		Name:  prefix + "-LeaseID",
		Value: secretMetadata.LeaseID,
	})
	metadataEntries = append(metadataEntries, concourse.MetadataEntry{
		Name:  prefix + "-LeaseDuration",
		Value: secretMetadata.LeaseDuration,
	})
	metadataEntries = append(metadataEntries, concourse.MetadataEntry{
		Name:  prefix + "-Renewable",
		Value: secretMetadata.Renewable,
	})

	return metadataEntries
}
