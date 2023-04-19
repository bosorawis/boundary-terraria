resource "aws_instance" "ssh-target" {
  ami                  = data.aws_ami.ubuntu.id
  instance_type        = "t3.micro"
  iam_instance_profile = aws_iam_instance_profile.ec2_profile.name
  subnet_id            = aws_subnet.private[1].id
  key_name             = aws_key_pair.generated_key.key_name
  tags = {
    Name = "ssh-target"
  }
}

locals {
  boundary_version           = var.boundary_version
  cluster_id                 = var.hcp_boundary_cluster_id
  controller_generated_token = boundary_worker.byow_worker.controller_generated_activation_token
}

resource "aws_instance" "boundary-server" {
  depends_on           = [boundary_worker.byow_worker]
  ami                  = data.aws_ami.ubuntu.id
  instance_type        = "t3.micro"
  iam_instance_profile = aws_iam_instance_profile.ec2_profile.name
  subnet_id            = aws_subnet.private[0].id
  tags = {
    Name = "boundary-server"
  }

  user_data_base64 = base64encode("${templatefile("${path.module}/boundary.sh", {
    boundary_version           = local.boundary_version
    cluster_id                 = local.cluster_id
    controller_generated_token = local.controller_generated_token
  })}")

  lifecycle {
    ignore_changes = [
      user_data_base64,
    ]
  }

}


resource "tls_private_key" "ssh_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "generated_key" {
  key_name_prefix = "terraform-ssh-key"
  public_key      = tls_private_key.ssh_key.public_key_openssh
}


# Create IAM Role
resource "aws_iam_role" "instance_role" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {"Service": "ec2.amazonaws.com"},
    "Action": "sts:AssumeRole"
  }
}
EOF
}

resource "aws_iam_instance_profile" "ec2_profile" {
  role = aws_iam_role.instance_role.name
}

# Create IAM Role Policy Attachment
resource "aws_iam_role_policy_attachment" "policy_attachment" {
  role       = aws_iam_role.instance_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}


