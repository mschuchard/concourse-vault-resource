package vault

import (
	"context"
	"log"
	"strconv"
	"time"

	vault "github.com/hashicorp/vault/api"
)

// secret engine with pseudo-enum
type secretEngine string

const (
	database  secretEngine = "database"
	aws       secretEngine = "aws"
	keyvalue1 secretEngine = "kv1"
	keyvalue2 secretEngine = "kv2"
)

// secret metadata
type Metadata struct {
	LeaseID       string
	LeaseDuration string
	Renewable     string
}

// secret defines a composite Vault secret configuration
type vaultSecret struct {
	engine  secretEngine
	mount   string
	path    string
	dynamic bool
}

// secret constructor
func NewVaultSecret(engineString string, mount string, path string) *vaultSecret {
	// validate mandatory fields specified
	if len(engineString) == 0 || len(path) == 0 {
		log.Fatal("the secret engine and path parameters are mandatory")
	}

	// validate engine parameter
	engine := secretEngine(engineString)
	if len(engine) == 0 {
		log.Fatalf("an invalid secrets engine was specified: %s", engineString)
	}

	// initialize vault secret
	vaultSecret := &vaultSecret{
		engine: engine,
		path:   path,
		mount:  mount,
	}

	// determine if secret is dynamic (currently unused)
	switch engine {
	case database, aws:
		vaultSecret.dynamic = true
	case keyvalue1, keyvalue2:
		vaultSecret.dynamic = false
	default:
		log.Fatalf("an invalid secret engine %s was selected", engine)
	}

	// determine default mount path if not specified
	// note current schema renders this pointless, but it would ensure safety to retain
	if len(mount) == 0 {
		switch engine {
		case database:
			vaultSecret.mount = "database"
		case aws:
			vaultSecret.mount = "aws"
		case keyvalue1:
			vaultSecret.mount = "kv"
		case keyvalue2:
			vaultSecret.mount = "secret"
		default:
			log.Fatalf("an invalid secret engine %s was selected", engine)
		}
	}

	return vaultSecret
}

// secret readers
func (secret *vaultSecret) Engine() secretEngine {
	return secret.engine
}

func (secret *vaultSecret) Mount() string {
	return secret.mount
}

func (secret *vaultSecret) Path() string {
	return secret.path
}

func (secret *vaultSecret) Dynamic() bool {
	return secret.dynamic
}

// return secret value, version, metadata, and possible error (GET/READ/READ)
func (secret *vaultSecret) SecretValue(client *vault.Client, version string) (map[string]interface{}, string, Metadata, error) {
	switch secret.engine {
	case database, aws:
		return secret.generateCredentials(client)
	case keyvalue1, keyvalue2:
		return secret.retrieveKVSecret(client, version)
	default:
		log.Fatalf("an invalid secret engine %s was selected", secret.engine)
		return map[string]interface{}{}, "0", Metadata{}, nil // unreachable code, but compile error otherwise
	}
}

// populate key-value pair secrets and return version, metadata, and error (PUT+POST/UPDATE+CREATE/PATCH+WRITE)
func (secret *vaultSecret) PopulateKVSecret(client *vault.Client, secretValue map[string]interface{}, patch bool) (string, Metadata, error) {
	// declare error for return to cmd, and kvSecret for metadata.version and raw secret assignments and returns (with dummies for kv1)
	var err error
	kvSecret := &vault.KVSecret{
		VersionMetadata: &vault.KVVersionMetadata{Version: 0},
		Raw:             &vault.Secret{},
	}

	switch secret.engine {
	case keyvalue1:
		// put kv1 secret
		err = client.KVv1(secret.mount).Put(
			context.Background(),
			secret.path,
			secretValue,
		)
	case keyvalue2:
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
	default:
		log.Fatalf("an invalid secret engine %s was selected", secret.engine)
	}

	// verify secret put
	if err != nil {
		log.Printf("failed to update secret %s into %s secrets Engine", secret.path, secret.engine)
		log.Print(err)
		return "0", Metadata{}, err
	}

	// return no error
	return strconv.Itoa(kvSecret.VersionMetadata.Version), rawSecretToMetadata(kvSecret.Raw), nil
}

// renew dynamic secret lease and return updated metadata
func (secret *vaultSecret) Renew(client *vault.Client, leaseIdSuffix string) (string, Metadata, error) {
	// semi-validate secret is renewable (better but not possible is *Secret.Renewable)
	if !secret.dynamic {
		log.Printf("the input secret with engine %s at mount %s and path %s is not renewable", secret.engine, secret.mount, secret.path)
		return "0", Metadata{}, nil
	}

	// determine full lease id TODO expand once more engines supported
	leaseId := secret.mount + "/creds/" + leaseIdSuffix

	// renew the secret lease
	rawSecret, err := client.Sys().Renew(leaseId, 0)
	if err != nil {
		log.Printf("the secret with lease ID %s could not be renewed", leaseId)
		log.Print(err)
		return "0", Metadata{}, err
	}

	// calculate the expiration time for version
	expirationTime := time.Now().Local().Add(time.Second * time.Duration(rawSecret.LeaseDuration))

	// convert raw secret to metadata and return metadata and version
	return expirationTime.String(), rawSecretToMetadata(rawSecret), nil
}
