package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/jedib0t/go-pretty/table"
	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
	"k8s_lightsail/app"
	"k8s_lightsail/config"
	"k8s_lightsail/utils"
	"k8s_lightsail/ymlconfig"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
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
	case config.LsActionStr:
		if isValidYmlConfig {
			a.AppConf.Init(config.LsActionStr, a.YmlConf.Template.Region)
			// Show table
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"#", "Node name", "Type", "Username", "Public IP", "Private IP", "Zone"})
			for index, node := range a.YmlConf.Template.Nodes {
				t.AppendRows([]table.Row{
					{index+1, *node.Name, *node.Type, *node.Username, *node.PublicIp, *node.PrivateIp, *node.Zone},
				})
			}
			t.Render()
		}
	case config.SSHActionStr:
		if isValidYmlConfig {
			num := os.Args[2]
			i, err := strconv.Atoi(num)
			if err == nil {
				nodes := a.YmlConf.Template.Nodes
				if i > 0 && i <= len(nodes) {
					node := nodes[i-1]
					sshKeyFile := utils.CreateTempFile(a.YmlConf.Template.SSHPrivateKeyStr, a.YmlConf.Template.SSHKeyPairName)
					defer os.Remove(sshKeyFile.Name())
					defer sshKeyFile.Close()
					auth, err := goph.Key(sshKeyFile.Name(), "")
					if err != nil {
						log.Println(err)
					}
					client, err := goph.NewConn(&goph.Config{
						User:     *node.Username,
						Addr:     *node.PublicIp,
						Port:     22,
						Auth:     auth,
						Callback: VerifyHost,
					})
					if err != nil {
						log.Println(err)
					}
					// Close client net connection
					defer client.Close()
					playWithSSH(client,fmt.Sprintf("%s(%s): ", *node.Type, *node.PublicIp))
				}
			}

		}

	default:
		fmt.Printf("expected %s or %s subcommands", config.UpActionStr, config.DownActionStr)
		os.Exit(1)
	}

}

func VerifyHost(host string, remote net.Addr, key ssh.PublicKey) error {

	//
	// If you want to connect to new hosts.
	// here your should check new connections public keys
	// if the key not trusted you shuld return an error
	//

	// hostFound: is host in known hosts file.
	// err: error if key not in known hosts file OR host in known hosts file but key changed!
	hostFound, err := goph.CheckKnownHost(host, remote, key, "")

	// Host in known hosts but key mismatch!
	// Maybe because of MAN IN THE MIDDLE ATTACK!
	if hostFound && err != nil {

		return err
	}

	// handshake because public key already exists.
	if hostFound && err == nil {

		return nil
	}

	// Ask user to check if he trust the host public key.
	if askIsHostTrusted(host, key) == false {

		// Make sure to return error on non trusted keys.
		return errors.New("you typed no, aborted")
	}

	// Add the new host to known hosts file.
	return goph.AddKnownHost(host, remote, key, "")
}
func askIsHostTrusted(host string, key ssh.PublicKey) bool {

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Unknown Host: %s \nFingerprint: %s \n", host, ssh.FingerprintSHA256(key))
	fmt.Print("Would you like to add it? type yes or no: ")

	a, err := reader.ReadString('\n')

	if err != nil {
		log.Fatal(err)
	}

	return strings.ToLower(strings.TrimSpace(a)) == "yes"
}

func playWithSSH(client *goph.Client, machineName string) {
	fmt.Printf("Connected to %s\n", client.Config.Addr)
	fmt.Println("Type your shell command and enter.")
	fmt.Println("To download file from remote type: download remote/path local/path")
	fmt.Println("To upload file to remote type: upload local/path remote/path")
	fmt.Println("To exit press CTRL + C")

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print(machineName)

	var (
		out   []byte
		err   error
		cmd   string
		parts []string
	)

	for scanner.Scan() {

		err = nil
		cmd = scanner.Text()
		parts = strings.Split(cmd, " ")

		if len(parts) < 1 {
			continue
		}

		switch parts[0] {
		case "download":
			if len(parts) != 3 {
				fmt.Println("please type valid download command!")
			}

			err = client.Download(parts[1], parts[2])

			if err != nil {
				log.Println(err)
			}

		case "upload":

			if len(parts) != 3 {
				fmt.Println("please type valid upload command!")
			}

			err = client.Upload(parts[1], parts[2])
			if err != nil {
				log.Println(err)
			}

		default:
			command, err := client.Command(parts[0], parts[1:]...)
			if err != nil {
				log.Println(err)
			}
			out, err = command.CombinedOutput()
			if err != nil {
				log.Println(err)
			}
			fmt.Println(string(out))
		}
		fmt.Print(machineName)
	}
}
