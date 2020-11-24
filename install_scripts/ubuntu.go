package install_scripts

import "k8s_lightsail/utils"

type Ubuntu struct {
	masterScripts			[]*string
	workerScripts			[]*string
	loadbalancingScripts	[]*string
}
type scripts []*string
func (u *Ubuntu) Init() {
	a := scripts{}
	a = append(a,u.iptablesResetScript()...)
	a = append(a,u.basicAppScript()...)
	a = append(a,u.addDockerRepoScript()...)
	a = append(a,u.containerdScript()...)
	a = append(a,u.addKubeRepoScript()...)
	a = append(a,u.addK8sConfigScript()...)
	a = append(a,u.kubeadmKubeletScript()...)
	a = append(a,u.kubectlScript()...)
	u.masterScripts = a

	b := scripts{}
	b = append(b,u.iptablesResetScript()...)
	b = append(b,u.basicAppScript()...)
	b = append(b,u.addDockerRepoScript()...)
	b = append(b,u.containerdScript()...)
	b = append(b,u.addKubeRepoScript()...)
	b = append(b,u.addK8sConfigScript()...)
	b = append(b,u.kubeadmKubeletScript()...)
	u.workerScripts = b

	c := scripts{}
	c = append(c,u.iptablesResetScript()...)
	c = append(c,u.basicAppScript()...)
	c = append(c,u.haproxyScript()...)
	u.loadbalancingScripts = c
}
func (u *Ubuntu) GetMasterScripts() []*string {
	return u.masterScripts
}
func (u *Ubuntu) GetWorkerScripts() []*string {
	return u.workerScripts
}
func (u *Ubuntu) GetLoadbalancingScripts() []*string {
	return u.loadbalancingScripts
}
func (u Ubuntu) basicAppScript() []*string {
	return scripts{
		utils.String("sudo apt-get update"),
		// Install packages to allow apt to use a repository over HTTPS
		utils.String("sudo apt-get install -y apt-transport-https ca-certificates curl software-properties-common"),
	}
}
func (u Ubuntu) addDockerRepoScript() []*string {
	return scripts{
		// Add Docker's official GPG key
		utils.String("curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o docker.gpg"),
		utils.String("sudo apt-key --keyring /etc/apt/trusted.gpg.d/docker.gpg add docker.gpg"),
		// Add Docker apt repository.
		utils.String("sudo add-apt-repository \"deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable\""),
	}
}
func (u *Ubuntu) iptablesResetScript() []*string {
	return scripts{
		utils.String("sudo iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X"),
	}
}
func (u *Ubuntu) addK8sConfigScript() []*string {
	return scripts{
		// Setup required sysctl params, these persist across reboots.
		utils.String("echo 'net.bridge.bridge-nf-call-iptables  = 1' > /etc/sysctl.d/99-kubernetes-cri.conf"),
		utils.String("echo 'net.ipv4.ip_forward  = 1' >> /etc/sysctl.d/99-kubernetes-cri.conf"),
		utils.String("echo 'net.bridge.bridge-nf-call-ip6tables  = 1' >> /etc/sysctl.d/99-kubernetes-cri.conf"),
		utils.String("sudo modprobe br_netfilter"),
		// Letting iptables see bridged traffic
		utils.String("echo 'net.bridge.bridge-nf-call-ip6tables = 1' > /etc/sysctl.d/k8s.conf"),
		utils.String("echo 'net.bridge.bridge-nf-call-iptables = 1' >> /etc/sysctl.d/k8s.conf"),
		// Apply sysctl params without reboot
		utils.String("sudo sysctl --system"),
	}
}
func (u *Ubuntu) addKubeRepoScript() []*string {
	return scripts{
		utils.String("sudo apt-get update"),
		utils.String("curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg -o google-apt-key.gpg"),
		utils.String("sudo apt-key add google-apt-key.gpg"),
		utils.String("echo 'deb https://apt.kubernetes.io/ kubernetes-xenial main' > /etc/apt/sources.list.d/kubernetes.list"),
	}
}
func (u Ubuntu) kubeadmKubeletScript() []*string {
	return scripts{
		utils.String("sudo apt-get update"),
		utils.String("sudo apt-get install -y kubelet kubeadm"),
		utils.String("sudo apt-mark hold kubelet kubeadm"),
		utils.String("sudo systemctl daemon-reload"),
		utils.String("sudo systemctl restart kubelet"),
	}
}
func (u Ubuntu) kubectlScript() []*string {
	return scripts{
		utils.String("sudo apt-get update"),
		utils.String("sudo apt-get install -y kubectl"),
		utils.String("sudo apt-mark hold kubectl"),
	}
}
func (u Ubuntu) containerdScript() []*string {
	return scripts{
		// Install containerd
		utils.String("sudo apt-get update"),
		utils.String("sudo apt-get update && sudo apt-get install -y containerd.io"),
		utils.String("sudo mkdir -p /etc/containerd"),
		utils.String("echo 'overlay' > /etc/modules-load.d/containerd.conf"),
		utils.String("echo 'br_netfilter' >> /etc/modules-load.d/containerd.conf"),
		utils.String("sudo modprobe overlay"),
		utils.String("sudo modprobe br_netfilter"),
		// Restart containerd
		utils.String("sudo systemctl restart containerd"),
	}
}

func (u Ubuntu) haproxyScript() []*string {
	return scripts{
		utils.String("sudo apt-get update"),
		utils.String("sudo apt-get install -y haproxy"),
	}
}