package main

import (
	"context"
	"flag"
	"fmt"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"path"
	"tag-service/internal/middleware"
	"tag-service/pkg/swagger"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/soheilhy/cmux"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"strings"
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

// tag-service同端口同方法实现双流量支持
var multiPort string

// 多协议运行方式
var way int

const (
	WayGrpcHttpSeparately = 1 // 多个端口分别处理HTTP/1.1和HTTP/2请求
	WayGrpcHttpOnTcp      = 2 // 同一个端口处理HTTP/1.1和HTTP/2请求, 但由不同的RPC方法处理
	WayGrpcHttpOnSameRpc  = 3 // 同一个端口处理HTTP/1.1和HTTP/2请求, 且由相同的RPC方法处理
)

func init() {
	flag.StringVar(&grpcPort, "grpc-port", "8080", "GRPC启动端口号")
	flag.StringVar(&httpPort, "http-port", "8081", "HTTP启动端口号")
	flag.StringVar(&blogHost, "blog-host", "localhost", "博客服务主机地址")
	flag.StringVar(&blogPort, "blog-port", "8001", "博客服务端口号")
	flag.StringVar(&port, "port", "8082", "多协议共用端口号")
	flag.StringVar(&multiPort, "multi_port", "8083", "多协议共用端口号和同方法")
	flag.IntVar(&way, "way", WayGrpcHttpSeparately, "多协议运行方式")
	flag.Parse()
}

func main() {
	switch way {
	case WayGrpcHttpSeparately:
		RunSeperately()
	case WayGrpcHttpOnTcp:
		RunMultiOnTCP()
	case WayGrpcHttpOnSameRpc:
		RunMultiOnSameRPC()
	default:
		log.Fatalf("unknown way:%d to run!", way)
	}
}

func RunSeperately() {
	log.Println("start to RunSeperately")
	errs := make(chan error)
	go func() {
		serveMux := http.NewServeMux()
		serveMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`pong`))
		})

		prefix := "/swagger-ui/"
		fileServer := http.FileServer(&assetfs.AssetFS{
			Asset:     swagger.Asset,
			AssetDir:  swagger.AssetDir,
			AssetInfo: nil,
			Prefix:    "third_party/swagger",
			Fallback:  "",
		})
		serveMux.Handle(prefix, http.StripPrefix(prefix, fileServer))
		serveMux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasSuffix(r.URL.Path, "swagger.json") {
				http.NotFound(w, r)
				return
			}

			p := strings.TrimPrefix(r.URL.Path, "/swagger/")
			p = path.Join("proto", p)

			http.ServeFile(w, r, p)
		})

		err := http.ListenAndServe(":" + httpPort, serveMux)
		if err != nil {
			errs <- err
			return
		}
	}()

	go func() {
		opts := []grpc.ServerOption{
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
				helloInterceptor,
				worldInterceptor,
				middleware.AccessLog,
				middleware.ErrorLog,
				middleware.Recovery,
			)),
		}
		s := grpc.NewServer(opts...)
		pb.RegisterTagServiceServer(s, server.NewTagServer(GetBlogURL()))
		reflection.Register(s)
		listener, err := net.Listen("tcp", ":" + grpcPort)
		if err != nil {
			errs <- err
			return
		}
		err = s.Serve(listener)
		if err != nil {
			errs <- err
			return
		}
	}()
	select {
	case err := <-errs:
		log.Fatalf("Run Server err: %v", err)
	}
}

func RunMultiOnTCP() {
	log.Println("start to RunMultiOnTCP")
	l, err := net.Listen("tcp", ":" + port)
	if err != nil {
		log.Fatalf("Run TCP Server err: %v", err)
	}

	m := cmux.New(l)
	grpcL := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	httpL := m.Match(cmux.HTTP1Fast())

	grpcS := grpc.NewServer()
	pb.RegisterTagServiceServer(grpcS, server.NewTagServer(GetBlogURL()))
	reflection.Register(grpcS)

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`pong`))
	})
	httpS := &http.Server{
		Addr:              ":"+port,
		Handler:           serveMux,
	}

	go grpcS.Serve(grpcL)
	go httpS.Serve(httpL)

	err = m.Serve()
	if err != nil {
		log.Fatalf("Run Server err: %v", err)
	}
}

func RunMultiOnSameRPC() {
	grpcS := grpc.NewServer()
	pb.RegisterTagServiceServer(grpcS, server.NewTagServer(GetBlogURL()))
	reflection.Register(grpcS)

	gatewayMux := runtime.NewServeMux()
	dopts := []grpc.DialOption{grpc.WithInsecure()}
	_ = pb.RegisterTagServiceHandlerFromEndpoint(context.Background(), gatewayMux, "0.0.0.0:" + multiPort, dopts)

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`pong`))
	})
	httpMux.Handle("/", gatewayMux)

	if err := http.ListenAndServe(":"+multiPort, grpcHandlerFunc(grpcS, httpMux)); err != nil {
		log.Fatalf("Run Server err: %v", err)
	}
}

func GetBlogURL() string {
	return fmt.Sprintf("http://%s:%s", blogHost, blogPort)
}

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

func helloInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	log.Println("你好")
	resp, err := handler(ctx, req)
	log.Println("再见")
	return resp, err
}

func worldInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	log.Println("世界我来了")
	resp, err := handler(ctx, req)
	log.Println("世界再见")
	return resp, err
}