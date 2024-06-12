package handler

import (
	"metro_calcu/manager"
	"net/http"
	"strings"

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

type SearchHtmlRes struct {
	Name string
	Time string
}

func Search(ctx *gin.Context, req *SearchReq) (SearchResp, error) {
	resp := SearchResp{}
	resp.Result = manager.CalculateDest(req.Region, req.StartName, req.Duration)
	return resp, nil
}

func SearchHtml(ctx *gin.Context, req *SearchReq) {
	res, _ := Search(ctx, req)
	stations := make([]SearchHtmlRes, 0)
	for _, entry := range res.Result {
		parts := strings.Split(strings.TrimSpace(entry), " ")
		stations = append(stations, SearchHtmlRes{Name: parts[0], Time: parts[1]})
	}
	ctx.HTML(http.StatusOK, "search.html", gin.H{
		"stations": stations,
	})
}
