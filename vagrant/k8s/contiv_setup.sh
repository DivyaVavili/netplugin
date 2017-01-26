vagrant ssh k8master -c "kubectl -n kube-system delete deployment kube-dns"
vagrant ssh k8master -c "/shared/netctl net create -t default --subnet=20.1.1.0/24 default-net"
vagrant ssh k8master -c "/shared/netctl group create -t default default-net default-epg"
