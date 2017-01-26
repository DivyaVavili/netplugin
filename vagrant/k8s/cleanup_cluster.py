vagrant ssh k8master -c "sudo kubeadm reset --skip-preflight-checks"
vagrant ssh k8node-01 -c "sudo kubeadm reset --skip-preflight-checks"
vagrant ssh k8node-02 -c "sudo kubeadm reset --skip-preflight-checks"
vagrant ssh k8node-03 -c "sudo kubeadm reset --skip-preflight-checks"

# init_out=$(vagrant ssh k8master -c "sudo kubeadm init --skip-preflight-checks --api-advertise-addresses 192.168.2.10")
# vagrant ssh k8master -c "kubectl apply -f /shared/.contiv.yaml"
# vagrant ssh k8master -c "kubectl -n kube-system delete deployment kube-dns"
# 
# echo $init_out
