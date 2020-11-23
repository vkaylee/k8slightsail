package nodeinfo

import "github.com/aws/aws-sdk-go/service/lightsail"

type Node struct {
	Name		*string
	Type		*string // master, worker, loadbalancing
	Bundle		*lightsail.Bundle
	Blueprint	*lightsail.Blueprint
}

const MasterNodeType = "master"
const WorkerNodeType = "worker"
const LoadbalancingType = "loadbalancing"