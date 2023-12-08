package vault

import (
	"context"
	"log"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/aws"
)

// authentication engine with pseudo-enum
type AuthEngine string

const (
	awsIam AuthEngine = "aws"
	token  AuthEngine = "token"
)

// VaultConfig defines vault api interface config
type VaultConfig struct {
	Engine       AuthEngine
	Address      string
	AWSMountPath string
	AWSRole      string
	Token        string
	Insecure     bool
}

// VaultConfig constructor
func (config *VaultConfig) New() {
	// vault address
	if len(config.Address) == 0 {
		config.Address = "http://127.0.0.1:8200"
	}

	// determine engine if unspecified and validate authentication parameters
	if len(config.Engine) == 0 {
		log.Print("Authentication engine for Vault not specified; using logic from other parameters to assist with determination")

		if len(config.Token) > 0 && len(config.AWSMountPath) > 0 {
			log.Fatal("Token and AWS mount path were simultaneously specified; these are mutually exclusive options")
		}
		if len(config.Token) == 0 {
			log.Print("AWS IAM authentication will be utilized with the Vault client")
			config.Engine = awsIam
		} else {
			log.Print("Token authentication will be utilized with the Vault client")
			config.Engine = token
		}
	}
	if config.Engine == token && len(config.Token) != 28 {
		log.Fatal("the specified Vault Token is invalid")
	}
	if config.Engine == awsIam && len(config.AWSMountPath) == 0 {
		log.Print("using default AWS authentication mount path at 'aws'")
		config.AWSMountPath = "aws"
	}
	if config.Engine == awsIam && len(config.AWSRole) == 0 {
		log.Print("using Vault role in utilized AWS authentication engine with the same name as the current utilized AWS IAM Role")
	}
}

// instantiate authenticated vault client with aws-iam or token auth
func (config *VaultConfig) AuthClient() *vault.Client {
	// initialize config
	VaultConfig := &vault.Config{Address: config.Address}
	err := VaultConfig.ConfigureTLS(&vault.TLSConfig{Insecure: config.Insecure})
	if err != nil {
		log.Print("Vault TLS configuration failed to initialize")
		log.Fatal(err)
	}

	// initialize client
	client, err := vault.NewClient(VaultConfig)
	if err != nil {
		log.Print("Vault client failed to initialize")
		log.Fatal(err)
	}

	// verify vault is unsealed
	sealStatus, err := client.Sys().SealStatus()
	if err != nil {
		log.Print("unable to verify that the Vault cluster is unsealed")
		log.Fatal(err)
	}
	if sealStatus.Sealed {
		log.Fatal("the Vault server cluster is sealed and no operations can be executed")
	}

	// determine authentication method
	switch config.Engine {
	case token:
		client.SetToken(config.Token)
	case awsIam:
		// determine iam role login option
		var loginOption auth.LoginOption

		if len(config.AWSRole) > 0 {
			// use explicitly specified iam role
			loginOption = auth.WithRole(config.AWSRole)
		} else {
			// use default iam role
			loginOption = auth.WithIAMAuth()
		}
		// authenticate with aws iam
		awsAuth, err := auth.NewAWSAuth(loginOption)
		if err != nil {
			log.Print("unable to initialize AWS IAM authentication")
			log.Fatal(err)
		}

		authInfo, err := client.Auth().Login(context.Background(), awsAuth)
		if err != nil {
			log.Print("unable to login with AWS IAM auth method")
			log.Fatal(err)
		}
		if authInfo == nil {
			log.Fatal("no auth info was returned after login")
		}
	}

	// return authenticated vault client
	return client
}
