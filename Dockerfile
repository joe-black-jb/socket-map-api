FROM golang:1.22.3-alpine As build

# Install git
RUN apk add --no-cache git

WORKDIR /app

# Copy dependencies list
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify
# RUN go mod tidy && go mod verify
# RUN go mod download && go get github.com/joe-black-jb/socket-map-api/internal/api && go get github.com/joe-black-jb/socket-map-api/internal/database

# Build with optional lambda.norpc tag
# COPY cmd/socket-map-api/main.go .
COPY . .
# COPY main.go .

# RUN go build -tags lambda.norpc -o main main.go
RUN go build -tags lambda.norpc -o main cmd/socket-map-api/main.go

# Copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2023

COPY --from=build /app/main ./main

ENTRYPOINT [ "./main" ]