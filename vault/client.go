package vault

import (
	"context"
	"errors"
	"log"
	"net/url"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/aws"
	"github.com/mschuchard/concourse-vault-resource/concourse"
)

// authentication engine with pseudo-enum
type AuthEngine string

const (
	awsIam AuthEngine = "aws"
	token  AuthEngine = "token"
)

// vaultConfig defines vault api interface config
type vaultConfig struct {
	engine       AuthEngine
	address      string
	awsMountPath string
	awsRole      string
	token        string
	insecure     bool
}

// VaultConfig constructor
func NewVaultConfig(source concourse.Source) (*vaultConfig, error) {
	// vault address default
	if len(source.Address) == 0 {
		source.Address = "http://127.0.0.1:8200"
	} else {
		// vault address validation
		if url, err := url.ParseRequestURI(source.Address); err != nil || len(url.Scheme) == 0 || len(url.Host) == 0 {
			log.Printf("%s is not a valid Vault server address", source.Address)

			// assign err if it is nil
			if err == nil {
				err = errors.New("invalid Vault server address")
			}

			return nil, err
		}
	}

	// insecure validation
	if !source.Insecure && source.Address[0:5] == "http:" {
		log.Print("insecure input parameter was omitted or specified as false, and address protocol is http")
		log.Print("insecure will be reset to value of true")
		source.Insecure = true
	}

	authEngine := AuthEngine(source.AuthEngine)
	if len(authEngine) == 0 {
		log.Print("authentication engine for Vault not specified or otherwise unknown; using logic from other parameters to assist with determination")

		if len(source.Token) > 0 && len(source.AWSMountPath) > 0 {
			log.Print("token and AWS mount path were simultaneously specified; these are mutually exclusive options")
			log.Print("intended authentication engine could not be determined from other parameters")
			return nil, errors.New("unable to deduce authentication engine")
		}
		if len(source.Token) == 0 {
			log.Print("AWS IAM authentication will be utilized with the Vault client")
			authEngine = awsIam
		} else {
			log.Print("Token authentication will be utilized with the Vault client")
			authEngine = token
		}
	} else if authEngine != awsIam && authEngine != token { // validate engine if unspecified
		log.Printf("%v was input as an authentication engine, but only token and aws are supported", authEngine)
		return nil, errors.New("invalid Vault authentication engine")
	}

	// validate vault token
	if authEngine == token && len(source.Token) != 28 {
		log.Print("the specified Vault Token is invalid")
		return nil, errors.New("invalid vault token")
	}

	// default aws mount path and role
	if authEngine == awsIam {
		if len(source.AWSMountPath) == 0 {
			log.Print("using default AWS authentication mount path at 'aws'")
			source.AWSMountPath = "aws"
		}
		if len(source.AWSVaultRole) == 0 {
			log.Print("using Vault role in utilized AWS authentication engine with the same name as the current utilized AWS IAM Role")
		}
	}

	return &vaultConfig{
		engine:       authEngine,
		address:      source.Address,
		awsMountPath: source.AWSMountPath,
		awsRole:      source.AWSVaultRole,
		token:        source.Token,
		insecure:     source.Insecure,
	}, nil
}

// instantiate authenticated vault client with aws-iam or token auth
func NewClient(config *vaultConfig) (*vault.Client, error) {
	// initialize config
	vaultConfig := &vault.Config{Address: config.address}
	if err := vaultConfig.ConfigureTLS(&vault.TLSConfig{Insecure: config.insecure}); err != nil {
		log.Print("Vault TLS configuration failed to initialize")
		return nil, err
	}

	// initialize client
	client, err := vault.NewClient(vaultConfig)
	if err != nil {
		log.Print("Vault client failed to initialize")
		return nil, err
	}

	// verify vault is unsealed
	sealStatus, err := client.Sys().SealStatus()
	if err != nil {
		log.Print("unable to verify that the Vault cluster is unsealed")
		return nil, err
	}
	if sealStatus.Sealed {
		log.Print("the Vault server cluster is sealed and no operations can be executed")
		return nil, errors.New("vault sealed")
	}

	// determine authentication method
	switch config.engine {
	case token:
		client.SetToken(config.token)
	case awsIam:
		// determine iam role login option
		var loginOption auth.LoginOption

		if len(config.awsRole) > 0 {
			// use explicitly specified iam role
			loginOption = auth.WithRole(config.awsRole)
		} else {
			// use default iam role
			loginOption = auth.WithIAMAuth()
		}
		// authenticate with aws iam
		awsAuth, err := auth.NewAWSAuth(loginOption)
		if err != nil {
			log.Print("unable to initialize AWS IAM authentication")
			return nil, err
		}

		authInfo, err := client.Auth().Login(context.Background(), awsAuth)
		if err != nil {
			log.Print("unable to login with AWS IAM auth method")
			return nil, err
		}
		if authInfo == nil {
			log.Print("no auth info was returned after login")
			return nil, errors.New("no auth info")
		}
	}

	// return authenticated vault client
	return client, nil
}
