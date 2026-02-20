package concourse

import (
	"encoding/json"
	"maps"
	"os"
	"slices"
	"testing"
)

const versionKey = "secret-foo/bar"

var respVersion = map[string]string{versionKey: "1"}
var version = Version{Version: "1"}

// test checkrequest constructor
func TestCheckRequest(test *testing.T) {
	pipelineJSON, err := os.OpenFile("../cmd/check/fixtures/token_kv.json", os.O_RDONLY, 0o444)
	if err != nil {
		test.Error(err)
	}

	checkRequest, err := NewCheckRequest(pipelineJSON)
	if err != nil {
		test.Error("check request failed to construct")
		test.Error(err)
	}
	source := checkRequest.Source
	expectedSecretSource := SecretSource{Engine: "kv2", Mount: "secret", Path: "foo/bar"}

	if checkRequest.Version != version || source.AuthEngine != "token" || source.Address != "http://localhost:8200" || !source.Insecure || source.Token != "abcdefghijklmnopqrstuvwxyz09" || source.VaultRole != "myrole" || source.AuthMount != "placeholder" || source.Secret != expectedSecretSource {
		test.Error("check request constructor returned unexpected values")
		test.Errorf("expected Version field to be %v, actual: %v", version, checkRequest.Version)
		test.Errorf("expected Source Auth Engine field to be: token, actual: %s", source.AuthEngine)
		test.Errorf("expected Source Address field to be: http://localhost:8200, actual: %s", source.Address)
		test.Errorf("expected Source Insecure field to be: true, actual: %t", source.Insecure)
		test.Errorf("expected Source Token field to be: abcdefghijklmnopqrstuvwxyz09, actual: %s", source.Token)
		test.Errorf("expected Source Vault Role field to be: myrole, actual: %s", source.VaultRole)
		test.Errorf("expected Source Auth Mount field to be: placeholder, actual: %s", source.AuthMount)
		test.Errorf("expected Source Secret field to be: %v, actual: %v", expectedSecretSource, source.Secret)
	}

	pipelineJSON, err = os.OpenFile("../cmd/check/fixtures/bad_lease_id.json", os.O_RDONLY, 0o444)
	if err != nil {
		test.Error(err)
	}

	if _, err = NewCheckRequest(pipelineJSON); err == nil || err.Error() != "invalid lease id parameter" {
		test.Error("invalid lease id parameter value did not fail validation")
	}
}

// test checkresponse constructor
func TestCheckResponse(test *testing.T) {
	checkResponse := NewCheckResponse([]Version{version})

	if len(checkResponse) != 1 || checkResponse[0] != version {
		test.Error("the check response constructor returned an unexpected value")
		test.Errorf("expected value: &[], actual: %v", checkResponse)
	}
}

// test inRequest constructor
func TestNewInRequest(test *testing.T) {
	pipelineJSON, err := os.OpenFile("../cmd/in/fixtures/token_kv_params.json", os.O_RDONLY, 0o444)
	if err != nil {
		test.Error(err)
	}

	newInRequest, err := NewInRequest(pipelineJSON)
	if err != nil {
		test.Error("in request failed to construct")
		test.Error(err)
	}

	pipelineJSON, err = os.OpenFile("../cmd/in/fixtures/token_kv_params.json", os.O_RDONLY, 0o444)
	if err != nil {
		test.Error(err)
	}

	var expectedIn inRequest
	json.NewDecoder(pipelineJSON).Decode(&expectedIn)

	source := newInRequest.Source
	expectedSource := expectedIn.Source
	params := newInRequest.Params
	expectedParams := expectedIn.Params

	if source != expectedSource || params["secret"].Engine != expectedParams["secret"].Engine || !slices.Equal(params["secret"].Paths, expectedParams["secret"].Paths) || params["kv"].Engine != expectedParams["kv"].Engine || !slices.Equal(params["kv"].Paths, expectedParams["kv"].Paths) {
		test.Error("in request constructor returned unexpected values")
		test.Errorf("expected Source field to be %v, actual: %v", expectedSource, source)
		test.Errorf("expected Params field to be %v, actual: %v", expectedParams, params)
	}
}

// test inResponse constructor
func TestNewInResponse(test *testing.T) {
	inResponse := NewResponse()
	inResponse.Version = respVersion

	if len(inResponse.Metadata) != 0 || !maps.Equal(inResponse.Version, respVersion) {
		test.Error("the in response constructor returned unexpected values")
		test.Errorf("expected Metadata field to be empty slice, actual: %v", inResponse.Metadata)
		test.Errorf("expected Version to be: %v, actual: %v", respVersion, inResponse.Version)
	}
}

// test outRequest constructor
func TestNewOutRequest(test *testing.T) {
	pipelineJSON, err := os.OpenFile("../cmd/out/fixtures/token_kv.json", os.O_RDONLY, 0o444)
	if err != nil {
		test.Error(err)
	}

	newOutRequest, err := NewOutRequest(pipelineJSON)
	if err != nil {
		test.Error("out request failed to construct")
		test.Error(err)
	}

	pipelineJSON, err = os.OpenFile("../cmd/out/fixtures/token_kv.json", os.O_RDONLY, 0o444)
	if err != nil {
		test.Error(err)
	}

	var expectedOut outRequest
	json.NewDecoder(pipelineJSON).Decode(&expectedOut)

	source := newOutRequest.Source
	expectedSource := expectedOut.Source
	params := newOutRequest.Params
	expectedParams := expectedOut.Params

	if source != expectedSource || params["secret"].Engine != expectedParams["secret"].Engine || !maps.Equal(params["secret"].Secrets["thefoo"], expectedParams["secret"].Secrets["thefoo"]) || params["kv"].Engine != expectedParams["kv"].Engine || !maps.Equal(params["kv"].Secrets["thebar"], expectedParams["kv"].Secrets["thebar"]) || !maps.Equal(params["kv"].Secrets["thebaz"], expectedParams["kv"].Secrets["thebaz"]) {
		test.Error("out request constructor returned unexpected values")
		test.Errorf("expected Source field to be %v, actual: %v", expectedSource, source)
		test.Errorf("expected Params field to be %v, actual: %v", expectedParams, params)
	}
}

// test outResponse constructor
func TestOutResponse(test *testing.T) {
	outResponse := NewResponse()

	if len(outResponse.Metadata) != 0 || len(outResponse.Version) != 0 {
		test.Error("the out response constructor returned unexpected values")
		test.Errorf("expected Metadata field to be slice of one element, actual: %v", outResponse.Metadata)
		test.Errorf("expected Metadata field only element to be empty map, actual: %v", outResponse.Metadata[0])
		test.Errorf("expected Version to be empty map, actual: %v", outResponse.Version)
	}
}
