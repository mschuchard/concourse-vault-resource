package vault

import (
	"strings"
	"testing"

	"github.com/mschuchard/concourse-vault-resource/concourse"
	"github.com/mschuchard/concourse-vault-resource/enum"
	"github.com/mschuchard/concourse-vault-resource/vault/util"
)

var (
	basicSourceConfig = concourse.Source{
		Address:    util.VaultAddress,
		AuthEngine: enum.VaultToken,
		Token:      util.VaultToken,
	}
	awsSourceConfig = concourse.Source{
		Address:    util.VaultAddress,
		AuthEngine: enum.AWSIAM,
		VaultRole:  "myIAMRole",
	}
	azSourceConfig = concourse.Source{
		Address:    util.VaultAddress,
		AuthEngine: enum.AzureIMDS,
		VaultRole:  "myAzureRole",
		AzResource: "https://management.azure.com/",
	}
	kubeSourceConfig = concourse.Source{
		Address:    util.VaultAddress,
		AuthEngine: enum.KubernetesSA,
		VaultRole:  "mySARole",
	}
	approleSourceConfig = concourse.Source{
		Address:    util.VaultAddress,
		AuthEngine: enum.AppRole,
	}
)

// test client constructor
func TestNewVaultClient(test *testing.T) {
	basicClient, err := NewVaultClient(basicSourceConfig)
	if err != nil {
		test.Error("authenticating a vault client with a basic token config errored")
		test.Error(err)
	}
	if basicClient.Address() != basicSourceConfig.Address || basicClient.Token() != basicSourceConfig.Token {
		test.Error("the authenticated Vault client return failed basic validation")
		test.Errorf("expected Vault token: %s, actual: %s", basicSourceConfig.Token, basicClient.Token())
		test.Errorf("expected Vault address: %s, actual: %s", basicSourceConfig.Address, basicClient.Address())
	}

	// test errors
	invalidServerConfig := concourse.Source{Address: "https//:foo.com"}
	if _, err := NewVaultClient(invalidServerConfig); err == nil || err.Error() != "parse \"https//:foo.com\": invalid URI for request" {
		test.Errorf("expected error: parse \"https//:foo.com\": invalid URI for request, actual: %s", err)
	}
}

// test client auth
func TestAuthClient(test *testing.T) {
	if err := authClient(awsSourceConfig, util.VaultClient); err == nil || !strings.Contains(err.Error(), "NoCredentialProviders: no valid providers in chain") {
		test.Error("authenticating a vault client with aws did not error in the expected manner")
		test.Errorf("expected error (contains): NoCredentialProviders: no valid providers in chain, actual: %v", err)
	}

	awsSourceConfig.VaultRole = ""
	if err := authClient(awsSourceConfig, util.VaultClient); err == nil || !strings.Contains(err.Error(), "NoCredentialProviders: no valid providers in chain") {
		test.Error("authenticating a vault client with aws did not error in the expected manner")
		test.Errorf("expected error (contains): NoCredentialProviders: no valid providers in chain, actual: %v", err)
	}

	if err := authClient(kubeSourceConfig, util.VaultClient); err == nil || !strings.Contains(err.Error(), "error reading service account token from default location") {
		test.Error("authenticating a vault client with kubernetes did not error in the expected manner")
		test.Errorf("expected error (contains): error reading service account token from default location, actual: %v", err)
	}

	if err := authClient(azSourceConfig, util.VaultClient); err == nil || !strings.Contains(err.Error(), "error calling Azure token endpoint") {
		test.Error("authenticating a vault client with azure did not error in the expected manner")
		test.Errorf("expected error (contains): error calling Azure token endpoint, actual: %v", err)
	}

	approleSourceConfig.VaultRole = util.RoleID
	approleSourceConfig.SecretID = util.SecretID
	if err := authClient(approleSourceConfig, util.VaultClient); err != nil {
		test.Error("authenticating a vault client with approle config errored")
		test.Error(err)
	}

	// reset client auth to prep for next test
	util.VaultClient.SetToken(util.VaultToken)

	// retrieve a wrapped secret id for testing approle auth in "pull" mode
	util.VaultClient.SetWrappingLookupFunc(func(operation, path string) string {
		if path == "auth/approle/role/myAppRole/secret-id" {
			return "60s"
		}
		return ""
	})
	wrappedSecretID, err := util.VaultClient.Logical().Write("auth/approle/role/myAppRole/secret-id", nil)
	// reset the wrapping lookup func immediately so it does not affect subsequent requests
	util.VaultClient.SetWrappingLookupFunc(nil)
	if err != nil {
		test.Error("failed to retrieve wrapped secret ID for approle pull auth")
		test.Error(err)
	}
	// access wrapping token from wrapped secret id and assign to source config
	if wrappedSecretID == nil || wrappedSecretID.WrapInfo == nil || len(wrappedSecretID.WrapInfo.Token) == 0 {
		test.Error("the secret id write did not return a wrapped response")
	}
	approleSourceConfig.WrapToken = wrappedSecretID.WrapInfo.Token

	if err := authClient(approleSourceConfig, util.VaultClient); err != nil {
		test.Error("authenticating a vault client with approle pull (wrapping token) config errored")
		test.Error(err)
	}

	// this needs to be last to ensure the client is authenticated with a root token for all other tests
	if err := authClient(basicSourceConfig, util.VaultClient); err != nil {
		test.Error("authenticating a vault client with a basic token config errored")
		test.Error(err)
	}

	// test errors
	invalidAuth := concourse.Source{AuthEngine: "does not exist"}
	if err := authClient(invalidAuth, util.VaultClient); err == nil || err.Error() != "invalid authengine enum" {
		test.Errorf("expected error: invalid authengine enum, actual: %s", err)
	}

	invalidToken := concourse.Source{AuthEngine: enum.VaultToken, Token: "foobarbaz123!"}
	if err := authClient(invalidToken, util.VaultClient); err == nil || err.Error() != "invalid vault token" {
		test.Errorf("expected error: invalid vault token, actual: %s", err)
	}

	kubeSourceConfig.VaultRole = ""
	if err := authClient(kubeSourceConfig, util.VaultClient); err == nil || err.Error() != "no kubernetes vault role specified" {
		test.Errorf("expected error: no kubernetes vault role specified, actual: %s", err)
	}

	azSourceConfig.VaultRole = ""
	if err := authClient(azSourceConfig, util.VaultClient); err == nil || err.Error() != "no azure vault role specified" {
		test.Errorf("expected error: no azure vault role specified, actual: %s", err)
	}

	approleSourceConfig.VaultRole = ""
	if err := authClient(approleSourceConfig, util.VaultClient); err == nil || err.Error() != "approle credentials absent" {
		test.Errorf("expected error: approle credentials absent, actual: %s", err)
	}
}

// test default mount
func TestCheckAuthParams(test *testing.T) {
	if mount := checkAuthParams("", "", enum.KubernetesSA); mount != "kubernetes" {
		test.Errorf("expected default mount: kubernetes, actual: %s", mount)
	}

	if mount := checkAuthParams("gcp", "", enum.AWSIAM); mount != "gcp" {
		test.Errorf("expected mount input param: gcp, actual: %s", mount)
	}
}

// test vault authenticate with authentication method
func TestLoginWithMethod(test *testing.T) {
	// for now this is encapsulated by TestAuthClient, but in the future it may be useful to here also
}
