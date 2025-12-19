package vault

import (
	"context"
	"errors"
	"log"
	"net/url"
	"regexp"
	"strings"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/aws"
	"github.com/mschuchard/concourse-vault-resource/concourse"
	"github.com/mschuchard/concourse-vault-resource/enum"
)

// configured vault client validated constructor
func NewVaultClient(source concourse.Source) (*vault.Client, error) {
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
	if !source.Insecure && strings.HasPrefix(source.Address, "http:") {
		log.Print("insecure input parameter was omitted or specified as false, and address protocol is http")
		log.Print("insecure will be reset to value of true")
		source.Insecure = true
	}

	// initialize vault api config
	vaultConfig := &vault.Config{Address: source.Address}
	if err := vaultConfig.ConfigureTLS(&vault.TLSConfig{Insecure: source.Insecure}); err != nil {
		log.Print("Vault TLS configuration failed to initialize")
		return nil, err
	}

	// initialize vault client
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

	// authenticate vault client
	if err := authClient(source, client); err != nil {
		log.Print("unable to authenticate Vault client")
		return nil, err
	}

	// return authenticated vault client
	return client, nil
}

// determine authentication method and authenticate client
func authClient(source concourse.Source, client *vault.Client) error {
	// initialize locals
	token := source.Token
	awsMountPath := source.AWSMountPath
	awsRole := source.AWSVaultRole
	engine, err := source.AuthEngine.New()

	// determine vault auth engine if unspecified
	if len(source.AuthEngine) == 0 {
		log.Print("authentication engine for Vault not specified, or specified but unsupported")
		log.Print("using logic from other input parameters to assist with determination")

		// validate inputs specified for only one engine
		if len(token) > 0 && (len(awsMountPath) > 0 || len(awsRole) > 0) {
			log.Print("token and AWS mount path or AWS role were simultaneously specified; these are mutually exclusive options")
			log.Print("intended authentication engine could not be determined from other parameters")
			return errors.New("unable to deduce authentication engine")
		}
		if len(token) == 0 {
			log.Print("AWS IAM authentication will be utilized with the Vault client")
			engine = enum.AWSIAM
		} else {
			log.Print("token authentication will be utilized with the Vault client")
			engine = enum.VaultToken
		}
	} else if err != nil { // return error if invalid engine was specified
		return err
	}

	// determine vault authentication method
	switch engine {
	case enum.VaultToken:
		// validate vault token
		if matched, _ := regexp.MatchString(`^[a-zA-Z0-9.]+$`, token); !matched {
			log.Print("the specified Vault Token is invalid")
			return errors.New("invalid vault token")
		}

		// authenticate with token
		client.SetToken(token)
	case enum.AWSIAM:
		// default aws mount path
		if len(awsMountPath) == 0 {
			log.Print("using default AWS authentication mount path at 'aws'")
			awsMountPath = "aws"
		}
		mountLoginOption := auth.WithMountPath(awsMountPath)

		// determine iam role login option
		var roleLoginOption auth.LoginOption

		if len(awsRole) > 0 {
			// use explicitly specified aws role
			log.Printf("using Vault AWS role %s for authentication", awsRole)
			roleLoginOption = auth.WithRole(awsRole)
		} else {
			// use default aws iam role (i.e. instance profile)
			log.Print("using Vault role in utilized AWS authentication engine with the same name as the currently utilized AWS IAM Role")
			roleLoginOption = auth.WithIAMAuth()
		}

		// authenticate with aws iam
		awsAuth, err := auth.NewAWSAuth(roleLoginOption, mountLoginOption)
		if err != nil {
			log.Print("unable to initialize Vault AWS IAM authentication")
			return err
		}

		// utilize aws authentication with vault client
		authInfo, err := client.Auth().Login(context.Background(), awsAuth)
		if err != nil {
			log.Print("unable to authenticate to Vault via AWS IAM auth method")
			return err
		}
		if authInfo == nil {
			return errors.New("no auth info was returned after login")
		}
	default:
		log.Printf("%s was input as an authentication engine, but only token and aws are supported", engine)
		return errors.New("invalid Vault authentication engine")
	}

	return nil
}
