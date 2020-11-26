# k8slightsail
## Description
This tool is used for provisioning kubernetes (k8s) cluster base on lightsail service of AWS.

It's just using for lab k8s. Provision when you need, delete it when you dont need to save money.
- Number of master (depend on your choice)
- Number of worker (depend on your choice)
- Loadbalancing: 1

Some tools are preinstalled:
- Master: kubeadm, containerd, curl, kubectl, kubelet (systemd is preconfigured)
- Worker: kubeadm, containerd, curl, kubelet (systemd is preconfigured)
- Loadbalancing: curl, haproxy
## Tested

- Ubuntu 20.04 LTS

## Install golang compiler
https://golang.org/
## Install this app
`go install`
## Run
### Init cluster
`k8s_lightsail up`

config.yml file is going to be created automatically like
    
    region: ap-northeast-1
    nodes:
    - name: k8s-master-8674665223082153552
      type: master
      zone: ap-northeast-1a
      username: ubuntu
      public_ip: 18.183.224.164
      private_ip: 172.26.11.256
    - name: k8s-master-6129484611666145822
      type: master
      zone: ap-northeast-1c
      username: ubuntu
      public_ip: 175.41.232.63
      private_ip: 172.26.16.44
    - name: k8s-master-4037200794235010052
      type: master
      zone: ap-northeast-1d
      username: ubuntu
      public_ip: 13.230.39.207
      private_ip: 172.26.43.48
    - name: k8s-loadbalancing-3916589616287113938
      type: loadbalancing
      zone: ap-northeast-1a
      username: ubuntu
      public_ip: 13.115.118.175
      private_ip: 172.26.15.164
    - name: k8s-worker-6334824724549167321
      type: worker
      zone: ap-northeast-1c
      username: ubuntu
      public_ip: 54.249.173.204
      private_ip: 172.26.20.230
    - name: k8s-worker-605394647632969759
      type: worker
      zone: ap-northeast-1d
      username: ubuntu
      public_ip: 52.195.19.202
      private_ip: 172.26.33.174
    - name: k8s-worker-1443635317331776149
      type: worker
      zone: ap-northeast-1a
      username: ubuntu
      public_ip: 54.238.182.222
      private_ip: 172.26.7.185
    - name: k8s-worker-894385949183117217
      type: worker
      zone: ap-northeast-1c
      username: ubuntu
      public_ip: 18.183.47.62
      private_ip: 172.26.29.169
    - name: k8s-worker-2775422040480279450
      type: worker
      zone: ap-northeast-1d
      username: ubuntu
      public_ip: 3.112.110.17
      private_ip: 172.26.39.114
    ssh_key_pair_name: sshkey-5577006791947779411
    ssh_private_key: |
      -----BEGIN RSA PRIVATE KEY-----
      MIICXQIBAAKBgQCyJ5onfJVLZmRGq7Jdm9ujdEC97snUL12qA7JS6YBOPGELeroe
      UGaTEqj5Ib0rhR+J7mOIVNKHrS/gmkVMe4vCNiiCtU1zROVQ9TdF6U9CvY7PUQlD
      dow778mW6RSQ1MZkcXsTSbSGdIBhPDV0mz71dipxoLuTAz6ckGfHzgGapQIDAQAB
      AoGBAIVYSPS3NhOajwGqb7XK+6mrUO4Yte5giY3AeI/AgC2O2eA6uuYHrc71T44x
      Z6MUYBfgW5VmT7IHuec18Rqe+mpiANlEmfYrZigS+Xu4S7PMCJpTv09KuA/NuXVD
      XPp5S+oyuOyX/Yarus2H4r7lcOU9vLlAlmqYHluUCi5E6Wi9AkEA4IXwOWezGGFP
      mqauTZ1CFeCirby1S0qNUcFLHE5f4IV29lHy+NNDg2g+mCqItREiLK6UiaPuhQFT
      HHAq+hmg1wJBAMshg5OcVqBHQrwaJEXagsByaVd3RXm8V9WR6sb8N7IHmJffxLxw
      j/MgjUzJj9Jz7oavd7ZnkCBbnCi/0PqsZOMCQAnOD5WSL8IKzd0lFkuRaIdoDfKk
      YQ5urQk69brAuXMmoPFU1tWC9FnSvZkLknjFzMZCwX3ZSNtKGYUOaPIPGHUCQCid
      muF48Rk7JmzWDUqqVlqEheunPY0Jy8Y4VulSpRBD1I8JfxzupNnIOHiSFN/PrnHf
      w+AE9RyDNMYxFGgK8GECQQC0p1SPTLqkZaU4ZmM1o7BI/8YEMXPPI2NUK2w6hR4h
      s6YYyJvlc+tKsJX/KkGP6yO24kd/yk4mRMaA0X9NAGG+
      -----END RSA PRIVATE KEY-----
    ssh_public_key: |
      ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQCyJ5onfJVLZmRGq7Jdm9ujdEC97snUL12qA7JS6YBOPGELeroeUGaTEqj5Ib0rhR+J7mOIVNKHrS/gmkVMe4vCNiiCtU1zROVQ9TdF6U9CvY7PUQlDdow778mW6RSQ1MZkcXsTSbSGdIBhPDV0mz71dipxoLuTAz6ckGfHzgGapG==
    
### Next step
    - Config your loadbalancer with IP masters as layer 4
    - Init k8s cluster at master node with control lane IP = loadbalancer IP
### Delete cluster
`k8s_lightsail down`