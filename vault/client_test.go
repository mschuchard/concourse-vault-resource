package vault

import (
	"testing"

	"github.com/mschuchard/concourse-vault-resource/vault/util"
)

// test config constructor
func TestNewVaultConfig(test *testing.T) {
	basicVaultConfig := &VaultConfig{
		Address: util.VaultAddress,
		Token:   util.VaultToken,
	}
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

	awsVaultConfig := &VaultConfig{
		Address: "https://192.168.9.10",
		AWSRole: "myIAMRole",
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
}

// test client token authentication
func TestAuthClient(test *testing.T) {
	basicVaultConfig := &VaultConfig{
		Address:  util.VaultAddress,
		Engine:   token,
		Token:    util.VaultToken,
		Insecure: true,
	}
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
}
