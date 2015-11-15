package commands

type CLIOptions struct {
	Name      string    `short:"n" long:"name"  description:"Name of environment to manipulate"`
	AWSConfig AWSConfig `group:"aws"`

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
	SSHKey      bool `short:"k" long:"ssh-key" description:"print the SSH key to stdout"`
}

type AWSConfig struct {
	Region    string `long:"aws-region" env:"AWS_DEFAULT_REGION" description:"defaults to"`
	AccessKey string `long:"aws-access-key" env:"AWS_ACCESS_KEY_ID" description:"defaults to"`
	SecretKey string `long:"aws-secret-key" env:"AWS_SECRET_ACCESS_KEY" description:"defaults to"`
}

func New() *CLIOptions {
	base := &CLIOptions{}
	base.Up.CLIOptions = base
	base.Down.CLIOptions = base
	base.Show.CLIOptions = base

	return base
}
