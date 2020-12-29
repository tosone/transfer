package router

import (
	"io"
	"mime/multipart"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"

	"transfer/database"
)

// Database ..
func Database(app *gin.Engine) (err error) {
	var databaseLocker = &sync.Mutex{}
	const backupDir = "./data"

	if !com.IsDir(backupDir) {
		if err = os.MkdirAll(backupDir, 0755); err != nil {
			return
		}
	}

	app.GET("/database", func(ctx *gin.Context) {
		databaseLocker.Lock()
		defer databaseLocker.Unlock()

		ctx.Header("Content-Disposition", `attachment; filename="backup.db"`)

		_ = ctx.Stream(func(w io.Writer) bool {
			if _, err = database.Backup(w, 0); err != nil {
				_ = ctx.Error(errServerInternal.Build(err))
			}
			return false
		})
	})
	app.POST("/database", func(ctx *gin.Context) {
		databaseLocker.Lock()
		defer databaseLocker.Unlock()

		var fileHeader *multipart.FileHeader
		if fileHeader, err = ctx.FormFile("file"); err != nil {
			_ = ctx.Error(errBadRequest.Build(err))
			return
		}
		var file multipart.File
		if file, err = fileHeader.Open(); err != nil {
			_ = ctx.Error(errServerInternal.Build(err))
			return
		}

		if err = database.Load(file, 4); err != nil {
			_ = ctx.Error(errServerInternal.Build(err))
			return
		}
		if err = file.Close(); err != nil {
			_ = ctx.Error(errServerInternal.Build(err))
			return
		}
	})
	return
}
