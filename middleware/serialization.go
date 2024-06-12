package middleware

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

func DeserializationReq(handler interface{}) gin.HandlerFunc {
	hType := reflect.TypeOf(handler)
	hValue := reflect.ValueOf(handler)
	realFc := func(ctx *gin.Context) {}
	realBody := func(args []reflect.Value) (results []reflect.Value) {
		ctx, ok := args[0].Interface().(*gin.Context)
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return nil
		}
		var realIN []reflect.Value
		realIN = append(realIN, args[0])
		if hType.NumIn() == 2 {
			param := hType.In(1)
			if param.Kind() == reflect.Ptr {
				param = param.Elem()
			}
			val := reflect.New(param)
			if ctx.Request.Method == http.MethodGet {
				if err := ctx.ShouldBindQuery(val.Interface()); err != nil {
					ctx.AbortWithStatus(http.StatusInternalServerError)
					return nil
				}
			} else {
				if err := ctx.ShouldBind(val.Interface()); err != nil {
					ctx.AbortWithStatus(http.StatusInternalServerError)
					return nil
				}
			}
			for _i := 0; _i < val.Elem().NumField(); _i++ {
				if val.Elem().Field(_i).Kind() == reflect.String {
					val.Elem().Field(_i).SetString(strings.Trim(val.Elem().Field(_i).String(), " "))
				}
			}
			realIN = append(realIN, val)
		}
		vals := hValue.Call(realIN)
		// 最后一个返回参数是error
		valNum := hType.NumOut()
		if valNum != 0 {
			statusCode := http.StatusOK
			var content interface{}
			content = http.StatusText(http.StatusOK)
			if vals[valNum-1].Interface() != nil {
				_, ok := vals[valNum-1].Interface().(error)
				if !ok {
					ctx.Abort()
				}
			} else if hType.NumOut() != 1 {
				content = vals[0].Interface()
			}
			if !ctx.IsAborted() {
				ctx.JSON(statusCode, content)
			}
		}
		ctx.Next()
		return nil
	}
	h := reflect.MakeFunc(reflect.TypeOf(realFc), realBody)
	return h.Interface().(func(ctx *gin.Context))
}
