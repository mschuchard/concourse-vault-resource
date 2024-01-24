package vault

import (
	"testing"

	"github.com/mitodl/concourse-vault-resource/vault/util"
)

// test config constructor
func TestNewVaultConfig(test *testing.T) {
	basicVaultConfig := &VaultConfig{
		Address:  util.VaultAddress,
		Token:    util.VaultToken,
		Insecure: true,
	}
	if err := basicVaultConfig.New(); err != nil {
		test.Error("the basic vault config did not successfully validate")
		test.Error(err)
	}

	if basicVaultConfig.Engine != token || basicVaultConfig.Address != util.VaultAddress || len(basicVaultConfig.AWSMountPath) != 0 || len(basicVaultConfig.AWSRole) != 0 || basicVaultConfig.Token != util.VaultToken || !basicVaultConfig.Insecure {
		test.Error("the Vault config constructor returned unexpected values.")
		test.Errorf("expected Auth Engine: %s, actual: %s", token, basicVaultConfig.Engine)
		test.Errorf("expected Vault Address: %s, actual: %s", util.VaultAddress, basicVaultConfig.Address)
		test.Errorf("expected AWS Mount Path: (empty), actual: %s", basicVaultConfig.AWSMountPath)
		test.Errorf("expected AWS IAM Role: (empty), actual: %s", basicVaultConfig.AWSRole)
		test.Errorf("expected Vault Token: %s, actual: %s", util.VaultToken, basicVaultConfig.Token)
		test.Errorf("expected Vault Insecure: true, actual: %t", basicVaultConfig.Insecure)
	}

	awsVaultConfig := &VaultConfig{
		Address: "https://192.168.9.10",
		AWSRole: "myIAMRole",
	}
	if err := awsVaultConfig.New(); err != nil {
		test.Error("the aws vault config did not successfully validate")
		test.Error(err)
	}

	if awsVaultConfig.Engine != awsIam || awsVaultConfig.Address != "https://192.168.9.10" || awsVaultConfig.AWSMountPath != "aws" || awsVaultConfig.AWSRole != "myIAMRole" || len(awsVaultConfig.Token) != 0 || awsVaultConfig.Insecure {
		test.Error("the Vault config constructor returned unexpected values.")
		test.Errorf("expected Auth Engine: %s, actual: %s", awsIam, awsVaultConfig.Engine)
		test.Errorf("expected Vault Address: https://192.168.9.10, actual: %s", awsVaultConfig.Address)
		test.Errorf("expected AWS Mount Path: aws, actual: %s", awsVaultConfig.AWSMountPath)
		test.Errorf("expected AWS IAM Role: myIAMRole, actual: %s", awsVaultConfig.AWSRole)
		test.Errorf("expected Vault Token: (empty), actual: %s", awsVaultConfig.Token)
		test.Errorf("expected Vault Insecure: false, actual: %t", awsVaultConfig.Insecure)
	}
}

// test client token authentication
func TestAuthClient(test *testing.T) {
	if util.VaultClient.Token() != util.VaultToken {
		test.Error("the authenticated Vault client return failed basic validation")
		test.Errorf("expected Vault token: %s, actual: %s", util.VaultToken, util.VaultClient.Token())
	}
}
