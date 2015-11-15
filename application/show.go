package application

import "fmt"

func (a *Application) Show(stackName string) error {
	val, err := a.ConfigStore.Get(fmt.Sprintf("%s/%s", stackName, "ssh-key"))
	if err != nil {
		return err
	}
	_, err = a.ResultWriter.Write(val)
	if err != nil {
		return err
	}
	return nil
}
