package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/mitodl/concourse-vault-resource/cmd"
	"github.com/mitodl/concourse-vault-resource/concourse"
	"github.com/mitodl/concourse-vault-resource/vault"
)

// GET for kv2 and credentials (expiration time) versions (kv1 not possible)
func main() {
	// initialize checkRequest and secretSource
	checkRequest := concourse.NewCheckRequest(os.Stdin)
	secretSource := checkRequest.Source.Secret

	// return immediately if secret unspecified in source TODO unblock dynamic later
	if secretSource == (concourse.SecretSource{}) || secretSource.Engine != "kv2" {
		// dummy check response
		dummyResponse := concourse.NewCheckResponse([]concourse.Version{concourse.Version{Version: "0"}})
		// format checkResponse into json
		if err := json.NewEncoder(os.Stdout).Encode(&dummyResponse); err != nil {
			log.Print("unable to marshal dummy check response struct to JSON")
			log.Fatal(err)
		}

		return
	}

	// initialize vault client from concourse source
	vaultClient := helper.VaultClientFromSource(checkRequest.Source)

	// initialize vault secret from concourse params and invoke constructor
	secret := vault.NewVaultSecret(secretSource.Engine, secretSource.Mount, secretSource.Path)

	// retrieve version for secret
	_, getVersion, _, err := secret.SecretValue(vaultClient, "")
	if err != nil {
		log.Fatalf("version could not be retrieved for %s engine, %s mount, and path %s secret", secretSource.Engine, secretSource.Mount, secretSource.Path)
	}

	// assign input and get version and initialize versions slice
	getVersionInt, err := strconv.Atoi(getVersion)
	inputVersion, _ := strconv.Atoi(checkRequest.Version.Version)
	versions := []concourse.Version{}

	// if getVersion could not be converted to int then just use the original string
	if err != nil {
		versions = []concourse.Version{concourse.Version{Version: getVersion}}
	} else {
		// validate that the input version is <= the latest retrieved version
		if inputVersion > getVersionInt {
			log.Printf("the input version %d is later than the retrieved version %s", inputVersion, getVersion)
			log.Print("only the retrieved version will be returned to Concourse")

			versions = []concourse.Version{concourse.Version{Version: getVersion}}
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
