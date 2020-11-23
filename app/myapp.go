package app

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"k8s_lightsail/config"
	"k8s_lightsail/nodeinfo"
	"k8s_lightsail/utils"
	"k8s_lightsail/ymlconfig"
	"log"
	"math/rand"
	"strings"
	"time"
)

type App struct {
	AppConf	*config.Config
	YmlConf	*ymlconfig.Config
	Ctx		context.Context
}

func (a *App) OpenInstancePublicPorts()  {
	for _, node := range a.YmlConf.Template.Nodes {
		a.waitForLightsailInstance(node.Name)
		log.Printf("Setting ports for %s", *node.Name)
		switch *node.Type {
		case nodeinfo.MasterNodeType:
			a.openMasterInstancePublicPorts(node.Name)
		case nodeinfo.WorkerNodeType:
			a.openWorkerInstancePublicPorts(node.Name)
		case nodeinfo.LoadbalancingType:
			a.openLoadbalancingInstancePublicPorts(node.Name)
		default:
			break
		}
	}
}
func (a *App) openMasterInstancePublicPorts(instanceName *string)  {
	ctx, cancelFn := context.WithTimeout(a.Ctx, time.Minute * 5)
	// Ensure the context is canceled to prevent leaking.
	// See context package for more information, https://golang.org/pkg/context/
	if cancelFn != nil {
		defer cancelFn()
	}
	_, _ = a.AppConf.Lightsail.PutInstancePublicPortsWithContext(ctx, &lightsail.PutInstancePublicPortsInput{
		InstanceName: instanceName,
		PortInfos:    []*lightsail.PortInfo{
			&lightsail.PortInfo{
				FromPort: utils.Init64(22),
				Protocol: utils.String("tcp"),
				ToPort:   utils.Init64(22),
			},
			&lightsail.PortInfo{
				FromPort: utils.Init64(80),
				Protocol: utils.String("tcp"),
				ToPort:   utils.Init64(80),
			},
			&lightsail.PortInfo{
				FromPort: utils.Init64(443),
				Protocol: utils.String("tcp"),
				ToPort:   utils.Init64(443),
			},
			&lightsail.PortInfo{
				FromPort: utils.Init64(2379),
				Protocol: utils.String("tcp"),
				ToPort:   utils.Init64(2380),
			},
			&lightsail.PortInfo{
				FromPort: utils.Init64(6443),
				Protocol: utils.String("tcp"),
				ToPort:   utils.Init64(6443),
			},
			&lightsail.PortInfo{
				FromPort: utils.Init64(10250),
				Protocol: utils.String("tcp"),
				ToPort:   utils.Init64(10252),
			},
		},
	})
}
func (a *App) openWorkerInstancePublicPorts(instanceName *string)  {
	ctx, cancelFn := context.WithTimeout(a.Ctx, time.Minute * 5)
	// Ensure the context is canceled to prevent leaking.
	// See context package for more information, https://golang.org/pkg/context/
	if cancelFn != nil {
		defer cancelFn()
	}
	_, _ = a.AppConf.Lightsail.PutInstancePublicPortsWithContext(ctx, &lightsail.PutInstancePublicPortsInput{
		InstanceName: instanceName,
		PortInfos: []*lightsail.PortInfo{
			&lightsail.PortInfo{
				FromPort: utils.Init64(22),
				Protocol: utils.String("tcp"),
				ToPort:   utils.Init64(22),
			},
			&lightsail.PortInfo{
				FromPort: utils.Init64(30000),
				Protocol: utils.String("tcp"),
				ToPort:   utils.Init64(32767),
			},
			&lightsail.PortInfo{
				FromPort: utils.Init64(10250),
				Protocol: utils.String("tcp"),
				ToPort:   utils.Init64(10250),
			},
		},
	})
}
func (a *App) openLoadbalancingInstancePublicPorts(instanceName *string)  {
	ctx, cancelFn := context.WithTimeout(a.Ctx, time.Minute * 5)
	// Ensure the context is canceled to prevent leaking.
	// See context package for more information, https://golang.org/pkg/context/
	if cancelFn != nil {
		defer cancelFn()
	}
	_, _ = a.AppConf.Lightsail.PutInstancePublicPortsWithContext(ctx, &lightsail.PutInstancePublicPortsInput{
		InstanceName: instanceName,
		PortInfos: []*lightsail.PortInfo{
			&lightsail.PortInfo{
				FromPort: utils.Init64(22),
				Protocol: utils.String("tcp"),
				ToPort:   utils.Init64(22),
			},
			&lightsail.PortInfo{
				FromPort: utils.Init64(6443),
				Protocol: utils.String("tcp"),
				ToPort:   utils.Init64(6443),
			},
		},
	})
}
func (a *App) CreateInstances(nodes []*nodeinfo.Node, prefixName *string) {
	ctx, cancelFn := context.WithTimeout(a.Ctx, time.Minute * 5)
	// Ensure the context is canceled to prevent leaking.
	// See context package for more information, https://golang.org/pkg/context/
	if cancelFn != nil {
		defer cancelFn()
	}
	strZones := a.getStrZones()
	if len(strZones) > 0 {
		userData := a.loadUserData()

		for index, node := range nodes {
			uData := userData
			if strings.EqualFold(nodeinfo.LoadbalancingType, *node.Type) {
				uData = utils.String("sudo apt-get install haproxy -y")
			}
			result, err := a.AppConf.Lightsail.CreateInstancesWithContext(
				ctx,
				&lightsail.CreateInstancesInput{
					AvailabilityZone: 	utils.String(strZones[index%len(strZones)]),
					BlueprintId:      	node.Blueprint.BlueprintId,
					BundleId:         	node.Bundle.BundleId,
					KeyPairName:		a.AppConf.KeyPairName,
					UserData:			uData,
					InstanceNames: []*string{
						utils.String(fmt.Sprintf("%s-%s-%d",*prefixName,*node.Type,rand.Int())),
					},
				},
			)

			if err != nil {
				panic(err)
			}
			log.Printf("Created: %s\n",*result.Operations[0].ResourceName)

			// Add node template after creating successfully
			a.YmlConf.Template.Nodes = append(a.YmlConf.Template.Nodes, &ymlconfig.Node{
				Name: result.Operations[0].ResourceName,
				Type: node.Type,
			})
		}
	}
}

func (a *App) getStrZones() []string {
	strZones := make([]string, 0)
	for _, v := range a.AppConf.Region.AvailabilityZones {
		if "available" == *v.State {
			strZones = append(strZones, *v.ZoneName)
		}
	}
	return strZones
}
func (a *App) DeleteInstances()  {
	ctx, cancelFn := context.WithTimeout(a.Ctx, time.Minute * 5)
	// Ensure the context is canceled to prevent leaking.
	// See context package for more information, https://golang.org/pkg/context/
	if cancelFn != nil {
		defer cancelFn()
	}
	for _, node := range a.YmlConf.Template.Nodes {
		result, err := a.AppConf.Lightsail.DeleteInstanceWithContext(
			ctx,
			&lightsail.DeleteInstanceInput{
				ForceDeleteAddOns: utils.Bool(true),
				InstanceName: node.Name,
			},
		)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				log.Printf("%s", awsErr.Message())
			}
		} else {
			log.Printf("Deleted: %s\n", *result.Operations[0].ResourceName)
		}
	}
}
func (a *App) loadUserData() *string {
	userDataFile := utils.InitFileWithPath(utils.String("userdata.sh"))
	err := userDataFile.Loadfile()
	if err != nil {
		return utils.String("")
	}
	return userDataFile.GetContents()
}
func (a *App) getInstanceState(instanceName *string) (*lightsail.GetInstanceStateOutput, error) {
	return a.AppConf.Lightsail.GetInstanceState(&lightsail.GetInstanceStateInput{
		InstanceName: instanceName,
	})
}
func (a *App) checkLightsailInstanceIsRunning(instanceName *string) bool {
	// Call AWS SDK
	result, err := a.getInstanceState(instanceName)
	if err != nil {
		log.Println(err)
		return false
	}
	// the instance is running if the state code == 16
	if *result.State.Code == 16 {
		return true
	}
	return false
}
func (a *App) waitForLightsailInstance(instanceName *string) {
	for i:= 0; i < 60; i++ {
		if a.checkLightsailInstanceIsRunning(instanceName) {
			break
		}
		time.Sleep(time.Second * 3)
	}
}

func (a *App) SetInstanceInfos()  {
	ctx, cancelFn := context.WithTimeout(a.Ctx, time.Minute * 5)
	// Ensure the context is canceled to prevent leaking.
	// See context package for more information, https://golang.org/pkg/context/
	if cancelFn != nil {
		defer cancelFn()
	}
	for _, node := range a.YmlConf.Template.Nodes {
		result, err := a.AppConf.Lightsail.GetInstanceWithContext(ctx, &lightsail.GetInstanceInput{InstanceName: node.Name})
		if err != nil {
			log.Println(err.(awserr.Error).Message())
		}
		a.waitForLightsailInstance(node.Name)
		if result != nil {
			node.PublicIp = result.Instance.PublicIpAddress
			node.PrivateIp = result.Instance.PrivateIpAddress
			node.Zone = result.Instance.Location.AvailabilityZone
			node.Username = result.Instance.Username
		}
	}
}
func (a *App) RemoveKeyPair() error {
	log.Println("Removing the lightsail key pair...")
	if _, err := a.AppConf.Lightsail.DeleteKeyPair(&lightsail.DeleteKeyPairInput{
		KeyPairName: a.AppConf.KeyPairName,
	}); err != nil {
		log.Println("We got an error when deleting the lightsail key pair.")
		return err
	}
	log.Println("The lightsail key pair has been removed.")
	return nil
}
func (a *App) getKeyPairInfo() (*lightsail.GetKeyPairOutput, error) {
	return a.AppConf.Lightsail.GetKeyPair(&lightsail.GetKeyPairInput{
		KeyPairName: a.AppConf.KeyPairName,
	})
}
func (a *App) ImportKeyPair() error {
	// Check KeyPairName in lightsail
	if _, err := a.getKeyPairInfo(); err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() != lightsail.ErrCodeNotFoundException {
				// Remove lightsail keypair
				if err := a.RemoveKeyPair(); err != nil {
					return err
				}
			}
		}
	}

	if _, err := a.AppConf.Lightsail.ImportKeyPair(&lightsail.ImportKeyPairInput{
		KeyPairName:     a.AppConf.KeyPairName,
		PublicKeyBase64: a.AppConf.PublicKeyStr,
	}); err != nil {
		return err
	}
	return nil
}