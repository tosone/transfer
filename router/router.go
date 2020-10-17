package router

import "github.com/gofiber/fiber/v2"

// Initialize ..
func Initialize(app *fiber.App) (err error) {
	if err = Task(app); err != nil {
		return
	}
	if err = Database(app); err != nil {
		return
	}
	return
}
