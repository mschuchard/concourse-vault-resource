package vault

import (
	"testing"

	"github.com/mitodl/concourse-vault-resource/vault/util"
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
		Address: util.VaultAddress,
		Token:   util.VaultToken,
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
		Address: "https://192.168.9.10",
		AWSRole: "myIAMRole",
	}

	if *awsVaultConfig != expectedVaultConfig {
		test.Error("the vault aws config constructor returned unexpected values.")
		test.Errorf("expected vault config: %v", expectedVaultConfig)
		test.Errorf("actual vault config: %v", *awsVaultConfig)
	}
}

// test client token authentication
func TestAuthClient(test *testing.T) {
	if util.VaultClient.Token() != util.VaultToken {
		test.Error("the authenticated Vault client return failed basic validation")
		test.Errorf("expected Vault token: %s, actual: %s", util.VaultToken, util.VaultClient.Token())
	}
}
