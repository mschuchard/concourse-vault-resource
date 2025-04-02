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
	/*basicVaultConfig, err := NewVaultConfig(basicSourceConfig)
	if err != nil {
		test.Error("the basic vault config did not successfully validate")
		test.Error(err)
	}
	expectedVaultConfig := vaultConfig{
		address:  util.VaultAddress,
		engine:   token,
		token:    util.VaultToken,
		insecure: true,
	}

	if *basicVaultConfig != expectedVaultConfig {
		test.Error("the vault basic config constructor returned unexpected values.")
		test.Errorf("expected vault config: %v", expectedVaultConfig)
		test.Errorf("actual vault config: %v", *basicVaultConfig)
	}

	basicVaultConfig, err = NewVaultConfig(awsSourceConfig)
	if err != nil {
		test.Error("the aws vault config did not successfully validate")
		test.Error(err)
	}
	expectedVaultConfig = vaultConfig{
		address:      "https://192.168.9.10",
		engine:       awsIam,
		awsMountPath: "aws",
		awsRole:      "myIAMRole",
	}

	if *basicVaultConfig != expectedVaultConfig {
		test.Error("the vault aws config constructor returned unexpected values.")
		test.Errorf("expected vault config: %v", expectedVaultConfig)
		test.Errorf("actual vault config: %v", *basicVaultConfig)
	}*/

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

	awsSourceConfig.Address = util.VaultAddress
	if _, err = NewVaultClient(awsSourceConfig); err == nil || !strings.Contains(err.Error(), "NoCredentialProviders: no valid providers in chain") {
		test.Error("authenticating a vault client with aws did not error in the expected manner")
		test.Errorf("expected error (contains): NoCredentialProviders: no valid providers in chain, actual: %v", err)
	}

	// test errors in reverse validation order
	invalidAuth := concourse.Source{AuthEngine: "does not exist"}
	if _, err := NewVaultClient(invalidAuth); err == nil || err.Error() != "invalid Vault authentication engine" {
		test.Errorf("expected error: invalid Vault authentication engine, actual: %s", err)
	}

	invalidToken := concourse.Source{Token: "foobarbaz"}
	if _, err := NewVaultClient(invalidToken); err == nil || err.Error() != "invalid vault token" {
		test.Errorf("expected error: invalid vault token, actual: %s", err)
	}

	ambiguousAuth := concourse.Source{
		Token:        util.VaultToken,
		AWSMountPath: "gcp",
	}
	if _, err := NewVaultClient(ambiguousAuth); err == nil || err.Error() != "unable to deduce authentication engine" {
		test.Errorf("expected error: unable to deduce authentication engine, actual: %s", err)
	}

	invalidServerConfig := concourse.Source{Address: "https//:foo.com"}
	if _, err := NewVaultClient(invalidServerConfig); err == nil || err.Error() != "invalid Vault server address" {
		test.Errorf("expected error: invalid Vault server address, actual: %s", err)
	}
}
