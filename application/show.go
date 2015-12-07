package application

import "fmt"

type ShowOptions struct {
	SSHKey bool
	BoshIP bool
}

func (a *Application) Show(stackName string, options ShowOptions) error {
	if (options == ShowOptions{}) {
		return fmt.Errorf("set at least one flag")
	}

	if options.SSHKey {
		val, err := a.ConfigStore.Get("ssh-key")
		if err != nil {
			return err
		}
		_, err = a.ResultWriter.Write(val)
		if err != nil {
			return err
		}
	}

	if options.BoshIP {
		val, err := a.ConfigStore.Get("bosh-ip")
		if err != nil {
			return err
		}
		val = append(val, byte('\n'))
		_, err = a.ResultWriter.Write(val)
		if err != nil {
			return err
		}
	}
	return nil
}
