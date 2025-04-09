package enum

import "testing"

func TestAuthEngineNew(test *testing.T) {
	authEngine, err := AuthEngine("token").New()
	if err != nil {
		test.Error(err)
	}
	if authEngine != VaultToken {
		test.Error("authengine did not type convert correctly")
		test.Errorf("expected: token, actual: %s", authEngine)
	}

	if _, err = AuthEngine("foo").New(); err == nil || err.Error() != "invalid authengine enum" {
		test.Error("authengine type conversion did not error expectedly")
		test.Errorf("expected: invalid authengine enum, actual: %s", err)
	}
}

func TestSecretEngineNew(test *testing.T) {
	secretEngine, err := SecretEngine("kubernetes").New()
	if err != nil {
		test.Error(err)
	}
	if secretEngine != Kubernetes {
		test.Error("secretengine did not type convert correctly")
		test.Errorf("expected: kubernetes, actual: %s", secretEngine)
	}

	if _, err = SecretEngine("foo").New(); err == nil || err.Error() != "invalid secretengine enum" {
		test.Error("secretengine type conversion did not error expectedly")
		test.Errorf("expected: invalid secretengine enum, actual: %s", err)
	}
}
