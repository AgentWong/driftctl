package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/terraform"
	tf "github.com/snyk/driftctl/enumeration/terraform"
)

type awsProviderConfig struct {
	AccessKey     string
	SecretKey     string
	CredsFilename string
	Profile       string
	Token         string
	Region        string `cty:"region"`
	MaxRetries    int

	AssumeRoleARN         string
	AssumeRoleExternalID  string
	AssumeRoleSessionName string
	AssumeRolePolicy      string

	AllowedAccountIDs   []string
	ForbiddenAccountIDs []string

	Endpoints        map[string]string
	IgnoreTagsConfig map[string]string
	Insecure         bool

	SkipCredsValidation     bool `cty:"skip_credentials_validation"`
	SkipGetEC2Platforms     bool
	SkipRegionValidation    bool
	SkipRequestingAccountID bool `cty:"skip_requesting_account_id"`
	SkipMetadataAPICheck    bool
	S3ForcePathStyle        bool
}

// TerraformProvider is the AWS-specific Terraform provider implementation.
type TerraformProvider struct {
	*terraform.Provider
	AwsCfg    aws.Config
	name      string
	version   string
	accountID string
}

// NewTerraformProvider creates and configures a new AWS Terraform provider.
func NewTerraformProvider(version string, progress enumeration.ProgressCounter, configDir string) (*TerraformProvider, error) {
	if version == "" {
		version = "6.38.0"
	}
	p := &TerraformProvider{
		version: version,
		name:    "aws",
	}
	installer, err := tf.NewProviderInstaller(tf.ProviderConfig{
		Key:       p.name,
		Version:   version,
		ConfigDir: configDir,
	})
	if err != nil {
		return nil, err
	}

	p.AwsCfg, err = awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	tfProvider, err := terraform.NewProvider(installer, terraform.ProviderConfig{
		Name:         p.name,
		DefaultAlias: p.AwsCfg.Region,
		GetProviderConfig: func(alias string) interface{} {
			return awsProviderConfig{
				Region: alias,
				// Those two parameters are used to make sure that the credentials are not validated when calling
				// Configure(). Credentials validation is now handled directly in driftctl
				SkipCredsValidation:     true,
				SkipRequestingAccountID: true,
				MaxRetries:              10, // TODO make this configurable
			}
		},
	}, progress)
	if err != nil {
		return nil, err
	}
	p.Provider = tfProvider
	return p, err
}

// Name returns the provider name.
func (a *TerraformProvider) Name() string {
	return a.name
}

// Version returns the provider version.
func (a *TerraformProvider) Version() string {
	return a.version
}

// ErrAWSCredentialsNotFound is returned when no valid AWS credentials are found.
var ErrAWSCredentialsNotFound = errors.New("Could not find a way to authenticate on AWS!\n" +
	"Please refer to AWS documentation: https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html")

// CheckCredentialsExist verifies that valid AWS credentials are available.
func (a *TerraformProvider) CheckCredentialsExist() error {
	creds, err := a.AwsCfg.Credentials.Retrieve(context.Background())
	if err != nil {
		return ErrAWSCredentialsNotFound
	}
	if !creds.HasKeys() {
		return ErrAWSCredentialsNotFound
	}
	// This call is to make sure that the credentials are valid
	// A more complex logic exist in terraform provider, but it's probably not worth to implement it
	stsClient := sts.NewFromConfig(a.AwsCfg)
	identity, err := stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		logrus.Debug(err)
		return errors.New("Could not authenticate successfully on AWS with the provided credentials.\n" +
			"Please refer to the AWS documentation: https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html\n")
	}

	a.accountID = aws.ToString(identity.Account)
	return nil
}
