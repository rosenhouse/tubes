package application

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rosenhouse/tubes/lib/awsclient"
)

const StackNamePattern = `^[a-zA-Z][-a-zA-Z0-9]*$`

func (a *Application) Boot(stackName string) error {
	regex := regexp.MustCompile(StackNamePattern)
	if !regex.MatchString(stackName) {
		return fmt.Errorf("invalid name: must match pattern %s", StackNamePattern)
	}

	emptyConfigStore, err := a.ConfigStore.IsEmpty()
	if err != nil {
		return err
	}
	if !emptyConfigStore {
		return fmt.Errorf("state directory must be empty")
	}

	a.Logger.Printf("Creating keypair...")
	pemBytes, err := a.AWSClient.CreateKeyPair(stackName)
	if err != nil {
		return err
	}

	err = a.ConfigStore.Set("ssh-key", []byte(pemBytes))
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
	a.Logger.Println("Retrieving resource ids")

	baseStackResources, err := a.AWSClient.GetBaseStackResources(stackName)
	if err != nil {
		return err
	}
	a.Logger.Println("Generating BOSH init manifest")

	accessKey, secretKey, err := a.AWSClient.CreateAccessKey(baseStackResources.BOSHUser)
	if err != nil {
		return err
	}

	manifestYAML, err := a.ManifestBuilder.Build(stackName, baseStackResources, accessKey, secretKey)
	if err != nil {
		return err
	}

	err = a.ConfigStore.Set("director.yml", manifestYAML)
	if err != nil {
		return err
	}

	a.Logger.Println("Downloading the concourse manifest from " + a.ConcourseTemplateURL)

	concourseManifestYAMLTemplate, err := a.HTTPClient.Get(a.ConcourseTemplateURL)
	if err != nil {
		return err
	}

	a.Logger.Println("Generating the concourse manifest")

	filledInConcourseTemplate := strings.Replace(
		string(concourseManifestYAMLTemplate),
		"REPLACE_WITH_AVAILABILITY_ZONE",
		baseStackResources.AWSRegion,
		-1)

	err = a.ConfigStore.Set("concourse.yml", []byte(filledInConcourseTemplate))
	if err != nil {
		return err
	}

	a.Logger.Println("Finished")
	return nil
}
