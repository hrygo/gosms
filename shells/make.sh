#!/bin/sh

go clean
go mod tidy

# 如果你想在Windows 32位系统下运行
# CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -trimpath ../cmd/server/gosmsn.go

# 如果你想在Windows 64位系统下运行
# CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath ../cmd/server/gosmsn.go

# 如果你想在Linux 32位系统下运行
# CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -trimpath ../cmd/server/gosmsn.go

# 如果你想在Linux 64位系统下运行
# CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath ../cmd/server/gosmsn.go

# 如果你想在Linux arm64系统下运行
# CGO_ENABLED=0 GOOS=linux GOARM=7 GOARCH=arm64 go build -trimpath ../cmd/server/gosmsn.go

# 如果你想在 本机环境 运行
go build -trimpath ../cmd/server/gosmsn.go

# 制作软件发布包
chmod +x gosmsn
chmod +x start.sh
cp -rf ../config ./
tar -zcvf gosmsn.tar.gz gosmsn start.sh config
rm -rf ./config