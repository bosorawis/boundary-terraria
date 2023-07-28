output "ecr" {
  description = "ID of project VPC"
  value       = aws_ecr_repository.main.repository_url
}
