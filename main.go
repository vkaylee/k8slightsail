package main

import (
	"context"
	"fmt"
	"k8s_lightsail/app"
	"k8s_lightsail/config"
	"k8s_lightsail/utils"
	"k8s_lightsail/ymlconfig"
	"log"
	"math/rand"
	"os"
	"strings"
)


func main() {

	a := app.App{
		Ctx: context.Background(),
		AppConf: &config.Config{
			KeyPairName:    utils.String(""),
			PublicKeyStr:   utils.String(""),
			PrivateKeyStr:  utils.String(""),
		},
		YmlConf: &ymlconfig.Config{
			ConfigFile: utils.InitFileWithPath(utils.String("config.yml")),
			Template: &ymlconfig.T{
				SSHKeyPairName:   utils.String(""),
				SSHPrivateKeyStr: utils.String(""),
				SSHPublicKeyStr:  utils.String(""),
			},
		},
	}

	if len(os.Args) < 2 {
		fmt.Printf("expected %s or %s subcommands", config.UpActionStr, config.DownActionStr)
		os.Exit(1)
	}
	/**
		Read yml file config first: config.yml
	 */
	isValidYmlConfig := true

	if err := a.YmlConf.ReadConfig(); err != nil {
		isValidYmlConfig = false
	}
	/**
		If the config file doest have SSH key
	 */
	if strings.EqualFold(*a.YmlConf.Template.SSHPublicKeyStr,"") || strings.EqualFold(*a.YmlConf.Template.SSHPrivateKeyStr,"") {
		// Gen ssh KEY
		pubKey, privKey, err := utils.MakeSSHKeyPair()
		if err != nil {
			log.Println(err)
		}
		a.YmlConf.Template.SSHPrivateKeyStr = privKey
		a.YmlConf.Template.SSHPublicKeyStr = pubKey
	}
	// Copy SSH key in YAML to app config
	a.AppConf.PrivateKeyStr = a.YmlConf.Template.SSHPrivateKeyStr
	a.AppConf.PublicKeyStr = a.YmlConf.Template.SSHPublicKeyStr
	// Generate key pair name
	if strings.EqualFold(*a.YmlConf.Template.SSHKeyPairName,"") {
		a.YmlConf.Template.SSHKeyPairName = utils.String(fmt.Sprintf("sshkey-%d",rand.Int()))
	}
	// Set key pair name to app config
	a.AppConf.KeyPairName = a.YmlConf.Template.SSHKeyPairName
	/**
	**
	**		MAIN PROCESS
	**
	 */
	switch os.Args[1] {
	case config.UpActionStr:
		if !isValidYmlConfig {
			a.AppConf.Init(config.UpActionStr, utils.String(""))
			if err := a.ImportKeyPair(); err != nil {
				log.Fatalln(err)
			}
			a.CreateInstances(a.AppConf.Nodes,utils.String("k8s"))
			// Set region to YAML
			a.YmlConf.Template.Region = a.AppConf.Region.Name
			a.OpenInstancePublicPorts()
			a.SetInstanceInfos()
			a.YmlConf.WriteConfig()
		}
	case config.DownActionStr:
		if isValidYmlConfig {
			a.AppConf.Init(config.DownActionStr, a.YmlConf.Template.Region)
			a.DeleteInstances()
			if err := a.RemoveKeyPair(); err != nil {
				log.Fatalln(err)
			}
			a.YmlConf.ConfigFile.RemoveFile()
		}

	default:
		fmt.Printf("expected %s or %s subcommands", config.UpActionStr, config.DownActionStr)
		os.Exit(1)
	}

}
