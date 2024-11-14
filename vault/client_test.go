package vault

import (
	"strings"
	"testing"

	"github.com/mschuchard/concourse-vault-resource/concourse"
	"github.com/mschuchard/concourse-vault-resource/vault/util"
)

var (
	basicSourceConfig = concourse.Source{
		Address: util.VaultAddress,
		Token:   util.VaultToken,
	}
	awsSourceConfig = concourse.Source{
		Address:      "https://192.168.9.10",
		AWSVaultRole: "myIAMRole",
	}
)

// test config constructor
func TestNewVaultConfig(test *testing.T) {
	vaultConfig, err := NewVaultConfig(basicSourceConfig)
	if err != nil {
		test.Error("the basic vault config did not successfully validate")
		test.Error(err)
	}
	expectedVaultConfig := VaultConfig{
		Address:  util.VaultAddress,
		Engine:   token,
		Token:    util.VaultToken,
		Insecure: true,
	}

	if *vaultConfig != expectedVaultConfig {
		test.Error("the vault basic config constructor returned unexpected values.")
		test.Errorf("expected vault config: %v", expectedVaultConfig)
		test.Errorf("actual vault config: %v", *vaultConfig)
	}

	vaultConfig, err = NewVaultConfig(awsSourceConfig)
	if err != nil {
		test.Error("the aws vault config did not successfully validate")
		test.Error(err)
	}
	expectedVaultConfig = VaultConfig{
		Address:      "https://192.168.9.10",
		Engine:       awsIam,
		AWSMountPath: "aws",
		AWSRole:      "myIAMRole",
	}

	if *vaultConfig != expectedVaultConfig {
		test.Error("the vault aws config constructor returned unexpected values.")
		test.Errorf("expected vault config: %v", expectedVaultConfig)
		test.Errorf("actual vault config: %v", *vaultConfig)
	}

	invalidServerConfig := concourse.Source{Address: "https//:foo.com"}
	if _, err := NewVaultConfig(invalidServerConfig); err == nil {
		test.Error("invalid vault server address did not error as expected")
	}

	ambiguousAuth := concourse.Source{
		Token:        util.VaultToken,
		AWSMountPath: "gcp",
	}
	if _, err := NewVaultConfig(ambiguousAuth); err == nil || err.Error() != "unable to deduce authentication engine" {
		test.Error("ambiguous unspecified authentication engine did not error as expected")
	}

	invalidAuth := concourse.Source{AuthEngine: "does not exist"}
	if _, err := NewVaultConfig(invalidAuth); err == nil || err.Error() != "invalid Vault authentication engine" {
		test.Error("invalid authentication engine did not error as expected")
	}

	invalidToken := concourse.Source{Token: "foobarbaz"}
	if _, err := NewVaultConfig(invalidToken); err == nil || err.Error() != "invalid vault token" {
		test.Error("invalid vault token did not error as expected")
	}
}

// test client token authentication
func TestAuthClient(test *testing.T) {
	vaultConfig, err := NewVaultConfig(basicSourceConfig)
	if err != nil {
		test.Error("the basic vault config did not successfully validate")
		test.Error(err)
	}
	basicClient, err := NewClient(vaultConfig)
	if err != nil {
		test.Error("authenticating a vault client with a basic token config errored")
		test.Error(err)
	}
	if basicClient.Address() != basicSourceConfig.Address || basicClient.Token() != basicSourceConfig.Token {
		test.Error("the authenticated Vault client return failed basic validation")
		test.Errorf("expected Vault token: %s, actual: %s", basicSourceConfig.Token, basicClient.Token())
		test.Errorf("expected Vault address: %s, actual: %s", basicSourceConfig.Address, basicClient.Address())
	}

	awsSourceConfig.Address = util.VaultAddress
	vaultConfig, err = NewVaultConfig(awsSourceConfig)
	if err != nil {
		test.Error("the aws vault config did not successfully validate")
		test.Error(err)
	}
	if _, err = NewClient(vaultConfig); err == nil || !strings.Contains(err.Error(), "NoCredentialProviders: no valid providers in chain") {
		test.Error("authenticating a vault client with aws did not error in the expected manner")
		test.Errorf("expected error (contains): NoCredentialProviders: no valid providers in chain, actual: %v", err)
	}
}
