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

	// retrieve role id and secret id for testing approle auth
	roleID, err := util.VaultClient.Logical().Read("auth/approle/role/myAppRole/role-id")
	if err != nil {
		test.Error("failed to retrieve role ID for approle auth")
		test.Error(err)
	}
	secretID, err := util.VaultClient.Logical().Write("auth/approle/role/myAppRole/secret-id", nil)
	if err != nil {
		test.Error("failed to retrieve secret ID for approle auth")
		test.Error(err)
	}
	approleSourceConfig.VaultRole = roleID.Data["role_id"].(string)
	approleSourceConfig.SecretID = secretID.Data["secret_id"].(string)

	if err := authClient(approleSourceConfig, util.VaultClient); err != nil {
		test.Error("authenticating a vault client with approle config errored")
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

	approleSourceConfig.VaultRole = ""
	if err := authClient(approleSourceConfig, util.VaultClient); err == nil || err.Error() != "approle credentials absent" {
		test.Errorf("expected error: approle credentials absent, actual: %s", err)
	}
}
