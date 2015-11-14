package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/rosenhouse/tubes/application"
	"github.com/rosenhouse/tubes/lib/awsclient"
)

func main() {
	logger := log.New(os.Stdout, "", 0)

	if len(os.Args) != 3 {
		_, programName := filepath.Split(os.Args[0])
		logger.Fatalf("usage: %s action stack-name", programName)
	}

	action := os.Args[1]
	stackName := os.Args[2]

	if action != "up" {
		logger.Fatalf("invalid action %q", action)
	}

	awsConfig, err := loadAWSConfigFromEnv()
	if err != nil {
		logger.Fatalf("%s", err)
	}

	awsClient := awsclient.New(awsConfig)

	app := application.Application{
		AWSClient: awsClient,
		Logger:    logger,
	}

	err = app.Boot(stackName)
	if err != nil {
		logger.Fatalf("%s", err)
	}
}

func loadAWSConfigFromEnv() (awsclient.Config, error) {
	missing := []string{}
	load := func(name string) string {
		val := os.Getenv(name)
		if val == "" {
			missing = append(missing, name)
		}
		return val
	}
	config := awsclient.Config{
		Region:    load("AWS_DEFAULT_REGION"),
		AccessKey: load("AWS_ACCESS_KEY_ID"),
		SecretKey: load("AWS_SECRET_ACCESS_KEY"),
	}

	if len(missing) > 0 {
		return config, fmt.Errorf("missing required environment variable(s): %s", missing)
	}
	return config, nil
}
