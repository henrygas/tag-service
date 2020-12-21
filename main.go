package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	pb "tag-service/proto"
	"tag-service/server"
)

var port string
var blogHost string
var blogPort string

func init() {
	flag.StringVar(&port, "port", "8080", "启动端口号")
	flag.StringVar(&blogHost, "blog-host", "localhost", "博客服务主机地址")
	flag.StringVar(&blogPort, "blog-port", "8001", "博客服务端口号")
	flag.Parse()
}

func main() {
	blogURL := GetBlogURL()

	s := grpc.NewServer()
	pb.RegisterTagServiceServer(s, server.NewTagServer(blogURL))
	reflection.Register(s)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
	}

	err = s.Serve(listener)
	if err != nil {
		log.Fatalf("server.Serve err: %v", err)
	}
}


func GetBlogURL() string {
	return fmt.Sprintf("http://%s:%s", blogHost, blogPort)
}