#!/bin/bash

plugin="docker"
vtep_ip="$VTEP_IP"
fwd_mode="bridge"
cstore="$CONTIV_ETCD"
cmode="bridge"
vlan_if="$VLAN_IF"
binpath="/contiv/bin"
loglocation="/var/contiv/log/netplugin.log"

while getopts ":p:v:i:f:c:b:l:" opt; do
    case $opt in
       b)
          binpath=$OPTARG
          ;;
       v)
          vtep_ip=$OPTARG
          ;;
       i)
          vlan_if=$OPTARG
          ;;
       f)
          fwd_mode=$OPTARG
          ;;
       c)
          cstore=$OPTARG
          ;; 
       p) 
          plugin=$OPTARG
          ;;
       l)
          loglocation=$OPTARG
          ;;
       :)
          echo "An argument required for $OPTARG was not passed"
          ;;
       ?)
          echo "Invalid option supplied"
          ;;
     esac
done


if [ "$cstore" != "" ]; then
   cstore_param="-cluster-store"
fi
if [ "$vtep_ip" != "" ]; then
   vtep_ip_param="-vtep-ip"
fi
if [ "$vlan_if" != "" ]; then
   vlan_if_param="-vlan-if"
fi


$binpath/netplugin $cstore_param $cstore $vtep_ip_param $vtep_ip $vlan_if_param $vlan_if -plugin-mode $plugin > $loglocation 2>&1
