package main

import (
	"flag"
	"fmt"
	"github.com/soheilhy/cmux"
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

// tag-service多协议共用端口号
var port string

// 多协议运行方式
var way int

const (
	WAY_GRPC_HTTP_SEPERATELY = 1
	WAY_MULTI_ON_TCP = 2
)

func init() {
	flag.StringVar(&grpcPort, "grpc-port", "8080", "GRPC启动端口号")
	flag.StringVar(&httpPort, "http-port", "8081", "HTTP启动端口号")
	flag.StringVar(&blogHost, "blog-host", "localhost", "博客服务主机地址")
	flag.StringVar(&blogPort, "blog-port", "8001", "博客服务端口号")
	flag.StringVar(&port, "port", "8082", "多协议共用端口号")
	flag.IntVar(&way, "way", WAY_GRPC_HTTP_SEPERATELY, "多协议运行方式")
	flag.Parse()
}

func main() {
	switch way {
	case WAY_GRPC_HTTP_SEPERATELY:
		RunSeperately()
	case WAY_MULTI_ON_TCP:
		RunMultiOnTCP()
	default:
		log.Fatalf("unknown way:%d to run!", way)
	}

}

func RunSeperately() {
	log.Println("start to RunSeperately")
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

func RunMultiOnTCP() {
	log.Println("start to RunMultiOnTCP")
	l, err := RunTCPServer(port)
	if err != nil {
		log.Fatalf("Run TCP Server err: %v", err)
	}

	m := cmux.New(l)
	grpcL := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	httpL := m.Match(cmux.HTTP1Fast())

	grpcS := RunGrpcServerOnTCP()
	httpS := RunHttpServerOnTCP(port)
	go grpcS.Serve(grpcL)
	go httpS.Serve(httpL)

	err = m.Serve()
	if err != nil {
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

func RunTCPServer(port string) (net.Listener, error) {
	return net.Listen("tcp", ":" + port)
}

func RunGrpcServerOnTCP() *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterTagServiceServer(s, server.NewTagServer(GetBlogURL()))
	reflection.Register(s)

	return s
}

func RunHttpServerOnTCP(port string) *http.Server {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`pong`))
	})

	return &http.Server{
		Addr:              ":"+port,
		Handler:           serveMux,
	}
}