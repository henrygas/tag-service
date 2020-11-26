package server

import (
	"context"
	"encoding/json"
	"tag-service/pkg/bapi"
	pb "tag-service/proto"
)

type TagServer struct {

}

func NewTagServer() *TagServer {
	return &TagServer{}
}

func (t *TagServer) GetTagList(c context.Context, r *pb.GetTagListRequest) (*pb.GetTagListReply, error) {
	api := bapi.NewAPI("http://127.0.0.1:8000")
	body, err := api.GetTagList(c, r.GetName())
	if err != nil {
		return nil, errcode.TogRPCError(errcode.ERROR_GET_TAG_LIST_FAIL)
	}

	tagList := pb.GetTagListReply{}
	err = json.Unmarshal(body, &tagList)
	if err != nil {
		return nil, errcode.TogRPCError(errcode.FAIL)
	}

	return &tagList, nil
}
