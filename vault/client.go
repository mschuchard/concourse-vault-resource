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
	"github.com/hashicorp/vault/api/auth/kubernetes"

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
	authMount := source.AuthMount
	vaultRole := source.VaultRole
	engine, err := source.AuthEngine.New()
	if err != nil {
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
	case enum.KubernetesSA:
		// default kubernetes mount path
		if len(authMount) == 0 {
			log.Print("using default Kubernetes authentication mount path at 'kubernetes'")
			authMount = "kubernetes"
		}

		// validate kubernetes vault role input
		if len(vaultRole) == 0 {
			log.Print("a Kubernetes Vault role must be specified for the Kubernetes authentication method")
			return errors.New("no kubernetes vault role specified")
		}

		// authenticate with kubernetes service account
		kubeAuth, err := kubernetes.NewKubernetesAuth(
			vaultRole,
			kubernetes.WithMountPath(authMount),
		)
		if err != nil {
			log.Print("unable to initialize Kubernetes service account authentication")
			return err
		}

		authInfo, err := client.Auth().Login(context.Background(), kubeAuth)
		if err != nil {
			log.Print("unable to authenticate to Vault via Kubernetes service account method")
			return err
		}
		if authInfo == nil {
			return errors.New("no auth info was returned after login")
		}
	case enum.AWSIAM:
		// default aws mount path
		if len(authMount) == 0 {
			log.Print("using default AWS authentication mount path at 'aws'")
			authMount = "aws"
		}
		mountLoginOption := auth.WithMountPath(authMount)

		// determine iam role login option
		var roleLoginOption auth.LoginOption

		if len(vaultRole) > 0 {
			// use explicitly specified aws role
			log.Printf("using Vault AWS role %s for authentication", vaultRole)
			roleLoginOption = auth.WithRole(vaultRole)
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
		log.Printf("%s was input as the authentication engine, but it is not currently supported", engine)
		return errors.New("invalid Vault authentication engine")
	}

	return nil
}
