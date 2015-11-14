package application

import (
	"fmt"
	"regexp"

	"github.com/rosenhouse/tubes/lib/awsclient"
)

type awsClient interface {
	GetLatestNATBoxAMIID() (string, error)
	UpsertStack(stackName string, template string, parameters map[string]string) error
	WaitForStack(stackName string) error
}

type logger interface {
	Printf(format string, v ...interface{})
	Println(a ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(a ...interface{})
}

type state interface {
}

type Application struct {
	AWSClient awsClient
	Logger    logger
}

const StackNamePattern = `^[a-zA-Z][-a-zA-Z0-9]*$`

func (a *Application) Boot(stackName string) error {
	regex := regexp.MustCompile(StackNamePattern)
	if !regex.MatchString(stackName) {
		return fmt.Errorf("invalid name: must match pattern %s", StackNamePattern)
	}

	a.Logger.Println("Looking for latest AWS NAT box AMI...")
	natInstanceAMI, err := a.AWSClient.GetLatestNATBoxAMIID()
	if err != nil {
		return err
	}
	a.Logger.Printf("Latest NAT box AMI is %q\n", natInstanceAMI)

	parameters := map[string]string{
		"NATInstanceAMI": natInstanceAMI,
		"KeyName":        stackName,
	}
	templateJSON := awsclient.BaseStackTemplate.String()
	a.Logger.Println("Upserting stack...")
	err = a.AWSClient.UpsertStack(stackName, templateJSON, parameters)
	if err != nil {
		return err
	}

	err = a.AWSClient.WaitForStack(stackName)
	if err != nil {
		return err
	}

	a.Logger.Println("Finished")
	return nil
}
