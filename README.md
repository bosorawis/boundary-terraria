# Welcome to the most paranoid Terraria server


## Pre-req
- Terraform
- aws account
- running HCP Boundary cluster
## Provisioning 

### Create TFVARS

Create tfvars file to hold sensitive information

```bash
# dev.tfvars
hcp_boundary_auth_method = "<HCP boundary auth method ID>"
hcp_boundary_username    = "<username>"
hcp_boundary_password    = "<password>"
hcp_boundary_cluster_id  = "<cluster>"

aws_profile = "personal"
aws_region = "us-west-2"
```

### Provision resource

- run `terraform init`


```bash
terraform plan -var-file=envs/personal.tfvars
terraform apply -var-file=envs/personal.tfvars
# Follow the prompt
```


