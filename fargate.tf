resource "aws_ecr_repository" "main" {
  name                 = "boundary-worker"
  image_tag_mutability = "MUTABLE"
  force_delete         = true
}

resource "aws_ecr_lifecycle_policy" "main" {
  repository = aws_ecr_repository.main.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "keep last 10 images"
        action       = {
          type = "expire"
        }
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
      }
    ]
  })
}

resource "aws_ecs_cluster" "boundary-worker-clusters" {
  name = "boundary-workers-cluster"

}

resource "aws_iam_role" "fargate_execution_role" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Principal": {"Service": "ecs-tasks.amazonaws.com"},
    "Action": "sts:AssumeRole"
  }
}
EOF
}

# Create IAM Role Policy Attachment
resource "aws_iam_role_policy_attachment" "fargate_execution_role" {
  role       = aws_iam_role.fargate_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_cloudwatch_log_group" "fargate_boundary_worker" {
  name              = "/fargate/service/boundary-workers"
  retention_in_days = 3
}

resource "aws_ecs_task_definition" "boundary_worker_task_def" {
  network_mode             = "awsvpc"
  family                   = "boundary-worker"
  requires_compatibilities = ["FARGATE"]
  execution_role_arn       = "${aws_iam_role.fargate_execution_role.arn}"
  cpu                      = 256
  memory                   = 512
  container_definitions    = jsonencode([
    {
      name        = "boundary-worker"
      image       = "${aws_ecr_repository.main.repository_url}:latest"
      essential   = true
      environment = [
        {
          name  = "HCP_BOUNDARY_CLUSTER_ID"
          value = "${var.hcp_boundary_cluster_id}"
        },
        {
          name : "ECS_ENABLE_CONTAINER_METADATA"
          value : "true"
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options   = {
          awslogs-group         = aws_cloudwatch_log_group.fargate_boundary_worker.name
          awslogs-stream-prefix = "ecs"
          awslogs-region        = "${var.aws_region}"
        }
      }

    }
  ])
}

resource "aws_ecs_cluster_capacity_providers" "capacity_provider" {
  cluster_name       = aws_ecs_cluster.boundary-worker-clusters.name
  capacity_providers = ["FARGATE"]
}

resource "aws_ecs_service" "boundary_worker" {
  name                    = "boundary_worker"
  cluster                 = aws_ecs_cluster.boundary-worker-clusters.id
  task_definition         = aws_ecs_task_definition.boundary_worker_task_def.arn
  launch_type             = "FARGATE"
  enable_ecs_managed_tags = true
  desired_count           = var.worker_count
  network_configuration {
    subnets = [for s in aws_subnet.private : s.id]
  }
}

resource "aws_cloudwatch_log_subscription_filter" "worker_auth_subscription" {
  name            = "worker_auth_watcher_subscription"
  log_group_name  = aws_cloudwatch_log_group.fargate_boundary_worker.name
  filter_pattern  = "\"Worker Auth Registration Request: \""
  destination_arn = aws_lambda_function.worker_auth_watcher_lambda.arn
}


resource "aws_cloudwatch_event_rule" "fargate_task_stopped" {
  name          = "fargate-stop-task-watcher"
  event_pattern = jsonencode({
    source = [
      "aws.ecs"
    ]
    detail-type = [
      "ECS Task State Change"
    ]
    detail = {
      lastStatus : ["STOPPED"]
      clusterArn : [aws_ecs_cluster.boundary-worker-clusters.arn]
    }
  })
}

resource "aws_cloudwatch_event_target" "fargate_stop_event_lambda_target" {
  depends_on = [aws_lambda_function.worker_stop_watcher_lambda]
  arn  = aws_lambda_function.worker_stop_watcher_lambda.arn
  rule = aws_cloudwatch_event_rule.fargate_task_stopped.name
}
