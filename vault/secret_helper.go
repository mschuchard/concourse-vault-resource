package vault

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	vault "github.com/hashicorp/vault/api"
)

// generate credentials
func (secret *vaultSecret) generateCredentials(client *vault.Client) (map[string]interface{}, Metadata, error) {
	// initialize api endpoint for cred generation
	endpoint := secret.mount + "/creds/" + secret.path

	// GET the secret from the API endpoint
	response, err := client.Logical().Read(endpoint)
	if err != nil {
		log.Printf("failed to generate credentials for %s with %s secrets engine", secret.path, secret.engine)
		log.Print(err)
		return map[string]interface{}{}, Metadata{}, err
	}

	// calculate the expiration time for version
	expirationTime := time.Now().Local().Add(time.Second * time.Duration(response.LeaseDuration))

	// initialize secret metadata and assign version
	metadata := rawSecretToMetadata(response)
	metadata.Version = expirationTime.String()

	// return secret value implicitly coerced to map[string]interface{}, expiration time as version, and metadata
	return response.Data, metadata, nil
}

// retrieve key-value pair secrets
func (secret *vaultSecret) retrieveKVSecret(client *vault.Client, version string) (map[string]interface{}, Metadata, error) {
	// declare error for return to cmd, and kvSecret for metadata.version and raw secret assignments and returns
	var err error
	var kvSecret *vault.KVSecret

	switch secret.engine {
	case keyvalue1:
		if len(version) > 0 {
			log.Print("versions cannot be used with the KV1 secrets engine")
		}
		// read kv secret
		kvSecret, err = client.KVv1(secret.mount).Get(
			context.Background(),
			secret.path,
		)
		// instantiate dummy metadata if secret successfully retrieved
		if err == nil && kvSecret != nil {
			kvSecret.VersionMetadata = &vault.KVVersionMetadata{Version: 0}
		}
	case keyvalue2:
		// read latest kv2 secret
		if len(version) == 0 {
			kvSecret, err = client.KVv2(secret.mount).Get(
				context.Background(),
				secret.path,
			)
		} else { // read specific version of kv2 secret
			// validate version if input
			versionInt, err := strconv.Atoi(version)
			if err != nil {
				log.Printf("KV2 version must be an integer, and %s was input instead", version)
				// return empty values since error triggers at end of execution
				return map[string]interface{}{}, Metadata{}, err
			}

			kvSecret, err = client.KVv2(secret.mount).GetVersion(
				context.Background(),
				secret.path,
				versionInt,
			)
			if err != nil {
				log.Printf("the KV2 secret at %s/%s could not be retrieved for version %d", secret.mount, secret.path, versionInt)
				// return empty values since error triggers at end of execution
				return map[string]interface{}{}, Metadata{}, err
			}
		}
	default:
		log.Printf("an invalid secret engine %s was selected", secret.engine)
		return map[string]interface{}{}, Metadata{}, errors.New("invalid secret engine")
	}

	// verify secret read
	if err != nil || kvSecret == nil {
		log.Printf("failed to read secret at mount %s and path %s from %s secrets engine", secret.mount, secret.path, secret.engine)
		log.Print(err)
		// return empty values since error triggers at end of execution
		return map[string]interface{}{}, Metadata{}, err
	}

	// initialize secret metadata
	metadata := rawSecretToMetadata(kvSecret.Raw)

	if kvSecret.Data == nil { // verify version exists
		log.Printf("the input version %s (0 signifies latest) does not exist for the secret at mount %s and path %s from %s secrets engine", version, secret.mount, secret.path, secret.engine)

		// return partial information values since error triggers at end of execution
		metadata.Version = version
		return map[string]interface{}{}, metadata, err
	}

	// return secret value and implicitly coerce type to map[string]interface{}
	metadata.Version = strconv.Itoa(kvSecret.VersionMetadata.Version)
	return kvSecret.Data, metadata, nil
}

// populate key-value v1 pair secrets
func (secret *vaultSecret) populateKV1Secret(client *vault.Client, secretValue map[string]interface{}) (Metadata, error) {
	// put kv1 secret
	err := client.KVv1(secret.mount).Put(
		context.Background(),
		secret.path,
		secretValue,
	)
	// verify secret put
	if err != nil {
		log.Printf("failed to update secret %s into %s secrets Engine", secret.path, secret.engine)
		log.Print(err)
		return Metadata{}, err
	}

	// initialize secret metadata and assign dummy version
	metadata := rawSecretToMetadata(&vault.Secret{})
	metadata.Version = "0"

	return metadata, nil
}

// populate key-value v2 pair secrets
func (secret *vaultSecret) populateKV2Secret(client *vault.Client, secretValue map[string]interface{}, patch bool) (Metadata, error) {
	// declare error and kvSecret for return to cmd
	var err error
	var kvSecret *vault.KVSecret

	if patch {
		// patch kv2 secret
		kvSecret, err = client.KVv2(secret.mount).Patch(
			context.Background(),
			secret.path,
			secretValue,
		)
	} else {
		// put kv2 secret
		kvSecret, err = client.KVv2(secret.mount).Put(
			context.Background(),
			secret.path,
			secretValue,
		)
	}

	// verify secret patch/put
	if err != nil {
		log.Printf("failed to update secret %s into %s secrets Engine", secret.path, secret.engine)
		log.Print(err)
		return Metadata{}, err
	}

	// initialize secret metadata and assign version
	metadata := rawSecretToMetadata(kvSecret.Raw)
	metadata.Version = strconv.Itoa(kvSecret.VersionMetadata.Version)

	// return no error
	return metadata, nil
}

// convert *vault.Secret raw secret to secret metadata
func rawSecretToMetadata(rawSecret *vault.Secret) Metadata {
	// returne metadata with fields populated from raw secret
	return Metadata{
		LeaseID:       rawSecret.LeaseID,
		LeaseDuration: strconv.Itoa(rawSecret.LeaseDuration),
		Renewable:     strconv.FormatBool(rawSecret.Renewable),
		Version:       "0", // default value to be overwritten later
	}
}
