package router

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrBody struct {
	Code int
	Msg  string
}

func (e ErrBody) Error() string {
	return e.Msg
}

func (e ErrBody) Build(str ...interface{}) ErrBody {
	e.Msg = fmt.Sprintf(e.Msg, str...)
	return e
}

var (
	errServerInternal = ErrBody{http.StatusInternalServerError, "server internal error: %v"}
	errDatabase       = ErrBody{http.StatusInternalServerError, `database error: %v`}
	errBadRequest     = ErrBody{http.StatusBadRequest, `bad request: %v`}
	errURLConflict    = ErrBody{http.StatusConflict, `url conflict: %s, or you should set force true`}
)

func errorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		var errorMessages []string
		for _, err := range ctx.Errors {
			errorMessages = append(errorMessages, err.Error())
		}

		if len(ctx.Errors) > 0 {
			if ctx.Errors[0].IsType(gin.ErrorTypePrivate) {
				var err = ctx.Errors[0].Err.(ErrBody)
				ctx.JSON(err.Code, gin.H{"error": err.Msg})
			} else {
				ctx.JSON(http.StatusInternalServerError, ctx.Errors[0].JSON())
			}
		}
	}
}
