package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/mitodl/concourse-vault-resource/cmd"
	"github.com/mitodl/concourse-vault-resource/concourse"
	"github.com/mitodl/concourse-vault-resource/vault"
)

// GET and primary
func main() {
	// initialize request from concourse pipeline and response storing secret values
	inRequest := concourse.NewInRequest(os.Stdin)
	inResponse := concourse.NewResponse()
	// initialize vault client from concourse source
	vaultClient := helper.VaultClientFromSource(inRequest.Source)

	// declare err specifically to track any SecretValue failure and trigger only after all secret operations
	var err error
	// initialize secretValues to store aggregated retrieved secrets and secretSource for efficiency
	var secretMetadata vault.Metadata
	secretValues := concourse.SecretValues{}
	secretSource := inRequest.Source.Secret

	// read secrets from params
	if secretSource == (concourse.SecretSource{}) {
		// perform secrets operations
		for mount, secretParams := range inRequest.Params {
			// iterate through secret params' paths and assign each to each vault secret path
			for _, secretPath := range secretParams.Paths {
				// initialize vault secret from concourse params
				secret := vault.NewVaultSecret(secretParams.Engine, mount, secretPath)
				// declare identifier and rawSecret
				identifier := mount + "-" + secretPath

				// renew or retrieve/generate
				if secretParams.Renew {
					// return updated metadata for dynamic secret after lease renewal
					inResponse.Version[identifier], secretMetadata, err = secret.Renew(vaultClient, secretPath)
				} else {
					// return and assign the secret values for the given path
					secretValues[identifier], inResponse.Version[identifier], secretMetadata, err = secret.SecretValue(vaultClient, "")
				}
				// convert rawSecret to concourse metadata and append to metadata
				inResponse.Metadata = append(inResponse.Metadata, helper.VaultToConcourseMetadata(identifier, secretMetadata)...)
			}
		}
	} else { // read secret from source
		// initialize vault secret from concourse params
		secret := vault.NewVaultSecret(secretSource.Engine, secretSource.Mount, secretSource.Path)
		// declare identifier and rawSecret
		identifier := secretSource.Mount + "-" + secretSource.Path
		// return and assign the secret values for the given path
		secretValues[identifier], inResponse.Version[identifier], secretMetadata, err = secret.SecretValue(vaultClient, inRequest.Version.Version)
		// convert rawSecret to concourse metadata and append to metadata
		inResponse.Metadata = append(inResponse.Metadata, helper.VaultToConcourseMetadata(identifier, secretMetadata)...)
	}

	// fatally exit if any secret Read operation failed TODO non nil can be overwritten by later nil
	if err != nil {
		log.Fatal("one or more attempted secret Read operations failed")
	}

	// write marshalled metadata to file at /opt/resource/vault.json
	helper.SecretsToJsonFile(os.Args[1], secretValues)

	// marshal, encode, and pass inResponse json as output to concourse
	if err = json.NewEncoder(os.Stdout).Encode(inResponse); err != nil {
		log.Print("unable to marshal in response struct to JSON")
		log.Fatal(err)
	}
}
