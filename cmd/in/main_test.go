package main

import (
	"os"
	"testing"
)

func ExampleMain() {
	// defer stdin close and establish workdir from argsp[1]
	defer os.Stdin.Close()
	os.Args[1] = "/opt/resource"

	// params secrets and source secret
	for _, secretKey := range []string{"params", "source"} {
		// deliver test pipeline file content as stdin to "in" the same as actual pipeline execution
		os.Stdin, _ = os.OpenFile("fixtures/token_kv_"+secretKey+".json", os.O_RDONLY, 0o644)

		// invoke main and validate stdout
		main()
		// Output: {"metadata":[{}],"version":{}}
	}
}

func TestMain(test *testing.T) {
	// defer stdin close and establish workdir from argsp[1]
	defer os.Stdin.Close()
	os.Args[1] = "/opt/resource"

	// deliver test pipeline file content as stdin to "in" the same as actual pipeline execution
	os.Stdin, _ = os.OpenFile("fixtures/token_kv_params.json", os.O_RDONLY, 0o644)

	// invoke main and validate stdout
	main()

	// verify vault.json output
	secretsContents := `{"kv-foo/bar":{"password":"supersecret"},"secret-bar/baz":{"password":"supersecret"},"secret-foo/bar":{"other_password":"ultrasecret","password":"supersecret"}}`
	secretsFile, _ := os.ReadFile("/opt/resource/vault.json")
	if string(secretsFile) != secretsContents {
		test.Error("vault.json did not contain expected secrets data")
		test.Errorf("actual file contents: %s", secretsFile)
		test.Errorf("expected file contents: %s", secretsContents)
	}
}
