
进度 P214

# 1. 安装指南

## 1.1 按照proto生成.pb.go
```
protoc --go_out=plugins=grpc:. ./proto/*.proto
```

## 1.2 修改配置
```
main.go中
(1) port 为tag-service启动端口
(2) blogHost和blogPort为blog-service的地址和端口号
```

## 1.3 安装grpc工具
```
go get github.com/fullstorydev/grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl
cd ${GOPATH}/pkg/mod/github.com/fullstorydev/grpcurl@v1.7.0/cmd/grpcurl
sudo go build .
sudo mv grpcurl ${GOROOT}/bin/
```

## 1.4 利用反射测试GRPC调用
```
grpcurl -plaintext localhost:8080 proto.TagService.GetTagList
```
返回
```
{
  "list": [
    {
      "id": "1",
      "name": "tag-of-1",
      "state": 1
    }
  ],
  "pager": {
    "page": "1",
    "pageSize": "10",
    "totalRows": "1"
  }
}
```

## 1.5 安装TCP多协议支持工具
```
go get -u github.com/soheilhy/cmux@v0.1.4
```

## 1.6 安装双流量(RESTful json & GRPC)支持工具
```
sudo go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.14.4
cd ${GOPATH}/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.4/protoc-gen-grpc-gateway
sudo go build .
sudo mv protoc-gen-grpc-gateway ${GOROOT}/bin/
```