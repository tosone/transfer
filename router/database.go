package router

import (
	"bufio"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/unknwon/com"

	"transfer/database"
)

// Database ..
func Database(app *fiber.App) (err error) {
	var databaseLocker = &sync.Mutex{}
	const backupFile = "backupFile.db"
	const backupDir = "./data"

	if !com.IsDir(backupDir) {
		if err = os.MkdirAll(backupDir, 0755); err != nil {
			return
		}
	}

	app.Get("/database", func(ctx *fiber.Ctx) (err error) {
		databaseLocker.Lock()
		defer databaseLocker.Unlock()

		var file *os.File
		if file, err = os.Create(filepath.Join(backupDir, backupFile)); err != nil {
			return
		}
		var bw = bufio.NewWriterSize(file, 64<<20)
		if _, err = database.Backup(bw, 0); err != nil {
			return
		}
		if err = bw.Flush(); err != nil {
			return
		}
		if err = file.Sync(); err != nil {
			return err
		}
		if err = file.Close(); err != nil {
			return
		}
		if err = ctx.SendFile(filepath.Join(backupDir, backupFile), true); err != nil {
			return
		}
		return
	})
	app.Post("/database", func(ctx *fiber.Ctx) (err error) {
		databaseLocker.Lock()
		defer databaseLocker.Unlock()

		var multipartForm *multipart.Form
		if multipartForm, err = ctx.MultipartForm(); err != nil {
			return
		}
		var files = multipartForm.File["file"]
		var filename string
		for _, file := range files {
			filename = filepath.Join(backupDir, file.Filename)
			if err := ctx.SaveFile(file, filename); err != nil {
				return err
			}
		}
		var file *os.File
		if file, err = os.Open(filename); err != nil {
			return
		}
		if err = database.Load(file, 4); err != nil {
			return
		}
		if err = file.Close(); err != nil {
			return
		}
		return ctx.SendString("ok")
	})
	return
}
