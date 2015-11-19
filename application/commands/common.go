package commands

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/rosenhouse/tubes/application"
	"github.com/rosenhouse/tubes/lib/awsclient"
	"github.com/rosenhouse/tubes/lib/boshio"
	"github.com/rosenhouse/tubes/lib/credentials"
	"github.com/rosenhouse/tubes/lib/director"
)

func parseError(fmtString string, args ...interface{}) *flags.Error {
	return &flags.Error{Message: fmt.Sprintf(fmtString, args...)}
}

func (c *CLIOptions) checkStackName() error {
	name := c.Name
	if name == "" {
		return parseError("missing required flag name")
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

func (options *CLIOptions) InitApp(args []string) (*application.Application, error) {
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

	stateDir := options.StateDir
	if stateDir != "" {
		fileInfo, err := os.Stat(stateDir)
		if err != nil {
			return nil, fmt.Errorf("state directory not found: %s", err)
		}
		if !fileInfo.IsDir() {
			return nil, fmt.Errorf("state directory not a directory: %s", stateDir)
		}

	} else {
		workingDir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		stateDir = filepath.Join(workingDir, "environments", options.Name)
		err = os.MkdirAll(stateDir, 0700)
		if err != nil {
			return nil, err
		}
	}

	configStore := &application.FilesystemConfigStore{RootDir: stateDir}

	httpClient := &boshio.HTTPClient{
		BaseURL: "https://bosh.io",
	}

	return &application.Application{
		AWSClient:    awsClient,
		Logger:       log.New(os.Stderr, "", 0),
		ResultWriter: os.Stdout,
		ConfigStore:  configStore,
		ManifestBuilder: &application.ManifestBuilder{
			DirectorManifestGenerator: director.DirectorManifestGenerator{},
			BoshIOClient: &boshio.Client{
				JSONClient: &boshio.JSONClient{httpClient},
				HTTPClient: httpClient,
			},
			CredentialsGenerator: credentials.Generator{Length: 12},
		},
	}, nil
}
