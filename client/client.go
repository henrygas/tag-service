package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"log"
	"tag-service/internal/middleware"
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
	opts = append(opts, grpc.WithUnaryInterceptor(
		grpc_middleware.ChainUnaryClient(
			middleware.UnaryContextTimeout(),
			grpc_retry.UnaryClientInterceptor(
				grpc_retry.WithMax(2),
				grpc_retry.WithCodes(
					codes.Unknown,
					codes.Internal,
					codes.DeadlineExceeded,
				),
			),
		),
	))
	opts = append(opts, grpc.WithStreamInterceptor(
		grpc_middleware.ChainStreamClient(
			middleware.StreamContextTimeout(),
		),
	))
	return grpc.DialContext(ctx, target, opts...)
}

func GetTargetURL() string {
	return fmt.Sprintf("%s:%s", host, port)
}