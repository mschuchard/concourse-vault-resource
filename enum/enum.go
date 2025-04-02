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
