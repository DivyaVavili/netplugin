#!/bin/bash

debug=false
dns_enable=false
listen_url="$LISTEN_URL"
plugin="docker"
cstore="$CONTIV_ETCD"
binpath="/contiv/bin"
loglocation="/var/contiv/log/netmaster.log"

while getopts ":vdl:b:p:c:" opt; do
    case $opt in
       b)
          binpath=$OPTARG
          ;;
       l)
          listen_url=$OPTARG
          ;;
       d)
          dns_enable=$OPTARG
          ;;
       c)
          cstore=$OPTARG
          ;; 
       p) 
          plugin=$OPTARG
          ;;
       v)
          debug=true
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
if [ "$listen_url" != "" ]; then
   listen_url_param="-listen_url"
fi

if [ $debug ]; then
    $binpath/netmaster -cluster-mode $plugin -dns-enable $dns_enable $cstore_param $cstore $listen_url_param $listen_url > $loglocation 2>&1
else
    $binpath/netmaster -cluster-mode $plugin -dns-enable $dns_enable $cstore_param $cstore $listen_url_param $listen_url -debug > $loglocation 2>&1
fi
