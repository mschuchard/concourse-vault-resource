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

// PUT/POST
func main() {
	// initialize request from concourse pipeline and response to satisfy concourse requirement
	outRequest, err := concourse.NewOutRequest(os.Stdin)
	if err != nil {
		log.Print("unable to construct request for out/put step")
		log.Fatal(err)
	}
	outResponse := concourse.NewResponse()
	// initialize vault client from concourse source
	vaultClient, err := helper.VaultClientFromSource(outRequest.Source)
	if err != nil {
		log.Print("vault client failed to initialize during out/put")
		log.Fatal(err)
	}

	// perform secrets operations
	for mount, secretParams := range outRequest.Params {
		// iterate through secrets and assign each path to each vault secret path, and write each secret value to the path
		for secretPath, secretValue := range secretParams.Secrets {
			// declare because implicit type deduction not allowed
			var secretMetadata vault.Metadata
			// initialize vault secret from concourse params
			secret, nestedErr := vault.NewVaultSecret(secretParams.Engine, mount, secretPath)
			if nestedErr != nil {
				log.Print("failed to construct secret from Concourse parameters")
				log.Fatal(nestedErr)
			}
			// declare identifier and rawSecret
			identifier := mount + "-" + secretPath
			// write the secret value to the path for the specified mount and engine
			outResponse.Version[identifier], secretMetadata, nestedErr = secret.PopulateKVSecret(vaultClient, secretValue, secretParams.Patch)
			// join error into collection
			err = errors.Join(err, nestedErr)
			// convert rawSecret to concourse metadata and append to metadata
			outResponse.Metadata = append(outResponse.Metadata, helper.VaultToConcourseMetadata(identifier, secretMetadata)...)
		}
	}

	// fatally exit if any secret Read operation failed
	if err != nil {
		log.Print("one or more attempted secret Create/Update operations failed")
		log.Fatal(err)
	}

	// format outResponse into json
	if err = json.NewEncoder(os.Stdout).Encode(outResponse); err != nil {
		log.Print("unable to marshal out response struct to JSON")
		log.Fatal(err)
	}
}
