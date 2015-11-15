package application

import "github.com/rosenhouse/tubes/lib/awsclient"

func (a *Application) Destroy(stackName string) error {
	a.Logger.Println("Deleting stack")
	err := a.AWSClient.DeleteStack(stackName)
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
