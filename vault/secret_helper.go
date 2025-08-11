package vault

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/mschuchard/concourse-vault-resource/enum"
)

// secret metadata
type Metadata struct {
	LeaseID       string
	LeaseDuration time.Duration
	Renewable     bool
	Version       string
}

// generate credentials
func (secret *vaultSecret) generateCredentials(client *vault.Client) (map[string]any, Metadata, error) {
	var rawSecret *vault.Secret
	var err error

	// generate credentials based on secret engine type
	if secret.engine == enum.SSH {
		rawSecret, err = client.SSHWithMountPoint(secret.mount).Credential(secret.path, map[string]any{})
	} else {
		rawSecret, err = client.Logical().Read(secret.mount + "/creds/" + secret.path)
	}
	if err != nil {
		log.Printf("failed to generate credentials for %s with %s secrets engine", secret.path, secret.engine)
		log.Print(err)
		return map[string]any{}, Metadata{}, err
	}

	// initialize secret metadata
	metadata, err := rawSecretToMetadata(rawSecret)
	if err != nil {
		log.Print("raw secret could not be converted to metadata")
		return map[string]any{}, Metadata{}, err
	}

	// calculate the expiration time for version and assign to metadata
	expirationTime := time.Now().Local().Add(time.Second * time.Duration(rawSecret.LeaseDuration))
	metadata.Version = expirationTime.Format("2006-01-02-150405")

	// return secret value implicitly coerced to map[string]any, expiration time as version, and metadata
	return rawSecret.Data, metadata, nil
}

// retrieve key-value pair secrets
func (secret *vaultSecret) retrieveKVSecret(client *vault.Client, version string) (map[string]any, Metadata, error) {
	// declare error for return to cmd, and kvSecret for metadata.version and raw secret assignments and returns
	var err error
	var kvSecret *vault.KVSecret

	switch secret.engine {
	case enum.KeyValue1:
		if len(version) > 0 {
			log.Print("versions cannot be used with the KV1 secrets engine, and the input parameter will be ignored")
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
	case enum.KeyValue2:
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
				return map[string]any{}, Metadata{}, err
			}

			kvSecret, err = client.KVv2(secret.mount).GetVersion(
				context.Background(),
				secret.path,
				versionInt,
			)
			if err != nil {
				log.Printf("the KV2 secret at %s/%s could not be retrieved for version %d", secret.mount, secret.path, versionInt)
				// return empty values since error triggers at end of execution
				return map[string]any{}, Metadata{}, err
			}
		}
	default:
		log.Printf("an invalid secret engine %s was selected", secret.engine)
		return map[string]any{}, Metadata{}, errors.New("invalid secret engine")
	}

	// verify secret read (err from latest version)
	if err != nil || kvSecret == nil {
		log.Printf("failed to read secret at mount %s and path %s from %s secrets engine", secret.mount, secret.path, secret.engine)
		log.Print(err)
		// return empty values since error triggers at end of execution
		return map[string]any{}, Metadata{}, err
	}

	// initialize secret metadata
	metadata, err := rawSecretToMetadata(kvSecret.Raw)
	if err != nil {
		log.Print("raw secret could not be converted to metadata")
		return map[string]any{}, Metadata{}, err
	}

	if kvSecret.Data == nil { // verify version exists
		log.Printf("the input version %s (0 signifies latest) does not exist for the secret at mount %s and path %s from %s secrets engine", version, secret.mount, secret.path, secret.engine)

		// return partial information values since error triggers at end of execution
		metadata.Version = version
		return map[string]any{}, metadata, err
	}

	// return secret value and implicitly coerce type to map[string]any
	metadata.Version = strconv.Itoa(kvSecret.VersionMetadata.Version)
	return kvSecret.Data, metadata, nil
}

// populate key-value v1 pair secrets
func (secret *vaultSecret) populateKV1Secret(client *vault.Client, secretValue map[string]any) (Metadata, error) {
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
	metadata, err := rawSecretToMetadata(&vault.Secret{})
	if err != nil {
		log.Print("raw secret could not be converted to metadata")
		return Metadata{}, err
	}
	metadata.Version = "0"

	return metadata, nil
}

// populate key-value v2 pair secrets
func (secret *vaultSecret) populateKV2Secret(client *vault.Client, secretValue map[string]any, patch bool) (Metadata, error) {
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
	metadata, err := rawSecretToMetadata(kvSecret.Raw)
	if err != nil {
		log.Print("raw secret could not be converted to metadata")
		return Metadata{}, err
	}
	metadata.Version = strconv.Itoa(kvSecret.VersionMetadata.Version)

	// return no error
	return metadata, nil
}

// convert *vault.Secret raw secret to secret metadata // TODO: seems like I convert raw secret lease duration to time.Duration multiple places
func rawSecretToMetadata(rawSecret *vault.Secret) (Metadata, error) {
	if rawSecret == nil {
		log.Print("the raw secret is nil, and metadata cannot be constructed from it")
		return Metadata{}, errors.New("nil raw secret")
	}

	// return metadata with fields populated from raw secret
	return Metadata{
		LeaseID:       rawSecret.LeaseID,
		LeaseDuration: time.Second * time.Duration(rawSecret.LeaseDuration),
		Renewable:     rawSecret.Renewable,
		Version:       "0", // default value to be overwritten later
	}, nil
}
