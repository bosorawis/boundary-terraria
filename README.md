# Boundary sandbox

Minimal setup for testing out HCP Boundary multi-hop feature with SSH credentials injection


## Pre-req
- Terraform
- AWS account with CLI access
    - configure AWS profile locally
- running [HCP Boundary cluster](https://portal.cloud.hashicorp.com/services/boundary/)
## Provisioning 

Provisioning AWS resources for the sandbox

1. VPC with private and public subnet
2. EC2 instance for running `boundary-worker` inside the private subnet
3. EC2 instance acting as a private ssh target

This way, both `boundary-worker` and `SSH target` are not reachable from the public internet (no public IP)

![Diagram](./img/network.drawio.png)

### Create TFVARS

Create tfvars file to hold sensitive information

```bash
# dev.tfvars
hcp_boundary_auth_method = "<HCP boundary auth method ID>"
hcp_boundary_username    = "<username>"
hcp_boundary_password    = "<password>"
hcp_boundary_cluster_id  = "<cluster>"

aws_profile = "<aws-profile-name>"
aws_region = "us-west-2"
availability_zones = ["us-west-2a", "us-west-2b"]
```


### Build Lambda apps

```bash
make build
```


### Provision resource

**Note**: This step and [Build and push container image](#Build-and-push-container-image) must back-to-back so don't look away!

- run `terraform init`

```bash
terraform plan -var-file=envs/dev.tfvars
terraform apply -var-file=envs/dev.tfvars
# Follow the prompt
```

### Build and push container image

At this point, all resources are created in AWS; however, there's no available docker image for the `boundary-worker` yet.
To build and push the image, first find out what's the repository URL of the created ECR

```bash

# fetch ECR repo URL
terraform output 
# ecr = "<account>.dkr.ecr.us-west-2.amazonaws.com/boundary-worker"

# login to ECR
aws ecr get-login-password --region <region> --profile <profile-name> | docker login --username AWS --password-stdin <ecr-repo-url>

# build and push
make docker

docker tag boundary-worker:latest <ecr-repo-url>:latest
docker push <ecr-repo-url>:latest
```

### Validate that the service is _up_

1. Login to AWS console
2. Navigate to `Elastic Container Service`
3. Navigate to `Clusters`
4. Navigate to `boundary-worker-cluster`
5. Click on `Services` tab and click into `boundary_worker`
6. Go to `Tasks` tab
7. There should be 1 running task
8. Optionally login to Boundary cluster under `Workers` tab and validate that there is now an active worker


## Connect

Install [Boundary client](https://developer.hashicorp.com/boundary/tutorials/oss-getting-started/oss-getting-started-desktop-app)

Login to your Boundary cluster in the client

Select "my first ssh target" and click `Connect`

run `ssh localhost -p <output-port>`
