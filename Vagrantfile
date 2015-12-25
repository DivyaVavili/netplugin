# -*- mode: ruby -*-
# vi: set ft=ruby :

require 'fileutils'

# netplugin_synced_gopath="/opt/golang"
gopath_folder="/opt/gopath"
FileUtils.cp "/etc/resolv.conf", Dir.pwd
proxy_env = { }
no_proxy = "192.168.2.10,192.168.2.11,127.0.0.1,localhost,netmaster"

%w[HTTP_PROXY HTTPS_PROXY http_proxy https_proxy].each do |name|
  if ENV[name]
    proxy_env[name] = ENV[name]
  end
end
proxy_env["no_proxy"] = no_proxy

ansible_groups = { }
ansible_playbook = "./vendor/ansible/site.yml"
ansible_extra_vars = {
    "env" => proxy_env
}

provision_common = <<SCRIPT
## setup the environment file. Export the env-vars passed as args to 'vagrant up'
echo Args passed: [[ $@ ]]

echo -n "$1" > /etc/hostname
hostname -F /etc/hostname

/sbin/ip addr add "$3/24" dev eth1
/sbin/ip link set eth1 up
/sbin/ip link set eth2 up

echo 'export GOPATH=#{gopath_folder}' > /etc/profile.d/envvar.sh
echo 'export GOBIN=$GOPATH/bin' >> /etc/profile.d/envvar.sh
echo 'export GOSRC=$GOPATH/src' >> /etc/profile.d/envvar.sh
echo 'export PATH=$PATH:/usr/local/go/bin:$GOBIN' >> /etc/profile.d/envvar.sh
echo "export http_proxy='$4'" >> /etc/profile.d/envvar.sh
echo "export https_proxy='$5'" >> /etc/profile.d/envvar.sh
echo "export no_proxy=#{no_proxy}" >> /etc/profile.d/envvar.sh

if [ $# -gt 5 ]; then
    shift; shift; shift; shift; shift
    echo "export $@" >> /etc/profile.d/envvar.sh
fi

source /etc/profile.d/envvar.sh

mv /etc/resolv.conf /etc/resolv.conf.bak
cp #{gopath_folder}/src/github.com/contiv/netplugin/resolv.conf /etc/resolv.conf

docker load --input #{gopath_folder}/src/github.com/contiv/netplugin/scripts/dnscontainer.tar
SCRIPT

VAGRANTFILE_API_VERSION = "2"
Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
    if ENV['CONTIV_NODE_OS'] && ENV['CONTIV_NODE_OS'] == "centos" then
        config.vm.box = "contiv/centos71-netplugin"
        config.vm.box_version = "0.3.1"
    else
        config.vm.box = "contiv/ubuntu1504-netplugin"
        config.vm.box_version = "0.3.1"
    end
    config.ssh.insert_key = false
    num_nodes = 2
    if ENV['CONTIV_NODES'] && ENV['CONTIV_NODES'] != "" then
        num_nodes = ENV['CONTIV_NODES'].to_i
    end
    base_ip = "192.168.2."
    node_ips = num_nodes.times.collect { |n| base_ip + "#{n+10}" }
    node_names = num_nodes.times.collect { |n| "netplugin-node#{n+1}" }
    node_peers = []

    num_nodes.times do |n|
        node_name = node_names[n]
        node_addr = node_ips[n]
        node_peers += ["#{node_name}=http://#{node_addr}:2380,#{node_name}=http://#{node_addr}:7001"]
        consul_join_flag = if n > 0 then "-join #{node_ips[0]}" else "" end
        consul_bootstrap_flag = "-bootstrap-expect=3"
        if num_nodes < 3 then
            if n == 0 then
                consul_bootstrap_flag = "-bootstrap"
            else
                consul_bootstrap_flag = ""
            end
        end
        config.vm.define node_name do |node|
            # node.vm.hostname = node_name
            # create an interface for etcd cluster
            node.vm.network :private_network, ip: node_addr, virtualbox__intnet: "true", auto_config: false
            # create an interface for bridged ne2twork
            node.vm.network :private_network, ip: "0.0.0.0", virtualbox__intnet: "true", auto_config: false
            node.vm.provider "virtualbox" do |v|
                # make all nics 'virtio' to take benefit of builtin vlan tag
                # support, which otherwise needs to be enabled in Intel drivers,
                # which are used by default by virtualbox
                v.customize ['modifyvm', :id, '--nictype1', 'virtio']
                v.customize ['modifyvm', :id, '--nictype2', 'virtio']
                v.customize ['modifyvm', :id, '--nictype3', 'virtio']
                v.customize ['modifyvm', :id, '--nicpromisc2', 'allow-all']
                v.customize ['modifyvm', :id, '--nicpromisc3', 'allow-all']
            end

            # mount the host directories
            node.vm.synced_folder "bin", File.join(gopath_folder, "bin")
            if ENV["GOPATH"] && ENV['GOPATH'] != ""
              node.vm.synced_folder "../../../", File.join(gopath_folder, "src"), rsync: true
            else
              node.vm.synced_folder ".", File.join(gopath_folder, "src/github.com/contiv/netplugin"), rsync: true
            end

            node.vm.provision "shell" do |s|
                s.inline = "echo '#{node_ips[0]} netmaster' >> /etc/hosts; echo '#{node_addr} #{node_name}' >> /etc/hosts"
            end
            node.vm.provision "shell" do |s|
                s.inline = provision_common
                s.args = [node_name, ENV["CONTIV_NODE_OS"] || "", node_addr, ENV["http_proxy"] || "", ENV["https_proxy"] || "", *ENV['CONTIV_ENV']]
            end
            # start netmaster on the first vm
            if ansible_groups["netplugin-node"] == nil then
                ansible_groups["netplugin-node"] = [ ]
            end
            ansible_groups["netplugin-node"] << node_name
            if n == 0 then
                # forward netmaster port
                node.vm.network "forwarded_port", guest: 9999, host: 9999
            end
            # Run the provisioner after all machines are up
            if n == (num_nodes - 1) then
                node.vm.provision 'ansible' do |ansible|
                    ansible.groups = ansible_groups
                    ansible.playbook = ansible_playbook
                    ansible.extra_vars = {
                        env: proxy_env,
                        monitor_interface: 'eth1',
                        etcd_peers_group: 'netplugin-node',
                        service_vip: node_ips[0],
                        deploy_released_version: false
                    }
                    ansible.limit = 'all'
                end
            end
provision_node = <<SCRIPT
## start consul
(nohup consul agent -server #{consul_join_flag} #{consul_bootstrap_flag} \
 -bind=#{node_addr} -data-dir /opt/consul 0<&- &>/tmp/consul.log &) || exit 1

SCRIPT
            node.vm.provision "shell", run: "always" do |s|
                s.inline = provision_node
            end
        end
    end
end
