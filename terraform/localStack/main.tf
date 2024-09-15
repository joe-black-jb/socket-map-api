provider "aws" {
  access_key = "test"
  secret_key = "test"
  region = "ap-northeast-1"

  # only required for non virtual hosted-style endpoint use case.
  # https://registry.terraform.io/providers/hashicorp/aws/latest/docs#s3_use_path_style
  s3_use_path_style           = true
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    apigateway     = "http://localhost:4566"
    dynamodb       = "http://localhost:4566"
    lambda         = "http://localhost:4566"
    route53        = "http://localhost:4566"
    s3             = "http://s3.localhost.localstack.cloud:4566"
    sts            = "http://localhost:4566"
  }
}

# LocalStack用設定値確認用
data "aws_caller_identity" "current" {}
output "is_localstack" {
  value = data.aws_caller_identity.current.id == "000000000000"
}

# S3
resource "aws_s3_bucket" "test_bucket" {
  bucket = "ls-socket-map-bucket"
}

# バケットに JSON ファイルをアップロード
resource "aws_s3_bucket_object" "json_file" {
  bucket = aws_s3_bucket.test_bucket.bucket
  key    = "stations.json"
  source = "${path.module}/../../seeder/stations.json" # ローカルにある JSON ファイル
}

# Lambda 関数を作成 (ZIP ファイルからデプロイ)
resource "aws_lambda_function" "socket_map" {
  function_name = "ls-socket-map-api"
  handler       = "main"             # Go のエントリーポイント（Lambda の Handler）
  # runtime       = "provided.al2"     # Go ランタイムを使用
  runtime       = "go1.x"     # Go ランタイムを使用
  # role          = "arn:aws:iam::000000000000:role/lambda-exec-role"  # デフォルトの IAM ロール
  role          = aws_iam_role.lambda_execution_role.arn
  filename      = "main.zip"         # ZIP ファイルパス
  source_code_hash = filebase64sha256("main.zip")

  environment {
    variables = {
      BUCKET_NAME = aws_s3_bucket.test_bucket.bucket
      FILE_KEY    = aws_s3_bucket_object.json_file.key
    }
  }
}

# Lambda の権限設定
resource "aws_lambda_permission" "allow_s3_trigger" {
  statement_id  = "AllowExecutionFromS3"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.socket_map.function_name
  principal     = "s3.amazonaws.com"
}

# API Gateway (v2 の HTTP は pro のみ使用可能)
resource "aws_api_gateway_rest_api" "socket_map" {
  name = "ls-socket-map-api-gateway"
  body = jsonencode({
    openapi = "3.0.1"
    info = {
      title   = "example"
      version = "1.0"
    }
    paths = {
      "/" = {
        get = {
          x-amazon-apigateway-integration = {
            httpMethod           = "GET"
            payloadFormatVersion = "1.0"
            type                 = "HTTP_PROXY"
            uri                  = "https://ip-ranges.amazonaws.com/ip-ranges.json"
          }
        }
      }
    }
  })
  description = "API for Lambda function"
}

output "api_endpoint" {
  value = "http://localhost:4566/restapis/${aws_api_gateway_rest_api.socket_map.id}/stages/dev"
}

resource "aws_api_gateway_resource" "socket_map" {
  rest_api_id = aws_api_gateway_rest_api.socket_map.id
  parent_id   = aws_api_gateway_rest_api.socket_map.root_resource_id
  path_part   = "{path}"
}

resource "aws_api_gateway_method" "socket_map" {
  rest_api_id   = aws_api_gateway_rest_api.socket_map.id
  resource_id   = aws_api_gateway_resource.socket_map.id
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "socket_map" {
  rest_api_id             = aws_api_gateway_rest_api.socket_map.id
  resource_id             = aws_api_gateway_resource.socket_map.id
  http_method             = aws_api_gateway_method.socket_map.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:local:lambda:path/2015-03-31/functions/${aws_lambda_function.socket_map.arn}/invocations"
}

resource "aws_api_gateway_deployment" "socket_map" {
  rest_api_id = aws_api_gateway_rest_api.socket_map.id
  stage_name  = "dev"
}

# resource "aws_api_gateway_stage" "socket_map" {
#   stage_name = "dev"
#   rest_api_id = aws_api_gateway_rest_api.socket_map.id
#   deployment_id = aws_api_gateway_deployment.socket_map.id
# }

# Lambda 関数に API Gateway からのアクセスを許可
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.socket_map.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_api_gateway_rest_api.socket_map.execution_arn}/*/*"
}

# Lambda の IAM ロール
resource "aws_iam_role" "lambda_execution_role" {
  name = "lambda_execution_role"

  # assume_role_policy = jsonencode({
  #   Version = "2012-10-17"
  #   Statement = [
  #     {
  #       Action = "sts:AssumeRole"
  #       Effect = "Allow"
  #       Principal = {
  #         Service = "lambda.amazonaws.com"
  #       }
  #     }
  #   ]
  # })
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

# IAM ポリシーアタッチメント
resource "aws_iam_role_policy_attachment" "dynamodb_full_access" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess"
  role     = aws_iam_role.lambda_execution_role.name
}

resource "aws_iam_role_policy_attachment" "s3_full_access" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonS3FullAccess"
  role     = aws_iam_role.lambda_execution_role.name
}

resource "aws_iam_role_policy_attachment" "vpc_cross_account_network_interface_operations" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonVPCCrossAccountNetworkInterfaceOperations"
  role     = aws_iam_role.lambda_execution_role.name
}

resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role     = aws_iam_role.lambda_execution_role.name
}

resource "aws_iam_role_policy_attachment" "lambda_dynamodb_execution" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaDynamoDBExecutionRole"
  role     = aws_iam_role.lambda_execution_role.name
}

# resource "aws_iam_role_policy_attachment" "lambda_invocation_dynamodb" {
#   policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaInvocation-DynamoDB"
#   role     = aws_iam_role.lambda_execution_role.name
# }

resource "aws_iam_role_policy_attachment" "lambda_vpc_access_execution" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
  role     = aws_iam_role.lambda_execution_role.name
}

resource "aws_iam_role_policy_attachment" "secrets_manager_read_write" {
  policy_arn = "arn:aws:iam::aws:policy/SecretsManagerReadWrite"
  role     = aws_iam_role.lambda_execution_role.name
}
