# Wi-Fi, 電源があるカフェ検索アプリの API

## ローカル開発環境構築

```sh
# Docker (localstack)
make up

# LocalStack 環境に tflocal で AWS リソース構築
## Lambdaに適用するファイルのビルド（ルートディレクトリで実行）
make zip

## tflocal apply
make terraform

# リソース構築状況確認
aws --endpoint-url=http://localhost:4566 s3 ls
aws --endpoint-url=http://localhost:4566 lambda get-function --function-name ls-socket-map-api

# Lambda 実行
aws lambda --endpoint-url=http://localhost:4566 invoke --function-name ls-socket-map-api result.log

```
