package client

type Config struct {
	Nodes []string
	Test  bool
}

func (c *Config) AddConfig(config Config) {
	c.Nodes = config.Nodes
	c.Test = config.Test
}
