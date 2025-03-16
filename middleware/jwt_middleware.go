package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/habbazettt/mahad-service-go/utils"
	"github.com/sirupsen/logrus"
)

// JWTMiddleware memeriksa token JWT di header Authorization
func JWTMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		logrus.Warn("Unauthorized access attempt: Missing Authorization header")
		return utils.ResponseError(c, fiber.StatusUnauthorized, "Missing Authorization header", nil)
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		logrus.Warn("Unauthorized access attempt: Invalid Authorization format")
		return utils.ResponseError(c, fiber.StatusUnauthorized, "Invalid Authorization format", nil)
	}

	claims, err := utils.VerifyToken(parts[1])
	if err != nil {
		logrus.WithError(err).Warn("Unauthorized access attempt: Invalid token")
		return utils.ResponseError(c, fiber.StatusUnauthorized, "Invalid token", err.Error())
	}

	logrus.WithFields(logrus.Fields{
		"user_id": claims.ID,
		"role":    claims.Role,
	}).Info("Token verified successfully")

	c.Locals("user", claims)
	return c.Next()
}

// RoleMiddleware memeriksa apakah pengguna memiliki peran yang diperlukan
func RoleMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user").(*utils.Claims)

		for _, role := range allowedRoles {
			if user.Role == role {
				logrus.WithFields(logrus.Fields{
					"user_id": user.ID,
					"role":    user.Role,
				}).Info("User authorized to access this route")
				return c.Next()
			}
		}

		logrus.WithFields(logrus.Fields{
			"user_id": user.ID,
			"role":    user.Role,
		}).Warn("Unauthorized access attempt")

		return utils.ResponseError(c, fiber.StatusForbidden, "You are not authorized to access this resource", nil)
	}
}
