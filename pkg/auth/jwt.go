package auth

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrMissingToken = errors.New("missing or malformed JWT")
	ErrInvalidToken = errors.New("invalid or expired JWT")
	ErrUnauthorized = errors.New("unauthorized")
)

type JWTConfig struct {
	Secret string
}

func JWTMiddleware(config JWTConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": ErrMissingToken.Error()})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": ErrMissingToken.Error()})
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(config.Secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": ErrInvalidToken.Error()})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": ErrInvalidToken.Error()})
		}

		if userID, ok := claims["sub"].(string); ok {
			c.Locals("userID", userID)
		}

		return c.Next()
	}
}
