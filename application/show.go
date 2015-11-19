package application

func (a *Application) Show(stackName string) error {
	val, err := a.ConfigStore.Get("ssh-key")
	if err != nil {
		return err
	}
	_, err = a.ResultWriter.Write(val)
	if err != nil {
		return err
	}
	return nil
}
