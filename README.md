
进度 P242

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
# 安装grpc-gateway
sudo go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.14.4
cd ${GOPATH}/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.4/protoc-gen-grpc-gateway
sudo go build .
sudo mv protoc-gen-grpc-gateway ${GOROOT}/bin/

# 修改和编译proto文件
protoc -I/usr/local/include -I. \
 -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis\
 --grpc-gateway_out=logtostderr=true:. ./proto/*.proto

# 测试方法
grpcurl -plaintext localhost:8083 proto.TagService.GetTagList
curl -v http://localhost:8083/api/v1/tags
```

## 1.7 安装接口文档生成工具swagger
```
# 安装protoc插件protoc-gen-swagger: 通过proto文件生成swagger定义(swagger.json)
sudo go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger

# 下载swagger UI源码
git clone https://github.com/swagger-api/swagger-ui.git
cp -rf swagger-ui/dist/* third_party/swagger/

# 安装go-bindata插件：将swagger UI的资源文件转换为go代码
sudo go get -u github.com/go-bindata/go-bindata/...

# 在项目根目录下进行资源转换
go-bindata --nocompress -pkg swagger -o pkg/swagger/data.go third_party/swagger/...

# 安装go-bindata-assetfs库：结合net/http标准库和go-bindata生成swagger UI的GO代码供外部访问
sudo go get -u github.com/elazarl/go-bindata-assetfs/...

# 生成swagger.json
protoc -I/usr/local/include -I. -I${GOPATH}/src \
 -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
 --swagger_out=logtostderr=true:. ./proto/*.proto
```
