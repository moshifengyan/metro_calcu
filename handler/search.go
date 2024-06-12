package handler

import (
	"metro_calcu/manager"

	"github.com/gin-gonic/gin"
)

type SearchReq struct {
	Region    string `form:"region"`
	StartName string `form:"start_name"`
	Duration  int64  `form:"duration"`
}

type SearchResp struct {
	Result []string `json:"result"`
}

func Search(ctx *gin.Context, req *SearchReq) (SearchResp, error) {
	resp := SearchResp{}
	resp.Result = manager.CalculateDest(req.Region, req.StartName, req.Duration)
	return resp, nil
}
