package concourse

import (
	"encoding/json"
	"errors"
	"io"
	"log"
)

// custom type structs
// key-value pairs would be arbitrary for kv1 and kv2, but are standardized schema for credential generators
type secretValue map[string]interface{}

// key is secret "<mount>-<path>", and value is secret keys and values
type SecretValues map[string]secretValue

// key is "<mount>-<path>" and value is version of secret
type responseVersion map[string]string

type MetadataEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type secrets struct {
	Engine string   `json:"engine"`
	Paths  []string `json:"paths"`
	Renew  bool     `json:"renew"`
}

type secretsPut struct {
	Engine string `json:"engine"`
	Patch  bool   `json:"patch"`
	// key is secret path
	Secrets SecretValues `json:"secrets"`
}

type SecretSource struct {
	Engine  string `json:"engine"`
	Mount   string `json:"mount"`
	Path    string `json:"path"`
	LeaseId string `json:"lease_id"`
}

type dbSecretValue struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type kvSecretValue map[string]interface{}
type awsSecretValue struct {
	AccessKey     string `json:"access_key"`
	SecretKey     string `json:"secret_key"`
	SecurityToken string `json:"security_token,omitempty"`
	ARN           string `json:"arn"`
}

// concourse standard type structs
type Source struct {
	AuthEngine   string       `json:"auth_engine,omitempty"`
	Address      string       `json:"address,omitempty"`
	AWSMountPath string       `json:"aws_mount_path,omitempty"`
	AWSVaultRole string       `json:"aws_vault_role,omitempty"`
	Token        string       `json:"token,omitempty"`
	Insecure     bool         `json:"insecure"`
	Secret       SecretSource `json:"secret"`
}

type Version struct {
	Version string `json:"version"`
}

// check/in/out custom type structs for inputs and outputs
type checkRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type checkResponse []Version

type inRequest struct {
	// key is secret mount
	Params  map[string]secrets `json:"params"`
	Source  Source             `json:"source"`
	Version Version            `json:"version"`
}

type outRequest struct {
	// key is secret mount
	Params map[string]secretsPut `json:"params"`
	Source Source                `json:"source"`
}

type response struct {
	Metadata []MetadataEntry `json:"metadata"`
	Version  responseVersion `json:"version"`
}

// inRequest constructor with pipeline param as io.Reader but typically os.Stdin *os.File input because concourse
func NewCheckRequest(pipelineJSON io.Reader) (*checkRequest, error) {
	// read, decode, and unmarshal the pipeline json io.Reader, and assign to the inRequest pointer
	var checkRequest checkRequest
	if err := json.NewDecoder(pipelineJSON).Decode(&checkRequest); err != nil {
		log.Print("error decoding pipline input from JSON")
		return nil, err
	}

	// validate version not specified for kv1
	if checkRequest.Source.Secret.Engine == "kv1" && checkRequest.Version != (Version{}) {
		log.Print("version cannot be specified in conjunction with a kv version 1 engine secret")
		return nil, errors.New("secret version specified with kv1")
	}

	return &checkRequest, nil
}

// checkResponse constructor
func NewCheckResponse(versions []Version) checkResponse {
	// return slice of version
	return versions
}

// inRequest constructor with pipeline param as io.Reader but typically os.Stdin *os.File input because concourse
func NewInRequest(pipelineJSON io.Reader) (*inRequest, error) {
	// read, decode, and unmarshal the pipeline json io.Reader, and assign to the inRequest pointer
	var inRequest inRequest
	if err := json.NewDecoder(pipelineJSON).Decode(&inRequest); err != nil {
		log.Print("error decoding pipline input from JSON")
		return nil, err
	}

	// these conditionals are evaluated multiple times so assign here
	noSourceSecret := inRequest.Source.Secret == (SecretSource{})
	noParamsSecret := inRequest.Params == nil

	// info message for request version specified and params usage
	if inRequest.Version != (Version{}) && !noParamsSecret {
		log.Print("version is ignored in the get step with params as it must be tied to a specific secret path")
	}

	// validate params versus source.secret
	if !noSourceSecret && !noParamsSecret {
		log.Print("secrets cannot be simultaneously specified in both source and params")
		return nil, errors.New("dual secrets specified")
	} else if noSourceSecret && noParamsSecret {
		log.Print("one secret must be specified in source, or one or more secrets in params, and neither was specified")
		return nil, errors.New("no secrets specified")
	}

	// return reference
	return &inRequest, nil
}

// in/out response constructor
func NewResponse() *response {
	// return initialized reference
	return &response{
		Version:  map[string]string{},
		Metadata: []MetadataEntry{},
	}
}

// outRequest constructor with pipeline param as io.Reader but typically os.Stdin *os.File input because concourse
func NewOutRequest(pipelineJSON io.Reader) (*outRequest, error) {
	// read, decode, and unmarshal the pipeline json io.Reader, and assign to the outRequest pointer
	var outRequest outRequest
	if err := json.NewDecoder(pipelineJSON).Decode(&outRequest); err != nil {
		log.Print("error decoding pipline input from JSON")
		return nil, err
	}
	// validate
	if outRequest.Source.Secret != (SecretSource{}) {
		log.Print("specifying a secret in source for a put step has no effect, and that value will be ignored during this step execution")
	}
	if outRequest.Params == nil {
		log.Print("no secret parameters were specified for this put step")
		return nil, errors.New("empty params")
	}

	// return reference
	return &outRequest, nil
}
