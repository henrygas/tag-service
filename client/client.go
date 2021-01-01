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

type Auth struct {
	AppKey string
	AppSecret string
}

func (a *Auth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"app_key": a.AppKey,
		"app_secret": a.AppSecret,
	}, nil
}

func (a *Auth) RequireTransportSecurity() bool {
	return false
}

func main() {
	auth := Auth{
		AppKey:    "henry-key",
		AppSecret: "henry-secret",
	}

	ctx := context.Background()
	var opts = []grpc.DialOption{grpc.WithPerRPCCredentials(&auth)}
	clientConn, err := GetClientConn(ctx, GetTargetURL(), opts)
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
	//opts = append(opts, grpc.WithUnaryInterceptor(
	//	grpc_middleware.ChainUnaryClient(
	//		grpc_retry.UnaryClientInterceptor(
	//			grpc_retry.WithMax(5),
	//			grpc_retry.WithCodes(
	//				codes.Unknown,
	//				codes.Internal,
	//				codes.DeadlineExceeded,
	//			),
	//		),
	//		middleware.UnaryContextTimeout(),
	//	),
	//))
	//opts = append(opts, grpc.WithStreamInterceptor(
	//	grpc_middleware.ChainStreamClient(
	//		middleware.StreamContextTimeout(),
	//	),
	//))
	return grpc.DialContext(ctx, target, opts...)
}

func GetTargetURL() string {
	return fmt.Sprintf("%s:%s", host, port)
}