package main

import (
	"metro_calcu/handler"
	"metro_calcu/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("search", middleware.DeserializationReq(handler.Search))
	router.GET("search_html", middleware.DeserializationReq(handler.SearchHtml))

	err := router.Run("0.0.0.0:8081")
	if err != nil {
		panic(err)
	}
}
