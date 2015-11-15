package commands

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/jessevdk/go-flags"
	"github.com/rosenhouse/tubes/application"
	"github.com/rosenhouse/tubes/lib/awsclient"
)

const StackNamePattern = `^[a-zA-Z][-a-zA-Z0-9]*$`

func parseError(fmtString string, args ...interface{}) *flags.Error {
	return &flags.Error{Message: fmt.Sprintf(fmtString, args...)}
}

func (c *CLIOptions) checkStackName() error {
	name := c.Name
	if name == "" {
		return parseError("missing required flag name")
	}

	regex := regexp.MustCompile(StackNamePattern)
	if !regex.MatchString(name) {
		return parseError("invalid name: must match pattern %s", StackNamePattern)
	}

	return nil
}

func (c *AWSConfig) buildClient() (*awsclient.Client, error) {
	var missing bool
	load := func(val string) string {
		if val == "" {
			missing = true
		}
		return val
	}
	config := awsclient.Config{
		Region:    load(c.Region),
		AccessKey: load(c.AccessKey),
		SecretKey: load(c.SecretKey),
	}

	if missing {
		return nil, parseError("missing one or more AWS config options/env vars")
	}
	return awsclient.New(config), nil
}

func (options *CLIOptions) initApp(args []string) (*application.Application, error) {
	if options == nil {
		return nil, errors.New("programming error: missing parent reference in command")
	}
	if len(args) > 0 {
		return nil, parseError("unknown args: %+v\n", args)
	}
	if err := options.checkStackName(); err != nil {
		return nil, err
	}

	awsClient, err := options.AWSConfig.buildClient()
	if err != nil {
		return nil, err
	}

	return &application.Application{
		AWSClient: awsClient,
		Logger:    log.New(os.Stderr, "", 0),
	}, nil
}
