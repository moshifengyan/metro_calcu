package main

import (
	"metro_calcu/handler"
	"metro_calcu/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("search", middleware.DeserializationReq(handler.Search))

	err := router.Run("localhost:8081")
	if err != nil {
		panic(err)
	}
}
