package vault

import (
	"strings"
	"testing"

	"github.com/mschuchard/concourse-vault-resource/vault/util"
)

var (
	basicVaultConfig = &VaultConfig{
		Address: util.VaultAddress,
		Token:   util.VaultToken,
	}
	awsVaultConfig = &VaultConfig{
		Address: "https://192.168.9.10",
		AWSRole: "myIAMRole",
	}
)

// test config constructor
func TestNewVaultConfig(test *testing.T) {
	if err := basicVaultConfig.New(); err != nil {
		test.Error("the basic vault config did not successfully validate")
		test.Error(err)
	}
	expectedVaultConfig := VaultConfig{
		Address:  util.VaultAddress,
		Engine:   token,
		Token:    util.VaultToken,
		Insecure: true,
	}

	if *basicVaultConfig != expectedVaultConfig {
		test.Error("the vault basic config constructor returned unexpected values.")
		test.Errorf("expected vault config: %v", expectedVaultConfig)
		test.Errorf("actual vault config: %v", *basicVaultConfig)
	}

	if err := awsVaultConfig.New(); err != nil {
		test.Error("the aws vault config did not successfully validate")
		test.Error(err)
	}
	expectedVaultConfig = VaultConfig{
		Address:      "https://192.168.9.10",
		Engine:       awsIam,
		AWSMountPath: "aws",
		AWSRole:      "myIAMRole",
	}

	if *awsVaultConfig != expectedVaultConfig {
		test.Error("the vault aws config constructor returned unexpected values.")
		test.Errorf("expected vault config: %v", expectedVaultConfig)
		test.Errorf("actual vault config: %v", *awsVaultConfig)
	}

	invalidServerConfig := VaultConfig{Address: "https//:foo.com"}
	if err := invalidServerConfig.New(); err == nil || err.Error() != "invalid Vault server address" {
		test.Error("invalid vault server address did not error as expected")
	}

	ambiguousAuth := VaultConfig{
		Token:        util.VaultToken,
		AWSMountPath: "gcp",
	}
	if err := ambiguousAuth.New(); err == nil || err.Error() != "unable to deduce authentication engine" {
		test.Error("ambiguous unspecified authentication engine did not error as expected")
	}

	invalidAuth := VaultConfig{Engine: "does not exist"}
	if err := invalidAuth.New(); err == nil || err.Error() != "invalid Vault authentication engine" {
		test.Error("invalid authentication engine did not error as expected")
	}

	invalidToken := VaultConfig{Token: "foobarbaz"}
	if err := invalidToken.New(); err == nil || err.Error() != "invalid vault token" {
		test.Error("invalid vault token did not error as expected")
	}
}

// test client token authentication
func TestAuthClient(test *testing.T) {
	basicVaultConfig.New()
	basicClient, err := basicVaultConfig.AuthClient()
	if err != nil {
		test.Error("authenticating a vault client with a basic token config errored")
		test.Error(err)
	}
	if basicClient.Address() != basicVaultConfig.Address || basicClient.Token() != basicVaultConfig.Token {
		test.Error("the authenticated Vault client return failed basic validation")
		test.Errorf("expected Vault token: %s, actual: %s", basicVaultConfig.Token, basicClient.Token())
		test.Errorf("expected Vault address: %s, actual: %s", basicVaultConfig.Address, basicClient.Address())
	}

	awsVaultConfig.Address = util.VaultAddress
	awsVaultConfig.New()
	if _, err = awsVaultConfig.AuthClient(); err == nil || !strings.Contains(err.Error(), "NoCredentialProviders: no valid providers in chain") {
		test.Error("authenticating a vault client with aws did not error in the expected manner")
		test.Errorf("expected error (contains): NoCredentialProviders: no valid providers in chain, actual: %v", err)
	}
}
