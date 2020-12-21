package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	pb "tag-service/proto"
)

var host string
var port string

func init() {
	flag.StringVar(&host, "host", "localhost", "目标服务地址")
	flag.StringVar(&port, "port", "8080", "目标服务端口")
	flag.Parse()
}

func main() {
	ctx := context.Background()
	clientConn, err := GetClientConn(ctx, GetTargetURL(), nil)
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	defer clientConn.Close()

	tagServiceClient := pb.NewTagServiceClient(clientConn)
	resp, err := tagServiceClient.GetTagList(ctx, &pb.GetTagListRequest{
		Name:                 "Go",
	})
	if err != nil {
		log.Fatalf("tagServiceClient.GetTagList err: %v", err)
	}
	respJson, err := json.Marshal(&resp)
	if err != nil {
		log.Fatalf("marshal json fail, err: %v", err)
	}
	log.Printf("resp: %v", string(respJson))
}

func GetClientConn(ctx context.Context, target string, opts []grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts, grpc.WithInsecure())
	return grpc.DialContext(ctx, target, opts...)
}

func GetTargetURL() string {
	return fmt.Sprintf("%s:%s", host, port)
}