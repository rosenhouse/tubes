package application

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rosenhouse/tubes/lib/awsclient"
)

const StackNamePattern = `^[a-zA-Z][-a-zA-Z0-9]*$`

type ConcourseCredentials struct {
	DBPassword string
}

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
	a.Logger.Println("Upserting base stack.  Check CloudFormation console for details.")
	err = a.AWSClient.UpsertStack(stackName+"-base", templateJSON, parameters)
	if err != nil {
		return err
	}

	err = a.AWSClient.WaitForStack(stackName+"-base", awsclient.CloudFormationUpsertPundit{})
	if err != nil {
		return err
	}
	a.Logger.Println("Stack update complete")
	a.Logger.Println("Retrieving resource ids")

	baseStackResources, err := a.AWSClient.GetBaseStackResources(stackName + "-base")
	if err != nil {
		return err
	}

	err = a.ConfigStore.Set("bosh-ip", []byte(baseStackResources.BOSHElasticIP))
	if err != nil {
		return err
	}

	err = a.ConfigStore.Set("nat-ip", []byte(baseStackResources.NATElasticIP))
	if err != nil {
		return err
	}

	a.Logger.Println("Generating BOSH init manifest")

	accessKey, secretKey, err := a.AWSClient.CreateAccessKey(baseStackResources.BOSHUser)
	if err != nil {
		return err
	}

	manifestYAML, boshPassword, err := a.ManifestBuilder.Build(stackName, baseStackResources, accessKey, secretKey)
	if err != nil {
		return err
	}

	boshEnvLines := []string{
		fmt.Sprintf(`export BOSH_TARGET="%s"`, baseStackResources.BOSHElasticIP),
		fmt.Sprintf(`export BOSH_USER="%s"`, "admin"),
		fmt.Sprintf(`export BOSH_PASSWORD="%s"`, boshPassword),
		fmt.Sprintf(`export NAT_IP="%s"`, baseStackResources.NATElasticIP),
	}
	err = a.ConfigStore.Set("bosh-environment", []byte(strings.Join(boshEnvLines, "\n")))
	if err != nil {
		return err
	}

	err = a.ConfigStore.Set("director.yml", manifestYAML)
	if err != nil {
		return err
	}

	err = a.ConfigStore.Set("bosh-password", []byte(boshPassword))
	if err != nil {
		return err
	}

	a.Logger.Println("Downloading the concourse manifest from " + a.ConcourseTemplateURL)

	concourseManifestYAMLTemplate, err := a.HTTPClient.Get(a.ConcourseTemplateURL)
	if err != nil {
		return err
	}

	concourseTemplateJSON := awsclient.ConcourseStackTemplate.String()
	a.Logger.Println("Upserting Concourse stack.  Check CloudFormation console for details.")
	err = a.AWSClient.UpsertStack(
		stackName+"-concourse", concourseTemplateJSON, map[string]string{
			"VPCID":                    baseStackResources.VPCID,
			"NATInstance":              baseStackResources.NATInstanceID,
			"PubliclyRoutableSubnetID": baseStackResources.BOSHSubnetID,
			"AvailabilityZone":         baseStackResources.AvailabilityZone,
		})
	if err != nil {
		return err
	}

	err = a.AWSClient.WaitForStack(stackName+"-concourse", awsclient.CloudFormationUpsertPundit{})
	if err != nil {
		return err
	}
	a.Logger.Println("Stack update complete")
	a.Logger.Println("Retrieving resource ids")
	concourseStackResources, err := a.AWSClient.GetStackResources(stackName + "-concourse")
	if err != nil {
		panic(err)
	}

	a.Logger.Println("Generating the concourse manifest")

	concourseCredentials := ConcourseCredentials{}
	err = a.CredentialsGenerator.Fill(&concourseCredentials)
	if err != nil {
		return err
	}

	filledInConcourseTemplate := strings.Replace(
		string(concourseManifestYAMLTemplate),
		"REPLACE_WITH_AVAILABILITY_ZONE",
		baseStackResources.AvailabilityZone,
		-1)

	filledInConcourseTemplate = strings.Replace(filledInConcourseTemplate,
		"REPLACE_WITH_DB_PASSWORD",
		concourseCredentials.DBPassword,
		-1)

	filledInConcourseTemplate = strings.Replace(filledInConcourseTemplate,
		"REPLACE_WITH_INTERNAL_SECURITY_GROUP_NAME",
		concourseStackResources["ConcourseSecurityGroup"],
		-1)

	filledInConcourseTemplate = strings.Replace(filledInConcourseTemplate,
		"REPLACE_WITH_INTERNAL_SUBNET",
		concourseStackResources["ConcourseSubnet"],
		-1)

	filledInConcourseTemplate = strings.Replace(filledInConcourseTemplate,
		"REPLACE_WITH_WEB_ELB_NAME",
		concourseStackResources["LoadBalancer"],
		-1)

	filledInConcourseTemplate = strings.Replace(filledInConcourseTemplate,
		"REPLACE_WITH_UUID",
		"YOUR_DIRECTOR_UUID_HERE",
		-1)

	err = a.ConfigStore.Set("concourse.yml", []byte(filledInConcourseTemplate))
	if err != nil {
		return err
	}

	a.Logger.Println("Finished")
	return nil
}
