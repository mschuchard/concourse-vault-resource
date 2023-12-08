package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/mitodl/concourse-vault-resource/cmd"
	"github.com/mitodl/concourse-vault-resource/concourse"
	"github.com/mitodl/concourse-vault-resource/vault"
)

// PUT/POST
func main() {
	// initialize request from concourse pipeline and response to satisfy concourse requirement
	outRequest := concourse.NewOutRequest(os.Stdin)
	outResponse := concourse.NewResponse()
	// initialize vault client from concourse source
	vaultClient := helper.VaultClientFromSource(outRequest.Source)

	// declare err specifically to track any SecretValue failure and trigger only after all secret operations
	var err error
	var secretMetadata vault.Metadata

	// perform secrets operations
	for mount, secretParams := range outRequest.Params {
		// iterate through secrets and assign each path to each vault secret path, and write each secret value to the path
		for secretPath, secretValue := range secretParams.Secrets {
			// initialize vault secret from concourse params
			secret := vault.NewVaultSecret(secretParams.Engine, mount, secretPath)
			// declare identifier and rawSecret
			identifier := mount + "-" + secretPath
			// write the secret value to the path for the specified mount and engine
			outResponse.Version[identifier], secretMetadata, err = secret.PopulateKVSecret(vaultClient, secretValue, secretParams.Patch)
			// convert rawSecret to concourse metadata and append to metadata
			outResponse.Metadata = append(outResponse.Metadata, helper.VaultToConcourseMetadata(identifier, secretMetadata)...)
		}
	}

	// fatally exit if any secret Read operation failed
	if err != nil {
		log.Fatal("one or more attempted secret Create/Update operations failed")
	}

	// format outResponse into json
	if err = json.NewEncoder(os.Stdout).Encode(outResponse); err != nil {
		log.Print("unable to marshal out response struct to JSON")
		log.Fatal(err)
	}
}
