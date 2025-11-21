package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/mschuchard/concourse-vault-resource/concourse"
	"github.com/mschuchard/concourse-vault-resource/vault"
)

// GET for kv2 and credentials (expiration time) versions (kv1 not possible)
func main() {
	// initialize checkRequest and secretSource
	checkRequest, err := concourse.NewCheckRequest(os.Stdin)
	if err != nil {
		log.Print("unable to construct request for check step")
		log.Fatal(err)
	}
	secretSource := checkRequest.Source.Secret

	// return immediately if secret unspecified in source or is kv1
	if secretSource == (concourse.SecretSource{}) || secretSource.Engine == "kv1" {
		// dummy check response
		dummyResponse := concourse.NewCheckResponse([]concourse.Version{{Version: "0"}})
		// format checkResponse into json
		if err := json.NewEncoder(os.Stdout).Encode(&dummyResponse); err != nil {
			log.Print("unable to marshal dummy check response struct to JSON")
			log.Fatal(err)
		}

		log.Print("source does not contain a secret, or a secret with kv version 1 engine")
		log.Print("concourse version will be set to value '0'")

		return
	}

	// initialize vault client from concourse source
	vaultClient, err := vault.NewVaultClient(checkRequest.Source)
	if err != nil {
		log.Print("vault client failed to initialize during check")
		log.Fatal(err)
	}

	// initialize vault secret from concourse source params and invoke constructor
	secret, err := vault.NewVaultSecret(secretSource.Engine, secretSource.Mount, secretSource.Path)
	if err != nil {
		log.Print("failed to construct secret from Concourse source parameters")
		log.Fatal(err)
	}

	// retrieve version for secret
	_, secretMetadata, err := secret.SecretValue(vaultClient, "")
	if err != nil {
		log.Printf("version could not be retrieved for %s engine, %s mount, and path %s secret", secretSource.Engine, secretSource.Mount, secretSource.Path)
		log.Fatal(err)
	}

	// assign input and get version and initialize versions slice
	inputVersion, err := strconv.Atoi(checkRequest.Version.Version)
	if err != nil {
		log.Printf("the input version '%s' in source is not a valid integer", checkRequest.Version.Version)
		log.Fatal(err)
	}
	versions := []concourse.Version{}
	getVersionInt, err := strconv.Atoi(secretMetadata.Version)

	// if getVersion could not be converted to int then this may be a dynamically generated credential
	if err != nil {
		if secret.Dynamic() {
			// this is a dynamically generated credential so renew it
			log.Printf("the secret '%s' is dynamic and will be renewed", secretSource.Path)

			secretMetadata, err = secret.Renew(vaultClient, secretSource.LeaseId)
			if err != nil {
				log.Printf("failed to renew dynamic secret for %s engine, %s mount, and path %s", secretSource.Engine, secretSource.Mount, secretSource.Path)
				log.Fatal(err)
			}
		}

		// assign versions through returned metadata re-assignment during renewal
		// OR dummy a return for the versions using the original metadata return
		versions = []concourse.Version{{Version: secretMetadata.Version}}
	} else {
		// validate that the input version is <= the latest retrieved version
		if inputVersion > getVersionInt {
			log.Printf("the input version %d is later than the retrieved version %s", inputVersion, secretMetadata.Version)
			log.Print("only the retrieved version will be returned to Concourse")

			versions = []concourse.Version{{Version: secretMetadata.Version}}
		} else {
			// populate versions slice with delta
			for versionDelta := inputVersion; versionDelta <= getVersionInt; versionDelta++ {
				versions = append(versions, concourse.Version{Version: strconv.Itoa(versionDelta)})
			}
		}
	}

	// input secret version to constructed response
	checkResponse := concourse.NewCheckResponse(versions)

	// format checkResponse into json
	if err := json.NewEncoder(os.Stdout).Encode(&checkResponse); err != nil {
		log.Print("unable to marshal check response struct to JSON")
		log.Fatal(err)
	}
}
