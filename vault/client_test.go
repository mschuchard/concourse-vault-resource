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
		Address:      util.VaultAddress,
		AuthEngine:   enum.AWSIAM,
		AWSVaultRole: "myIAMRole",
	}
	kubeSourceConfig = concourse.Source{
		Address:             util.VaultAddress,
		AuthEngine:          enum.KubernetesSA,
		KubernetesVaultRole: "mySARole",
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
	if err := authClient(basicSourceConfig, util.VaultClient); err != nil {
		test.Error("authenticating a vault client with a basic token config errored")
		test.Error(err)
	}

	if err := authClient(awsSourceConfig, util.VaultClient); err == nil || !strings.Contains(err.Error(), "NoCredentialProviders: no valid providers in chain") {
		test.Error("authenticating a vault client with aws did not error in the expected manner")
		test.Errorf("expected error (contains): NoCredentialProviders: no valid providers in chain, actual: %v", err)
	}

	awsSourceConfig.AWSVaultRole = ""
	if err := authClient(awsSourceConfig, util.VaultClient); err == nil || !strings.Contains(err.Error(), "NoCredentialProviders: no valid providers in chain") {
		test.Error("authenticating a vault client with aws did not error in the expected manner")
		test.Errorf("expected error (contains): NoCredentialProviders: no valid providers in chain, actual: %v", err)
	}

	if err := authClient(kubeSourceConfig, util.VaultClient); err == nil || !strings.Contains(err.Error(), "error reading service account token from default location") {
		test.Error("authenticating a vault client with kubernetes did not error in the expected manner")
		test.Errorf("expected error (contains): error reading service account token from default location, actual: %v", err)
	}

	// test errors
	invalidAuth := concourse.Source{AuthEngine: "does not exist"}
	if err := authClient(invalidAuth, util.VaultClient); err == nil || err.Error() != "invalid authengine enum" {
		test.Errorf("expected error: invalid authengine enum, actual: %s", err)
	}

	invalidToken := concourse.Source{Token: "foobarbaz123!"}
	if err := authClient(invalidToken, util.VaultClient); err == nil || err.Error() != "invalid vault token" {
		test.Errorf("expected error: invalid vault token, actual: %s", err)
	}

	kubeSourceConfig.KubernetesVaultRole = ""
	if err := authClient(kubeSourceConfig, util.VaultClient); err == nil || err.Error() != "no kubernetes vault role specified" {
		test.Errorf("expected error: no kubernetes vault role specified, actual: %s", err)
	}
}
