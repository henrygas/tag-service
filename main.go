package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	pb "tag-service/proto"
	"tag-service/server"
)

// 目标博客服务的地址
var blogHost string
var blogPort string

// tag-service本身监听的端口号
var grpcPort string
var httpPort string

func init() {
	flag.StringVar(&grpcPort, "grpc-port", "8080", "GRPC启动端口号")
	flag.StringVar(&httpPort, "http-port", "8081", "HTTP启动端口号")
	flag.StringVar(&blogHost, "blog-host", "localhost", "博客服务主机地址")
	flag.StringVar(&blogPort, "blog-port", "8001", "博客服务端口号")
	flag.Parse()
}

func main() {
	errs := make(chan error)
	go func() {
		err := RunHttpServer(httpPort)
		if err != nil {
			errs <- err
		}
	}()

	go func() {
		err := RunGrpcServer(grpcPort)
		if err != nil {
			errs <- err
		}
	}()
	select {
	case err := <-errs:
		log.Fatalf("Run Server err: %v", err)
	}
}


func GetBlogURL() string {
	return fmt.Sprintf("http://%s:%s", blogHost, blogPort)
}

func RunHttpServer(port string) error {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`pong`))
	})
	return http.ListenAndServe(":" + port, serveMux)
}

func RunGrpcServer(port string) error {
	s := grpc.NewServer()
	pb.RegisterTagServiceServer(s, server.NewTagServer(GetBlogURL()))
	reflection.Register(s)
	listener, err := net.Listen("tcp", ":" + port)
	if err != nil {
		return err
	}
	return s.Serve(listener)
}