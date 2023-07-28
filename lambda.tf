data "archive_file" "worker_auth_watcher_zip" {
  type        = "zip"
  source_file = "./bin/worker-auth-watcher"
  output_path = "./bin/worker-auth-watcher.zip"
}

resource "aws_lambda_function" "worker_auth_watcher_lambda" {
  function_name    = "worker-auth-watcher"
  filename         = "./bin/worker-auth-watcher.zip"
  handler          = "worker-auth-watcher"
  source_code_hash = data.archive_file.worker_auth_watcher_zip.output_base64sha256
  role             = aws_iam_role.lambda_execution_role.arn
  runtime          = "go1.x"
  memory_size      = 128
  timeout          = 10
  environment {
    variables = {
      CLUSTER_URL             = "https://${var.hcp_boundary_cluster_id}.boundary.hashicorp.cloud"
      BOUNDARY_USERNAME       = var.hcp_boundary_username
      BOUNDARY_PASSWORD       = var.hcp_boundary_password
      BOUNDARY_AUTH_MATHOD_ID = var.hcp_boundary_auth_method
      TABLE_NAME              = aws_dynamodb_table.dynamo.name
    }
  }
}

resource "aws_lambda_permission" "allow_cloudwatch" {
  statement_id  = "AllowCloudwatchLogsInvokeLambda"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.worker_auth_watcher_lambda.function_name
  principal     = "logs.${var.aws_region}.amazonaws.com"
  source_arn    = "${aws_cloudwatch_log_group.fargate_boundary_worker.arn}:*"
}

data "archive_file" "worker_stop_watcher_zip" {
  type        = "zip"
  source_file = "./bin/worker-stop-watcher"
  output_path = "./bin/worker-stop-watcher.zip"
}

resource "aws_lambda_function" "worker_stop_watcher_lambda" {
  function_name    = "worker-stop-watcher"
  filename         = "./bin/worker-stop-watcher.zip"
  handler          = "worker-stop-watcher"
  source_code_hash = data.archive_file.worker_stop_watcher_zip.output_base64sha256
  role             = aws_iam_role.lambda_execution_role.arn
  runtime          = "go1.x"
  memory_size      = 128
  timeout          = 10
  environment {
    variables = {
      CLUSTER_URL             = "https://${var.hcp_boundary_cluster_id}.boundary.hashicorp.cloud"
      BOUNDARY_USERNAME       = var.hcp_boundary_username
      BOUNDARY_PASSWORD       = var.hcp_boundary_password
      BOUNDARY_AUTH_MATHOD_ID = var.hcp_boundary_auth_method
      TABLE_NAME              = aws_dynamodb_table.dynamo.name
    }
  }
}

resource "aws_lambda_permission" "allow_event_bridge" {
  statement_id  = "AllowEventBridgeInvokeLambda"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.worker_stop_watcher_lambda.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.fargate_task_stopped.arn
}



resource "aws_dynamodb_table" "dynamo" {
  name         = "BoundaryWorkers"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "taskID"
  attribute {
    name = "taskID"
    type = "S"
  }
}

resource "aws_iam_role" "lambda_execution_role" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}


resource "aws_iam_role_policy" "dynamodb_lambda_policy" {
  name   = "lambda-dynamodb-policy"
  role   = aws_iam_role.lambda_execution_role.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Sid": "APIAccessForDynamoDB",
        "Effect": "Allow",
        "Action": [
            "dynamodb:DeleteItem",
            "dynamodb:PutItem",
            "dynamodb:GetItem"
        ],
        "Resource": "${aws_dynamodb_table.dynamo.arn}"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "lambda_execution_policy_attachment" {
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}


