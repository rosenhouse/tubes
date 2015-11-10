package application

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

type Application struct {
	AWSClient         awsClient
	BaseStackTemplate string
	SSHKeyName        string
	Logger            logger
}

func (a *Application) Boot(stackName string) error {
	a.Logger.Println("Looking for latest AWS NAT box AMI...")
	natInstanceAMI, err := a.AWSClient.GetLatestNATBoxAMIID()
	if err != nil {
		return err
	}
	a.Logger.Printf("Latest NAT box AMI is %q\n", natInstanceAMI)

	parameters := map[string]string{
		"NATInstanceAMI": natInstanceAMI,
		"KeyName":        a.SSHKeyName,
	}
	a.Logger.Println("Upserting stack...")
	err = a.AWSClient.UpsertStack(stackName, a.BaseStackTemplate, parameters)
	if err != nil {
		return err
	}

	return a.AWSClient.WaitForStack(stackName)
}
