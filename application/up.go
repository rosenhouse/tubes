package application

import (
	"fmt"
	"regexp"

	"github.com/rosenhouse/tubes/lib/awsclient"
)

const StackNamePattern = `^[a-zA-Z][-a-zA-Z0-9]*$`

func (a *Application) Boot(stackName string) error {
	regex := regexp.MustCompile(StackNamePattern)
	if !regex.MatchString(stackName) {
		return fmt.Errorf("invalid name: must match pattern %s", StackNamePattern)
	}

	a.Logger.Printf("Creating keypair...")
	pemBytes, err := a.AWSClient.CreateKeyPair(stackName)
	if err != nil {
		return err
	}

	err = a.ConfigStore.Set(fmt.Sprintf("%s/%s", stackName, "ssh-key"), []byte(pemBytes))
	if err != nil {
		return err
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

	err = a.AWSClient.WaitForStack(stackName, awsclient.CloudFormationUpsertPundit{})
	if err != nil {
		return err
	}
	a.Logger.Println("Stack update complete")

	baseStackResources, err := a.AWSClient.GetBaseStackResources(stackName)
	if err != nil {
		return err
	}

	manifestYAML, err := a.ManifestBuilder.Build(stackName, baseStackResources)
	if err != nil {
		return err
	}

	err = a.ConfigStore.Set(fmt.Sprintf("%s/%s", stackName, "director.yml"), manifestYAML)
	if err != nil {
		return err
	}

	a.Logger.Println("Finished")
	return nil
}
