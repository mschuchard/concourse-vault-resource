package concourse

import (
	"encoding/json"
	"io"
	"log"
)

// TODO https://itnext.io/how-to-use-golang-generics-with-structs-8cabc9353d75

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

// TODO potentially combine both below with above by converting Paths to any (also probably rename) and doing a bunch of type checks BUT wow that seems like not great cost/benefit
type secretsPut struct {
	Engine string `json:"engine"`
	Patch  bool   `json:"patch"`
	// key is secret path
	Secrets SecretValues `json:"secrets"`
}

type SecretSource struct {
	Engine string `json:"engine"`
	Mount  string `json:"mount"`
	Path   string `json:"path"`
}

// TODO: for future fine-tuning of secret value (enum?)
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
func NewCheckRequest(pipelineJSON io.Reader) *checkRequest {
	// read, decode, and unmarshal the pipeline json io.Reader, and assign to the inRequest pointer
	var checkRequest checkRequest
	if err := json.NewDecoder(pipelineJSON).Decode(&checkRequest); err != nil {
		log.Print("error decoding pipline input from JSON")
		log.Fatal(err)
	}

	// initialize empty version if unspecified
	if checkRequest.Source.Secret.Engine == "kv1" && checkRequest.Version != (Version{}) {
		// validate version not specified for kv1
		log.Fatal("version cannot be specified in conjunction with a kv version 1 engine secret")
	}

	return &checkRequest
}

// checkResponse constructor
func NewCheckResponse(versions []Version) checkResponse {
	// return reference to slice of version
	return versions
}

// inRequest constructor with pipeline param as io.Reader but typically os.Stdin *os.File input because concourse
func NewInRequest(pipelineJSON io.Reader) *inRequest {
	// read, decode, and unmarshal the pipeline json io.Reader, and assign to the inRequest pointer
	var inRequest inRequest
	if err := json.NewDecoder(pipelineJSON).Decode(&inRequest); err != nil {
		log.Print("error decoding pipline input from JSON")
		log.Fatal(err)
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
		log.Fatal("secrets cannot be simultaneously specified in both source and params")
	} else if noSourceSecret && noParamsSecret {
		log.Fatal("one secret must be specified in source, or one or more secrets in params, and neither was specified")
	}

	// return reference
	return &inRequest
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
func NewOutRequest(pipelineJSON io.Reader) *outRequest {
	// read, decode, and unmarshal the pipeline json io.Reader, and assign to the outRequest pointer
	var outRequest outRequest
	if err := json.NewDecoder(pipelineJSON).Decode(&outRequest); err != nil {
		log.Print("error decoding pipline input from JSON")
		log.Fatal(err)
	}
	// validate
	if outRequest.Source.Secret != (SecretSource{}) {
		log.Print("specifying a secret in source for a put step has no effect, and that value will be ignored during this step execution")
	}
	if outRequest.Params == nil {
		log.Fatal("no secret parameters were specified for this put step")
	}

	// return reference
	return &outRequest
}
