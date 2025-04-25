package utils

import "github.com/gofiber/fiber/v2"

// Response struct untuk standar response API
type Response struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// SuccessResponse mengembalikan response sukses
func SuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"status":  true,
		"message": message,
		"data":    data,
	})
}

// ResponseError mengembalikan response error
func ResponseError(c *fiber.Ctx, statusCode int, message string, details interface{}) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"status":  false,
		"message": message,
		"error":   details,
	})
}

// Pagination struct untuk menyimpan informasi pagination
type Pagination struct {
	CurrentPage int `json:"current_page"`
	TotalData   int `json:"total_data"`
	TotalPages  int `json:"total_pages"`
}
