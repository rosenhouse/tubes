package commands

import "time"

type CLIOptions struct {
	Name      string    `short:"n" long:"name"  description:"Name of environment to manipulate"`
	AWSConfig AWSConfig `group:"aws"`
	StateDir  string    `short:"s" long:"state-dir" description:"Path to directory where state is stored.  Typically you'd track this in a private git repository or other secure location.  Defaults to <working_dir>/environments/<name>"`

	BoshIOURL string `long:"bosh-io-url" default:"https://bosh.io" env:"TUBES_BOSH_IO_URL" description:"URL of BOSH hub.  Override for testing."`

	Up   Up   `command:"up" description:"Boot a new environment with the given name"`
	Down Down `command:"down" description:"Tear down the named environment"`
	Show Show `command:"show" description:"Show information about the named environment"`
}

type Up struct {
	*CLIOptions `no-flag:"true"`
}

type Down struct {
	*CLIOptions `no-flag:"true"`
}

type Show struct {
	*CLIOptions `no-flag:"true"`

	SSHKey          bool `long:"ssh" description:"print the SSH key needed to login to the VMs instances"`
	BoshIP          bool `long:"bosh-ip" description:"print the IP address of the BOSH director"`
	BoshPassword    bool `long:"bosh-password" description:"print the admin password for the BOSH director"`
	BoshEnvironment bool `long:"bosh-environment" description:"print the BOSH environment variables, suitable for sourcing in bash"`
}

type AWSConfig struct {
	Region    string `long:"aws-region" env:"AWS_DEFAULT_REGION" description:"defaults to"`
	AccessKey string `long:"aws-access-key" env:"AWS_ACCESS_KEY_ID" description:"defaults to"`
	SecretKey string `long:"aws-secret-key" env:"AWS_SECRET_ACCESS_KEY" description:"defaults to"`

	EndpointOverrides string        `long:"endpoint-overrides" env:"TUBES_AWS_ENDPOINTS" description:"JSON hash of AWS endpoint URLs.  Override for testing."`
	StackWaitTimeout  time.Duration `long:"stack-wait-timeout" default:"7m" description:"maximum time to wait for CloudFormation stack changes"`
}

func New() *CLIOptions {
	base := &CLIOptions{}
	base.Up.CLIOptions = base
	base.Down.CLIOptions = base
	base.Show.CLIOptions = base

	return base
}
