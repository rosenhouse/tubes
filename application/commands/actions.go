package commands

import "github.com/rosenhouse/tubes/application"

func (c *Up) Execute(args []string) error {
	app, err := c.InitApp(args)
	if err != nil {
		return err
	}

	return app.Boot(c.Name)
}

func (c *Down) Execute(args []string) error {
	app, err := c.InitApp(args)
	if err != nil {
		return err
	}

	return app.Destroy(c.Name)
}

func (c *Show) Execute(args []string) error {
	app, err := c.InitApp(args)
	if err != nil {
		return err
	}
	return app.Show(c.Name, application.ShowOptions{
		SSHKey:          c.SSHKey,
		BoshIP:          c.BoshIP,
		BoshPassword:    c.BoshPassword,
		BoshEnvironment: c.BoshEnvironment,
	})
}
