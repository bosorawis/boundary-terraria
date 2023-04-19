#!/bin/bash

mkdir /boundary
cd /boundary
sudo apt update; sudo apt upgrade -y;
sudo apt-get install unzip;
wget -q https://releases.hashicorp.com/boundary-worker/${boundary_version}+hcp/boundary-worker_${boundary_version}+hcp_linux_amd64.zip;
unzip boundary-worker_${boundary_version}+hcp_linux_amd64.zip
cat <<EOT >> worker.hcl
name = "private-worker"
disable_mlock = true
listener "tcp" {
    purpose = "proxy"
    address = "0.0.0.0:9200"
}

hcp_boundary_cluster_id = "${cluster_id}"

worker {
    auth_storage_path = "/boundary/auth/worker1"
    controller_generated_activation_token = "${controller_generated_token}"
    tags {
        type = ["worker1", "downstream"]
    }
}
EOT

sudo screen -dmS boundary ./boundary-worker server -config=./worker.hcl