package vault

import (
	"testing"

	"github.com/mitodl/concourse-vault-resource/vault/util"
)

// test secret constructor
func TestNewVaultSecret(test *testing.T) {
	dbVaultSecret, err := NewVaultSecret("database", "", util.KVPath)
	if err != nil {
		test.Error("db secret failed to construct")
		test.Error(err)
	}

	if dbVaultSecret.engine != database || dbVaultSecret.path != util.KVPath || dbVaultSecret.mount != "database" || dbVaultSecret.dynamic != true {
		test.Error("the database Vault secret constructor returned unexpected values")
		test.Errorf("expected engine: %s, actual: %s", dbVaultSecret.engine, database)
		test.Errorf("expected path: %s, actual: %s", dbVaultSecret.path, util.KVPath)
		test.Errorf("expected mount: %s, actual: %s", dbVaultSecret.mount, "database")
		test.Errorf("expected dynamic to be true, actual: %t", dbVaultSecret.dynamic)
	}

	awsVaultSecret, err := NewVaultSecret("aws", "gcp", util.KVPath)
	if err != nil {
		test.Error("aws secret failed to construct")
		test.Error(err)
	}

	if awsVaultSecret.engine != aws || awsVaultSecret.path != util.KVPath || awsVaultSecret.mount != "gcp" || dbVaultSecret.dynamic != true {
		test.Error("the AWS Vault secret constructor returned unexpected values")
		test.Errorf("expected engine: %s, actual: %s", awsVaultSecret.engine, aws)
		test.Errorf("expected path: %s, actual: %s", awsVaultSecret.path, util.KVPath)
		test.Errorf("expected mount: gcp, actual: %s", awsVaultSecret.mount)
		test.Errorf("expected dynamic to be true, actual: %t", dbVaultSecret.dynamic)
	}
}

// test secret renew
func TestRenew(test *testing.T) {}
