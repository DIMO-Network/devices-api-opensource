package controllers

import "github.com/gofiber/fiber/v2"

func errorResponseHandler(c *fiber.Ctx, err error, status int) error {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return c.Status(status).JSON(fiber.Map{
		"error_message": msg,
	})
}
