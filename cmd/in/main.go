package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/mschuchard/concourse-vault-resource/cmd"
	"github.com/mschuchard/concourse-vault-resource/concourse"
	"github.com/mschuchard/concourse-vault-resource/vault"
)

// GET and primary
func main() {
	// initialize request from concourse pipeline and response storing secret values
	inRequest, err := concourse.NewInRequest(os.Stdin)
	if err != nil {
		log.Print("unable to construct request for in/get step")
		log.Fatal(err)
	}
	inResponse := concourse.NewResponse()
	// initialize vault client from concourse source
	vaultClient, err := helper.VaultClientFromSource(inRequest.Source)
	if err != nil {
		log.Print("vault client failed to initialize during in/get")
		log.Fatal(err)
	}

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
				secret, nestedErr := vault.NewVaultSecret(secretParams.Engine, mount, secretPath)
				if nestedErr != nil {
					log.Print("failed to construct secret from Concourse parameters")
					log.Fatal(nestedErr)
				}
				// declare identifier
				identifier := mount + "-" + secretPath

				// renew or retrieve/generate
				if secretParams.Renew {
					// return updated metadata for dynamic secret after lease renewal
					inResponse.Version[identifier], secretMetadata, nestedErr = secret.Renew(vaultClient, secretPath)
				} else {
					// return and assign the secret values for the given path
					secretValues[identifier], inResponse.Version[identifier], secretMetadata, nestedErr = secret.SecretValue(vaultClient, "")
				}
				// join error into collection
				err = errors.Join(err, nestedErr)
				// convert rawSecret to concourse metadata and append to metadata
				inResponse.Metadata = append(inResponse.Metadata, helper.VaultToConcourseMetadata(identifier, secretMetadata)...)
			}
		}
	} else { // read secret from source
		// initialize vault secret from concourse source params
		secret, nestedErr := vault.NewVaultSecret(secretSource.Engine, secretSource.Mount, secretSource.Path)
		if nestedErr != nil {
			log.Print("failed to construct secret from Concourse source parameters")
			log.Fatal(nestedErr)
		}
		// declare identifier and rawSecret
		identifier := secretSource.Mount + "-" + secretSource.Path
		// return and assign the secret values for the given path
		secretValues[identifier], inResponse.Version[identifier], secretMetadata, nestedErr = secret.SecretValue(vaultClient, inRequest.Version.Version)
		// join error into collection
		err = errors.Join(err, nestedErr)
		// convert rawSecret to concourse metadata and append to metadata
		inResponse.Metadata = append(inResponse.Metadata, helper.VaultToConcourseMetadata(identifier, secretMetadata)...)
	}

	// fatally exit if any secret Read operation failed
	if err != nil {
		log.Print("one or more attempted secret Read operations failed")
		log.Fatal(err)
	}

	// write marshalled metadata to file at /opt/resource/vault.json
	err = helper.SecretsToJsonFile(os.Args[1], secretValues)
	if err != nil {
		log.Print("failed to output secrets in json format to file")
		log.Fatal(err)
	}

	// marshal, encode, and pass inResponse json as output to concourse
	if err = json.NewEncoder(os.Stdout).Encode(inResponse); err != nil {
		log.Print("unable to marshal in response struct to JSON")
		log.Fatal(err)
	}
}
