package concourse

import (
	"maps"
	"slices"
	"os"
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

	if checkRequest.Version != version || source.AuthEngine != "token" || source.Address != "http://localhost:8200" || !source.Insecure || source.Token != "abcdefghijklmnopqrstuvwxyz09" || source.Secret != (SecretSource{Engine: "kv2", Mount: "secret", Path: "foo/bar"}) {
		test.Error("check request constructor returned unexpected values")
		test.Errorf("expected Version field to be %v, actual: %v", version, checkRequest.Version)
		test.Errorf("expected Source Auth Engine field to be: token, actual: %s", source.AuthEngine)
		test.Errorf("expected Source Address field to be: http://localhost:8200, actual: %s", source.Address)
		test.Errorf("expected Source Insecure field to be: true, actual: %t", source.Insecure)
		test.Errorf("expected Source Token field to be: abcdefghijklmnopqrstuvwxyz09, actual: %s", source.Token)
		test.Errorf("expected Source Secret field to be: {kv2 secret foo/bar}, actual: %v", source.Secret)
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

	inRequest, err := NewInRequest(pipelineJSON)
	if err != nil {
		test.Error("in request failed to construct")
		test.Error(err)
	}

	source := inRequest.Source
	expectedSource := Source{
		AuthEngine: "token",
		Address:    "http://localhost:8200",
		Insecure:   true,
		Token:      "abcdefghijklmnopqrstuvwxyz09",
	}
	params := inRequest.Params
	expectedParams := map[string]secrets{
		"secret": secrets{
			Engine: "kv2",
			Paths:  []string{"foo/bar", "bar/baz"},
		},
		"kv": secrets{
			Engine: "kv1",
			Paths:  []string{"foo/bar"},
		},
	}

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

	outRequest, err := NewOutRequest(pipelineJSON)
	if err != nil {
		test.Error("out request failed to construct")
		test.Error(err)
	}

	source := outRequest.Source
	expectedSource := Source{
		AuthEngine: "token",
		Address:    "http://localhost:8200",
		Insecure:   true,
		Token:      "abcdefghijklmnopqrstuvwxyz09",
	}
	params := outRequest.Params
	expectedParams := map[string]secretsPut{
		"secret": secretsPut{
			Engine: "kv2",
			Secrets:  map[string]secretValue{
				"thefoo": map[string]interface{} {
					"newpassword": "newsecret",
					"newerpassword": "newersecret",
				},
			},
		},
		"kv": secretsPut{
			Engine: "kv1",
			Secrets:  map[string]secretValue{
				"thebar": map[string]interface{} {"key": "value"},
			  "thebaz": map[string]interface{} {"key": "value"},
			},
		},
	}

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
