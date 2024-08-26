# Build Docker image
docker build --platform linux/amd64 -t $ECR_BASE_URI/$ECR_NAME:$IMAGE_TAG .

# Login to ECR
aws ecr get-login-password --region ap-northeast-1 | docker login --username AWS --password-stdin $ECR_BASE_URI

# Push Docker image to ECR
docker push $ECR_BASE_URI/$ECR_NAME:$IMAGE_TAG

# Update Lambda function code
aws lambda update-function-code --function-name $LAMBDA_FUNCTION_NAME --image-uri $ECR_BASE_URI/$ECR_NAME:$IMAGE_TAG