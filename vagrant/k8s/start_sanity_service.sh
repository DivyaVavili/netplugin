#!/bin/bash

top_dir=$(git rev-parse --show-toplevel | sed 's|/[^/]*$||')
# run ansible
ansible-playbook -vvvv -i .contiv_k8s_inventory ./contrib/ansible/cluster.yml --tags "contiv_restart" -e "networking=contiv contiv_fabric_mode=default contiv_bin_path=$top_dir/netplugin/bin"
