package server

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc/metadata"
	"tag-service/pkg/bapi"
	"tag-service/pkg/errcode"
	pb "tag-service/proto"
	"time"
)

type Auth struct {

}

func (a *Auth) GetAppKey() string {
	return "henry-key"
}

func (a *Auth) GetAppSecret() string {
	return "henry-secret"
}

func (a *Auth) Check(ctx context.Context) error {
	md, _ := metadata.FromIncomingContext(ctx)

	var appKey, appSecret string
	if value, ok := md["app_key"]; ok {
		appKey = value[0]
	}
	if value, ok := md["app_secret"]; ok {
		appSecret = value[0]
	}
	if appKey != a.GetAppKey() || appSecret != a.GetAppSecret() {
		return errcode.TogRPCError(errcode.Unauthorized)
	}
	return nil
}

type TagServer struct {
	BlogURL string
	auth *Auth
}

func NewTagServer(blogURL string) *TagServer {
	return &TagServer{BlogURL: blogURL}
}

func (t *TagServer) GetTagList(c context.Context, r *pb.GetTagListRequest) (*pb.GetTagListReply, error) {
	//panic("测试抛出异常!")
	if err := t.auth.Check(c); err != nil {
		return nil, err
	}
	time.Sleep(15 * time.Second)
	api := bapi.NewAPI(t.BlogURL)
	body, err := api.GetTagList(c, r.GetName())
	if err != nil {
		return nil, errcode.TogRPCError(errcode.ErrorGetTagListFail)
	}

	tagList := pb.GetTagListReply{}
	err = json.Unmarshal(body, &tagList)
	if err != nil {
		return nil, errcode.TogRPCError(errcode.Fail)
	}

	return &tagList, nil
}
