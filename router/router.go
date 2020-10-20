package router

import (
	"github.com/gin-gonic/gin"
)

// Initialize ..
func Initialize(app *gin.Engine) (err error) {
	app.Use(errorHandler())
	if err = Task(app); err != nil {
		return
	}
	if err = Database(app); err != nil {
		return
	}
	return
}
