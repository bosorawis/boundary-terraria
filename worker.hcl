name = "fargate-worker"
disable_mlock = true
listener "tcp" {
  purpose = "proxy"
  address = "0.0.0.0:9200"
}

hcp_boundary_cluster_id = "%%CLUSTERID%%"

worker {
  auth_storage_path = "/boundary/auth/worker1"
  tags {
    taskid = ["%%TASKID%%"]
    type = ["worker1", "downstream"]
  }
}