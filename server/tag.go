package server

import (
	"context"
	"encoding/json"
	"tag-service/pkg/bapi"
	"tag-service/pkg/errcode"
	pb "tag-service/proto"
)

type TagServer struct {
	BlogURL string
}

func NewTagServer(blogURL string) *TagServer {
	return &TagServer{BlogURL: blogURL}
}

func (t *TagServer) GetTagList(c context.Context, r *pb.GetTagListRequest) (*pb.GetTagListReply, error) {
	panic("测试抛出异常!")
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
