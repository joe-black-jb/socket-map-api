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

resource "aws_dynamodb_table" "places" {
  name = "places"
  billing_mode = "PAY_PER_REQUEST"
  hash_key = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "address"
    type = "S"
  }

  attribute {
    name = "businessHours"
    type = "S"
  }

  attribute {
    name = "created_at"
    type = "S"
  }

  attribute {
    name = "image"
    type = "S"
  }

  attribute {
    name = "latitude"
    type = "N"
  }

  attribute {
    name = "longitude"
    type = "N"
  }

  attribute {
    name = "memo"
    type = "S"
  }

  attribute {
    name = "name"
    type = "S"
  }

  attribute {
    name = "smoke"
    type = "N"
  }

  attribute {
    name = "socket"
    type = "N"
  }

  attribute {
    name = "socketNum"
    type = "N"
  }

  attribute {
    name = "tel"
    type = "S"
  }

  attribute {
    name = "updated_at"
    type = "S"
  }

  attribute {
    name = "url"
    type = "S"
  }

  attribute {
    name = "wifi"
    type = "N"
  }

  global_secondary_index {
    name               = "LatitudeLongitudeIndex"
    hash_key           = "latitude"
    range_key          = "longitude"
    projection_type    = "ALL"
  }
}