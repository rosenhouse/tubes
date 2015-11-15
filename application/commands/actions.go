package commands

func (c *Up) Execute(args []string) error {
	app, err := c.initApp(args)
	if err != nil {
		return err
	}

	return app.Boot(c.Name)
}

func (c *Down) Execute(args []string) error {
	app, err := c.initApp(args)
	if err != nil {
		return err
	}

	return app.Destroy(c.Name)
}

func (c *Show) Execute(args []string) error {
	app, err := c.initApp(args)
	if err != nil {
		return err
	}
	return app.Show(c.Name)
}
