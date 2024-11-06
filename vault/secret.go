package vault

import (
	"errors"
	"log"
	"time"

	vault "github.com/hashicorp/vault/api"
)

// secret engine with pseudo-enum
type secretEngine string

const (
	// dynamic credential generators
	database   secretEngine = "database"
	aws        secretEngine = "aws"
	azure      secretEngine = "azure"
	consul     secretEngine = "consul"
	kubernetes secretEngine = "kubernetes"
	nomad      secretEngine = "nomad"
	rabbitmq   secretEngine = "rabbitmq"
	ssh        secretEngine = "ssh"
	terraform  secretEngine = "terraform"
	// static secret storage
	keyvalue1 secretEngine = "kv1"
	keyvalue2 secretEngine = "kv2"
)

// secret metadata
type Metadata struct {
	LeaseID       string
	LeaseDuration string
	Renewable     string
	Version       string
}

// secret defines a composite Vault secret configuration
type vaultSecret struct {
	engine  secretEngine
	mount   string
	path    string
	dynamic bool
}

// secret constructor
func NewVaultSecret(engineString string, mount string, path string) (*vaultSecret, error) {
	// validate mandatory fields specified
	if len(engineString) == 0 || len(path) == 0 {
		log.Print("the secret engine and path parameters are mandatory")
		return nil, errors.New("required param(s) missing")
	}

	// validate engine parameter
	engine := secretEngine(engineString)
	if len(engine) == 0 {
		log.Printf("an invalid secrets engine was specified: %s", engineString)
		return nil, errors.New("invalid secret engine")
	}

	// initialize vault secret
	vaultSecret := &vaultSecret{
		engine: engine,
		path:   path,
		mount:  mount,
	}

	// determine if secret is dynamic and default mount point
	// note current schema renders default mount setting pointless, but it would ensure safety to retain
	switch engine {
	case keyvalue1:
		vaultSecret.dynamic = false

		if len(mount) == 0 {
			vaultSecret.mount = "kv"
		}
	case keyvalue2:
		vaultSecret.dynamic = false

		if len(mount) == 0 {
			vaultSecret.mount = "secret"
		}
	case database, aws, azure, consul, kubernetes, nomad, rabbitmq, ssh, terraform:
		vaultSecret.dynamic = true

		if len(mount) == 0 {
			vaultSecret.mount = string(engine)
		}
	default:
		log.Printf("an invalid secret engine %s was selected", engine)
		return nil, errors.New("invalid secret engine")
	}

	return vaultSecret, nil
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
func (secret *vaultSecret) SecretValue(client *vault.Client, version string) (map[string]interface{}, Metadata, error) {
	if secret.dynamic {
		return secret.generateCredentials(client)
	} else {
		return secret.retrieveKVSecret(client, version)
	}
}

// populate key-value pair secrets and return version, metadata, and error (PUT+POST/UPDATE+CREATE/PATCH+WRITE)
func (secret *vaultSecret) PopulateKVSecret(client *vault.Client, secretValue map[string]interface{}, patch bool) (Metadata, error) {
	switch secret.engine {
	case keyvalue1:
		return secret.populateKV1Secret(client, secretValue)
	case keyvalue2:
		return secret.populateKV2Secret(client, secretValue, patch)
	default:
		log.Printf("an invalid secret engine %s was selected", secret.engine)
		return Metadata{}, errors.New("invalid secret engine")
	}
}

// renew dynamic secret lease and return updated metadata
func (secret *vaultSecret) Renew(client *vault.Client, leaseIdSuffix string) (Metadata, error) {
	// semi-validate secret is renewable (better but not possible is *Secret.Renewable)
	if !secret.dynamic {
		log.Printf("the input secret with engine %s at mount %s and path %s is not renewable", secret.engine, secret.mount, secret.path)
		return Metadata{}, nil
	}

	// determine full lease id
	leaseId := secret.mount + "/creds/" + secret.path + "/" + leaseIdSuffix

	// renew the secret lease
	rawSecret, err := client.Sys().Renew(leaseId, 0)
	if err != nil {
		log.Printf("the secret with lease ID %s could not be renewed", leaseId)
		log.Print(err)
		return Metadata{}, err
	}
	log.Printf("the lease for %s was successfully renewed", leaseId)

	// calculate the expiration time for version
	expirationTime := time.Now().Local().Add(time.Second * time.Duration(rawSecret.LeaseDuration))

	// initialize secret metadata and assign version
	metadata := rawSecretToMetadata(rawSecret)
	metadata.Version = expirationTime.String()

	// convert raw secret to metadata and return metadata and version
	return metadata, nil
}
