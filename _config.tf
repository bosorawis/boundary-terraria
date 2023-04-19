# Configure the AWS Provider
provider "aws" {
  region  = var.aws_region
  profile = var.aws_profile
}


provider "boundary" {
  addr                            = "https://${var.hcp_boundary_cluster_id}.boundary.hashicorp.cloud/"
  auth_method_id                  = var.hcp_boundary_auth_method
  password_auth_method_login_name = var.hcp_boundary_username
  password_auth_method_password   = var.hcp_boundary_password
}
