package vault

import (
	"testing"

	"github.com/mschuchard/concourse-vault-resource/enum"
	"github.com/mschuchard/concourse-vault-resource/vault/util"
)

// test secret constructor
func TestNewVaultSecret(test *testing.T) {
	dbVaultSecret, err := NewVaultSecret("database", "", util.KVPath)
	if err != nil {
		test.Error("db secret failed to construct")
		test.Error(err)
	}
	expectedVaultSecret := vaultSecret{
		engine:  enum.Database,
		mount:   "database",
		path:    util.KVPath,
		dynamic: true,
	}

	if *dbVaultSecret != expectedVaultSecret {
		test.Error("the database vault secret constructor returned unexpected values")
		test.Errorf("expected values: %v", expectedVaultSecret)
		test.Errorf("actual values: %v", *dbVaultSecret)
	}

	awsVaultSecret, err := NewVaultSecret("aws", "gcp", util.KVPath)
	if err != nil {
		test.Error("aws secret failed to construct")
		test.Error(err)
	}
	expectedVaultSecret = vaultSecret{
		engine:  enum.AWS,
		mount:   "gcp",
		path:    util.KVPath,
		dynamic: true,
	}

	if *awsVaultSecret != expectedVaultSecret {
		test.Error("the aws vault secret constructor returned unexpected values")
		test.Errorf("expected values: %v", expectedVaultSecret)
		test.Errorf("actual values: %v", *awsVaultSecret)
	}

	if _, err = NewVaultSecret("", "", ""); err == nil || err.Error() != "required param(s) missing" {
		test.Error("constructor did not return expected error for missing parameters")
		test.Errorf("expected: required param(s) missing, actual: %s", err)
	}

	if _, err = NewVaultSecret("foo", "bar", "baz"); err == nil || err.Error() != "invalid secretengine enum" {
		test.Error("constructor did not return expected error for invalid secrets engine")
		test.Errorf("expected: invalid secretengine enum, actual: %s", err)
	}
}

// test secret renew
func TestRenew(test *testing.T) {
	staticSecret := vaultSecret{dynamic: false}
	if _, err := staticSecret.Renew(util.VaultClient, ""); err == nil || err.Error() != "non-renewable secret" {
		test.Error("renew did not return expected error for non-dynamic secret")
		test.Errorf("expected: non-renewable secret, actual: %s", err)
	}
}
