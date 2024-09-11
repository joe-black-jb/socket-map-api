provider "aws" {
  region = "ap-northeast-1"  # 使用する AWS リージョンに変更してください
}

# Lambda 関数
resource "aws_lambda_function" "socket_map_api_public" {
  function_name = "socket-map-api-public"
  image_uri      = var.image_uri  # ECR にプッシュした Docker イメージ URI

  role          = var.lambda_role_arn
  package_type   = "Image"

  # Lambda のタイムアウト設定（オプション）
  timeout        = 60  # タイムアウト時間（秒）

  # 環境変数（オプション）
  environment {
    variables = {
      CF_DOMAIN = var.cf_domain
      ENV = var.env
      REGION = var.region
    }
  }
}