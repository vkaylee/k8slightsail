package config

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/manifoldco/promptui"
	"k8s_lightsail/nodeinfo"
	"k8s_lightsail/utils"
	"strconv"
	"strings"
)

const UpActionStr = "up"
const DownActionStr = "down"
const LsActionStr 	= "ls"
const SSHActionStr	= "ssh"

type Config struct {
	Region			*lightsail.Region
	Lightsail		*lightsail.Lightsail
	Nodes			[]*nodeinfo.Node
	KeyPairName		*string
	PublicKeyStr	*string
	PrivateKeyStr	*string
}

type OptionChoiceStruct struct {
	Col1		string
	Col2		string
	Description string
}

func (c *Config) Init(action string, clientStr *string) {
	if "" == *clientStr {
		c.setClient("ap-northeast-1")
	} else {
		c.setClient(*clientStr)
	}
	if action == UpActionStr {
		if c.Region == nil {
			c.getRegionsAndZones()
		}
		c.inputNodeSpecs()
	}
}
func (c *Config) inputNodeSpecs() {
	numberOfMaster := c.inputNumber("Please input number of master node")
	numberOfWorker := c.inputNumber("Please input number of worker node")
	isTheSameSpec := c.yesNoChoose("Master, worker and loadbalancing nodes are the same spec?")
	if isTheSameSpec {
		fmt.Println("Please choose your node spec:")
		masterNode := c.inputNodeSpec(nil, utils.String(nodeinfo.MasterNodeType))
		loadbalancingNode := *masterNode
		loadbalancingNode.Type = utils.String(nodeinfo.LoadbalancingType)
		workerNode := *masterNode
		workerNode.Type = utils.String(nodeinfo.WorkerNodeType)
		c.Nodes = c.nodeAppend(c.Nodes, masterNode, numberOfMaster)
		c.Nodes = c.nodeAppend(c.Nodes, &loadbalancingNode, 1)
		c.Nodes = c.nodeAppend(c.Nodes, &workerNode, numberOfWorker)
	} else {
		fmt.Println("Please choose your master node spec:")
		c.Nodes = c.nodeAppend(c.Nodes, c.inputNodeSpec(nil, utils.String(nodeinfo.MasterNodeType)), numberOfMaster)
		fmt.Println("Please choose your loadbalancing node spec:")
		c.Nodes = c.nodeAppend(c.Nodes, c.inputNodeSpec(nil, utils.String(nodeinfo.LoadbalancingType)), 1)
		fmt.Println("Please choose your worker node spec:")
		c.Nodes = c.nodeAppend(c.Nodes, c.inputNodeSpec(nil, utils.String(nodeinfo.WorkerNodeType)), numberOfWorker)
	}
}
func (c *Config) nodeAppend(nodes []*nodeinfo.Node, node *nodeinfo.Node, number uint64) []*nodeinfo.Node {
	var i uint64 = 0
	for i = 0; i < number; i++ {
		nodes = append(nodes,node)
	}
	return nodes
}
func (c *Config) inputNumber(msg string) uint64 {
	validate := func(input string) error {
		_, err := strconv.ParseUint(input,10,64)
		if err != nil {
			return errors.New("invalid number")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    msg,
		Validate: validate,
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return 0
	}
	number, _ := strconv.ParseUint(result,10,64)
	return number
}
func (c *Config) inputNodeSpec(nodeName *string, nodeType *string) *nodeinfo.Node {
	return &nodeinfo.Node{
		Name:      	nodeName,
		Bundle:    	c.getBundles(),
		Blueprint: 	c.getBlueprints(),
		Type: 		nodeType,
	}
}

func (c *Config) yesNoChoose(msg string) bool {
	const yesItem = "yes"
	prompt := promptui.Select{
		Label: msg,
		Items: []string{yesItem,"no"},
	}
	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return false
	}
	return strings.EqualFold(result,yesItem)
}
func (c *Config) getRegionsAndZones() {
	outputPtrRegions, err := c.Lightsail.GetRegionsWithContext(
		aws.BackgroundContext(),
		&lightsail.GetRegionsInput{
			IncludeAvailabilityZones:                   aws.Bool(true),
			IncludeRelationalDatabaseAvailabilityZones: aws.Bool(false),
		},
	)

	if err != nil {
		panic(err)
	}

	options := make([]OptionChoiceStruct,0)

	for _, outputPtrRegion := range outputPtrRegions.Regions {
		options = append(options, OptionChoiceStruct{
			Col1: *outputPtrRegion.DisplayName,
			Col2: *outputPtrRegion.Name,
			Description: *outputPtrRegion.Description,
		})
	}
	choiceIndex, err := c.optionsChoice(options)
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}
	c.Region = outputPtrRegions.Regions[choiceIndex]
	c.setClientAgain(choiceIndex)
}
func (c *Config) setClientAgain(lastIndex int)  {
	c.setClient(*c.Region.Name)
	outputPtrRegions, err := c.Lightsail.GetRegionsWithContext(
		aws.BackgroundContext(),
		&lightsail.GetRegionsInput{
			IncludeAvailabilityZones:                   aws.Bool(true),
			IncludeRelationalDatabaseAvailabilityZones: aws.Bool(false),
		},
	)
	if err != nil {
		panic(err)
	}
	c.Region = outputPtrRegions.Regions[lastIndex]
}
func (c *Config) getBlueprints() *lightsail.Blueprint {
	outputPtrBlueprints, err := c.Lightsail.GetBlueprintsWithContext(
		aws.BackgroundContext(),
		&lightsail.GetBlueprintsInput{
			IncludeInactive: aws.Bool(false),
			PageToken:       nil,
		},
	)

	if err != nil {
		panic(err)
	}

	var blueprints []*lightsail.Blueprint
	options := make([]OptionChoiceStruct,0)

	for _, v := range outputPtrBlueprints.Blueprints {
		if "os" == *v.Type && "LINUX_UNIX" == *v.Platform {
			options = append(options, OptionChoiceStruct{
				Col1:        *v.Name,
				Col2:     *v.Version,
				Description: fmt.Sprintf("Platform: %s, Type: %s",*v.Platform,*v.Type),
			})
			blueprints = append(blueprints,v)
		}
	}

	choiceIndex, err := c.optionsChoice(options)
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return nil
	}
	if blueprints != nil {
		return blueprints[choiceIndex]
	}
	return nil
}
func (c *Config) getBundles() *lightsail.Bundle {
	outputPtrBundles, err := c.Lightsail.GetBundlesWithContext(
		aws.BackgroundContext(),
		&lightsail.GetBundlesInput{
			IncludeInactive: aws.Bool(false),
			PageToken:       nil,
		},
	)

	if err != nil {
		panic(err)
	}

	options := make([]OptionChoiceStruct,0)

	for _, v1 := range outputPtrBundles.Bundles {
		supportedPlatforms := v1.SupportedPlatforms
		isOk := false
		for _, v2 := range supportedPlatforms {
			if "LINUX_UNIX" == *v2 {
				isOk = true
				break
			}
		}
		if isOk {
			options = append(options, OptionChoiceStruct{
				Col1: *v1.Name,
				Col2: *v1.BundleId,
				Description: fmt.Sprintf(
					"CPU: %d, Ram: %fG, Disk %dG, Price: %f$",
					*v1.CpuCount,
					*v1.RamSizeInGb,
					*v1.DiskSizeInGb,
					*v1.Price,
				),
			})
		}
	}

	choiceIndex, err := c.optionsChoice(options)
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return nil
	}
	return outputPtrBundles.Bundles[choiceIndex]
}

func (c *Config) setClient(regionStr string)  {
	// All clients require a Session. The Session provides the client with
	// shared configuration such as region, endpoint, and credentials. A
	// Session should be shared where possible to take advantage of
	// configuration and credential caching. See the session package for
	// more information.
	sess := session.Must(session.NewSession())
	c.Lightsail = lightsail.New(sess, aws.NewConfig().WithRegion(regionStr))

}

func (c *Config) optionsChoice(choices []OptionChoiceStruct) (int, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "✅  {{ .Col1 | cyan }} ({{ .Col2 | red }})",
		Inactive: "   {{ .Col1 | cyan }} ({{ .Col2 | red }})",
		Selected: "✅  {{ .Col1 | cyan }} ({{ .Col2 | red }})",
		Details: `{{ .Description }}`,
	}

	searcher := func(input string, index int) bool {
		choice := choices[index]
		name := strings.Replace(strings.ToLower(fmt.Sprintf(
			"%s (%s)",
			choice.Col1,
			choice.Col2,
		),
		), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "Select:",
		Items:     choices,
		Templates: templates,
		Size:      4,
		Searcher:  searcher,
	}

	i, _, err := prompt.Run()

	return i, err
}