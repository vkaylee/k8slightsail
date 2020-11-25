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
	a = append(a,u.containerdConfigScript()...)
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
	b = append(b,u.containerdConfigScript()...)
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
func (u Ubuntu) containerdConfigScript() []*string {
	return scripts{
		utils.String("cat <<EOF | tee /etc/containerd/config.toml\nversion = 2\nroot = \"/var/lib/containerd\"\nstate = \"/run/containerd\"\nplugin_dir = \"\"\ndisabled_plugins = []\nrequired_plugins = []\noom_score = 0\n\n[grpc]\n  address = \"/run/containerd/containerd.sock\"\n  tcp_address = \"\"\n  tcp_tls_cert = \"\"\n  tcp_tls_key = \"\"\n  uid = 0\n  gid = 0\n  max_recv_message_size = 16777216\n  max_send_message_size = 16777216\n\n[ttrpc]\n  address = \"\"\n  uid = 0\n  gid = 0\n\n[debug]\n  address = \"\"\n  uid = 0\n  gid = 0\n  level = \"\"\n\n[metrics]\n  address = \"\"\n  grpc_histogram = false\n\n[cgroup]\n  path = \"\"\n\n[timeouts]\n  \"io.containerd.timeout.shim.cleanup\" = \"5s\"\n  \"io.containerd.timeout.shim.load\" = \"5s\"\n  \"io.containerd.timeout.shim.shutdown\" = \"3s\"\n  \"io.containerd.timeout.task.state\" = \"2s\"\n\n[plugins]\n  [plugins.\"io.containerd.gc.v1.scheduler\"]\n\tpause_threshold = 0.02\n\tdeletion_threshold = 0\n\tmutation_threshold = 100\n\tschedule_delay = \"0s\"\n\tstartup_delay = \"100ms\"\n  [plugins.\"io.containerd.grpc.v1.cri\"]\n\tdisable_tcp_service = true\n\tstream_server_address = \"127.0.0.1\"\n\tstream_server_port = \"0\"\n\tstream_idle_timeout = \"4h0m0s\"\n\tenable_selinux = false\n\tsandbox_image = \"k8s.gcr.io/pause:3.1\"\n\tstats_collect_period = 10\n\tsystemd_cgroup = false\n\tenable_tls_streaming = false\n\tmax_container_log_line_size = 16384\n\tdisable_cgroup = false\n\tdisable_apparmor = false\n\trestrict_oom_score_adj = false\n\tmax_concurrent_downloads = 3\n\tdisable_proc_mount = false\n\t[plugins.\"io.containerd.grpc.v1.cri\".containerd]\n\t  snapshotter = \"overlayfs\"\n\t  default_runtime_name = \"runc\"\n\t  no_pivot = false\n\t  [plugins.\"io.containerd.grpc.v1.cri\".containerd.default_runtime]\n\t\truntime_type = \"\"\n\t\truntime_engine = \"\"\n\t\truntime_root = \"\"\n\t\tprivileged_without_host_devices = false\n\t  [plugins.\"io.containerd.grpc.v1.cri\".containerd.untrusted_workload_runtime]\n\t\truntime_type = \"\"\n\t\truntime_engine = \"\"\n\t\truntime_root = \"\"\n\t\tprivileged_without_host_devices = false\n\t  [plugins.\"io.containerd.grpc.v1.cri\".containerd.runtimes]\n\t\t[plugins.\"io.containerd.grpc.v1.cri\".containerd.runtimes.runc]\n\t\t  runtime_type = \"io.containerd.runc.v1\"\n\t\t  runtime_engine = \"\"\n\t\t  runtime_root = \"\"\n\t\t  privileged_without_host_devices = false\n\t\t[plugins.\"io.containerd.grpc.v1.cri\".containerd.runtimes.runc.options]\n\t\t  SystemdCgroup = true\n\t[plugins.\"io.containerd.grpc.v1.cri\".cni]\n\t  bin_dir = \"/opt/cni/bin\"\n\t  conf_dir = \"/etc/cni/net.d\"\n\t  max_conf_num = 1\n\t  conf_template = \"\"\n\t[plugins.\"io.containerd.grpc.v1.cri\".registry]\n\t  [plugins.\"io.containerd.grpc.v1.cri\".registry.mirrors]\n\t\t[plugins.\"io.containerd.grpc.v1.cri\".registry.mirrors.\"docker.io\"]\n\t\t  endpoint = [\"https://registry-1.docker.io\"]\n\t[plugins.\"io.containerd.grpc.v1.cri\".x509_key_pair_streaming]\n\t  tls_cert_file = \"\"\n\t  tls_key_file = \"\"\n  [plugins.\"io.containerd.internal.v1.opt\"]\n\tpath = \"/opt/containerd\"\n  [plugins.\"io.containerd.internal.v1.restart\"]\n\tinterval = \"10s\"\n  [plugins.\"io.containerd.metadata.v1.bolt\"]\n\tcontent_sharing_policy = \"shared\"\n  [plugins.\"io.containerd.monitor.v1.cgroups\"]\n\tno_prometheus = false\n  [plugins.\"io.containerd.runtime.v1.linux\"]\n\tshim = \"containerd-shim\"\n\truntime = \"runc\"\n\truntime_root = \"\"\n\tno_shim = false\n\tshim_debug = false\n  [plugins.\"io.containerd.runtime.v2.task\"]\n\tplatforms = [\"linux/amd64\"]\n  [plugins.\"io.containerd.service.v1.diff-service\"]\n\tdefault = [\"walking\"]\n  [plugins.\"io.containerd.snapshotter.v1.devmapper\"]\n\troot_path = \"\"\n\tpool_name = \"\"\n\tbase_image_size = \"\"\nEOF"),
		// Restart containerd
		utils.String("systemctl restart containerd"),
	}
}