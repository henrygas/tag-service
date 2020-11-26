

# tag-service
grpc demo working with blog-service

## 按照proto生成.pb.go
protoc -I /usr/local/include --proto_path=.  --go_out=plugins=grpc:. ./proto/*.proto