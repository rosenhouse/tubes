package main

import (
	"log"
	"os"

	"github.com/rosenhouse/tubes/application"
	"github.com/rosenhouse/tubes/lib/awsclient"
)

func main() {
	logger := log.New(os.Stderr, "", 0)

	var stackName string
	if len(os.Args) < 2 {
		logger.Fatalln("expecting stack name")
	}
	stackName = os.Args[1]

	load := func(name string) string {
		val := os.Getenv(name)
		if val == "" {
			logger.Fatalf("missing required environment variable %s", name)
		}
		return val
	}

	awsClient := awsclient.New(awsclient.Config{
		Region:    load("AWS_DEFAULT_REGION"),
		AccessKey: load("AWS_ACCESS_KEY_ID"),
		SecretKey: load("AWS_SECRET_ACCESS_KEY"),
	})

	app := application.Application{
		AWSClient: awsClient,
		Logger:    logger,
	}

	err := app.Boot(stackName)
	if err != nil {
		logger.Fatalf("%s", err)
	}
}
