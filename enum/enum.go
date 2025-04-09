package enum

import (
	"errors"
	"log"
	"slices"
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

// secretengine type conversion
func (s SecretEngine) New() (SecretEngine, error) {
	if !slices.Contains([]SecretEngine{Database, AWS, Azure, Consul, Kubernetes, Nomad, RabbitMQ, SSH, Terraform, KeyValue1, KeyValue2}, s) {
		log.Printf("string %s could not be converted to SecretEngine enum", s)
		return "", errors.New("invalid secretengine enum")
	}
	return s, nil
}
