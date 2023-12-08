package vault

import "testing"

// globals for vault package testing
const (
	KVPath   = "foo/bar"
	KVKey    = "password"
	KVValue  = "supersecret"
	KV1Mount = "kv"
	KV2Mount = "secret"
)

// test secret constructor
func TestNewVaultSecret(test *testing.T) {
	dbVaultSecret := NewVaultSecret("database", "", KVPath)
	if dbVaultSecret.engine != database || dbVaultSecret.path != KVPath || dbVaultSecret.mount != "database" || dbVaultSecret.dynamic != true {
		test.Error("the database Vault secret constructor returned unexpected values")
		test.Errorf("expected engine: %s, actual: %s", dbVaultSecret.engine, database)
		test.Errorf("expected path: %s, actual: %s", dbVaultSecret.path, KVPath)
		test.Errorf("expected mount: %s, actual: %s", dbVaultSecret.mount, "database")
		test.Errorf("expected dynamic to be true, actual: %t", dbVaultSecret.dynamic)
	}

	awsVaultSecret := NewVaultSecret("aws", "gcp", KVPath)
	if awsVaultSecret.engine != aws || awsVaultSecret.path != KVPath || awsVaultSecret.mount != "gcp" || dbVaultSecret.dynamic != true {
		test.Error("the AWS Vault secret constructor returned unexpected values")
		test.Errorf("expected engine: %s, actual: %s", awsVaultSecret.engine, aws)
		test.Errorf("expected path: %s, actual: %s", awsVaultSecret.path, KVPath)
		test.Errorf("expected mount: gcp, actual: %s", awsVaultSecret.mount)
		test.Errorf("expected dynamic to be true, actual: %t", dbVaultSecret.dynamic)
	}
}

// test secret Read operation

// test secret renew
