package enum

import (
	"errors"
	"log"
)

// authentication engine with pseudo-enum
type AuthEngine string

const (
	AWSIAM     AuthEngine = "aws"
	VaultToken AuthEngine = "token"
)

// authengine type conversion
func (a AuthEngine) New() (AuthEngine, error) {
	if a != AWSIAM && a != VaultToken {
		log.Printf("string %s could not be converted to AuthEngine enum", a)
		return "", errors.New("invalid authengine enum")
	}
	return a, nil
}

// secret engine with pseudo-enum
type SecretEngine string

const (
	// dynamic credential generators
	Database   SecretEngine = "database"
	AWS        SecretEngine = "aws"
	Azure      SecretEngine = "azure"
	Consul     SecretEngine = "consul"
	Kubernetes SecretEngine = "kubernetes"
	Nomad      SecretEngine = "nomad"
	RabbitMQ   SecretEngine = "rabbitmq"
	SSH        SecretEngine = "ssh"
	Terraform  SecretEngine = "terraform"
	// static secret storage
	KeyValue1 SecretEngine = "kv1"
	KeyValue2 SecretEngine = "kv2"
)
