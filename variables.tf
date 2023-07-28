variable "cidr_block" {
  default     = "10.0.0.0/8"
  type        = string
  description = "CIDR block for the VPC"
}

variable "public_subnet_cidr_blocks" {
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
  type        = list(any)
  description = "List of public subnet CIDR blocks"
}

variable "private_subnet_cidr_blocks" {
  default     = ["10.0.101.0/24", "10.0.102.0/24"]
  type        = list(any)
  description = "List of private subnet CIDR blocks"
}

variable "availability_zones" {
  default     = ["us-west-2a", "us-west-2b"]
  type        = list(any)
  description = "List of availability zones"
}

variable "boundary_version" {
  default     = "0.12.2"
  type        = string
  description = "Boundary release version"
}

variable "aws_region" {
  default = "us-west-2"
  type    = string
}
variable "aws_profile" {
  type    = string
}


variable "hcp_boundary_cluster_id" {
  type        = string
  description = "HCP Boundary cluster ID"
}

variable "hcp_boundary_auth_method" {
  type        = string
  description = "Auth method ID from HCP Boundary cluster"
}

variable "hcp_boundary_username" {
  type = string
  description = "HCP Boundary cluster username"
}

variable "hcp_boundary_password" {
  type = string
  description = "HCP Boundary cluster password"
}

variable "worker_count" {
  type = number
  description = "Number of Boundary worker to spin up"
}