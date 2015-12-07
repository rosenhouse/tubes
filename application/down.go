package application

import "github.com/rosenhouse/tubes/lib/awsclient"

func (a *Application) Destroy(stackName string) error {
	a.Logger.Println("Inspecting stack")
	resources, err := a.AWSClient.GetBaseStackResources(stackName)
	if err != nil {
		return err
	}

	a.Logger.Println("Inspecting user")
	accessKeys, err := a.AWSClient.ListAccessKeys(resources.BOSHUser)
	if err != nil {
		return err
	}

	a.Logger.Println("Deleting access keys")
	for _, accessKey := range accessKeys {
		err = a.AWSClient.DeleteAccessKey(resources.BOSHUser, accessKey)
		if err != nil {
			return err
		}
	}

	a.Logger.Println("Deleting Concourse stack")
	err = a.AWSClient.DeleteStack(stackName + "-concourse")
	if err != nil {
		return err
	}

	err = a.AWSClient.WaitForStack(stackName+"-concourse", awsclient.CloudFormationDeletePundit{})
	if err != nil {
		return err
	}
	a.Logger.Printf("Delete complete")

	a.Logger.Println("Deleting base stack")
	err = a.AWSClient.DeleteStack(stackName)
	if err != nil {
		return err
	}

	err = a.AWSClient.WaitForStack(stackName, awsclient.CloudFormationDeletePundit{})
	if err != nil {
		return err
	}
	a.Logger.Printf("Delete complete")
	a.Logger.Printf("Deleting keypair...")
	err = a.AWSClient.DeleteKeyPair(stackName)
	if err != nil {
		return err
	}

	a.Logger.Println("Finished")
	return nil
}
