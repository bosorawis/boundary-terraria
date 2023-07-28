resource "boundary_scope" "global" {
  global_scope = true
  scope_id     = "global"
}

resource "boundary_scope" "org" {
  name                     = "my first org"
  description              = "My first scope!"
  scope_id                 = boundary_scope.global.id
  auto_create_admin_role   = true
  auto_create_default_role = true
}

resource "boundary_scope" "project" {
  name                   = "my first project"
  description            = "My first project!"
  scope_id               = boundary_scope.org.id
  auto_create_admin_role = true
}

resource "boundary_credential_store_static" "ssh_creds_store" {
  name        = "My static creds store"
  description = "My first static credential store!"
  scope_id    = boundary_scope.project.id
}

resource "boundary_credential_ssh_private_key" "my_private_key" {
  name                = "my private key"
  description         = "My first ssh private key credential!"
  credential_store_id = boundary_credential_store_static.ssh_creds_store.id
  username            = "ubuntu"
  private_key         = tls_private_key.ssh_key.private_key_pem
}

resource "boundary_target" "foo" {
  name                 = "Terraria Server - SSH"
  description          = "Used for SSH-ing into Terraria Server"
  type                 = "ssh"
  default_port         = "22"
  scope_id             = boundary_scope.project.id
  address              = aws_instance.ssh-target.private_ip
  egress_worker_filter = "\"downstream\" in \"/tags/type\""
  injected_application_credential_source_ids = [
    boundary_credential_ssh_private_key.my_private_key.id
  ]
}


resource "boundary_target" "terraria_game_target" {
  name                 = "Terraria Server - Gameplay"
  description          = "Used for connecting to Terraria Server with Terraria Client"
  type                 = "tcp"
  default_port         = "7777"
  scope_id             = boundary_scope.project.id
  address              = aws_instance.ssh-target.private_ip
  egress_worker_filter = "\"downstream\" in \"/tags/type\""
}