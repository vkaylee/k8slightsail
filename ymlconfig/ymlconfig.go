package ymlconfig

import (
	"k8s_lightsail/utils"
	"log"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ConfigFile	*utils.File
	Template	*T
}

type T struct {
	Region				*string		`yaml:"region"`
	Nodes				[]*Node		`yaml:"nodes"`
	SSHKeyPairName		*string		`yaml:"ssh_key_pair_name"`
	SSHPrivateKeyStr	*string		`yaml:"ssh_private_key"`
	SSHPublicKeyStr		*string		`yaml:"ssh_public_key"`
}
type Node struct {
	Name		*string	`yaml:"name"`
	Type 		*string	`yaml:"type"`
	Zone		*string	`yaml:"zone"`
	Username	*string	`yaml:"username"`
	PublicIp	*string	`yaml:"public_ip"`
	PrivateIp	*string	`yaml:"private_ip"`
}

func (c *Config) ReadConfig() error {
	err := c.ConfigFile.Loadfile()
	if err != nil {
		return err
	}
	initTem := T{}
	err = yaml.Unmarshal([]byte(*c.ConfigFile.GetContents()), &initTem)
	if err != nil {
		return err
	}
	c.Template = &initTem
	return nil
}
func (c *Config) WriteConfig() {
	d, err := yaml.Marshal(c.Template)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	c.ConfigFile.SetContents(utils.String(string(d)))
	c.ConfigFile.WriteFile()
}

