package http

import (
	"strconv"

	"recommendation-system/internal/analytics/service"
	"recommendation-system/pkg/auth"
	log "recommendation-system/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// @title           Analytics Service API
// @version         1.0
// @description     API for analytics in the recommendation system.
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

type Handler struct {
	service service.AnalyticsService
	logger  *log.Logger
}

func NewHandler(s service.AnalyticsService, logger *log.Logger) *Handler {
	return &Handler{service: s, logger: logger}
}

func NewFiberApp(s service.AnalyticsService, jwtSecret string, logger *log.Logger) *fiber.App {
	app := fiber.New()
	api := app.Group("/api")
	analytics := api.Group("/analytics")

	api.Use(auth.JWTMiddleware(auth.JWTConfig{
		Secret: jwtSecret,
	}))

	handler := NewHandler(s, logger)
	analytics.Get("/products/:id", handler.getProductAnalytics())
	analytics.Get("/users/:id", handler.getUserAnalytics())

	return app
}

// getProductAnalytics godoc
// @Summary      Get product analytics
// @Description  Retrieve analytics data for a specific product (likes, dislikes, purchases).
// @Tags         analytics :8083
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Param Authorization header string true "Bearer {token}"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /analytics/products/{id} [get]
func (h *Handler) getProductAnalytics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		h.logger.Println("Processing request to get product analytics")
		pid, err := strconv.Atoi(c.Params("id"))
		if err != nil || pid <= 0 {
			h.logger.Printf("Invalid product ID: %s", c.Params("id"))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product ID"})
		}

		pa, err := h.service.GetProductAnalytics(c.Context(), int64(pid))
		if err != nil {
			h.logger.Printf("Failed to retrieve product analytics for ID %d: %v", pid, err)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}

		h.logger.Printf("Successfully retrieved analytics for product ID: %d", pid)
		return c.JSON(pa)
	}
}

// getUserAnalytics godoc
// @Summary      Get user analytics
// @Description  Retrieve analytics data for a specific user (total likes, dislikes, purchases).
// @Tags         analytics :8083
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Param Authorization header string true "Bearer {token}"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /analytics/users/{id} [get]
func (h *Handler) getUserAnalytics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		h.logger.Println("Processing request to get user analytics")
		uid, err := strconv.Atoi(c.Params("id"))
		if err != nil || uid <= 0 {
			h.logger.Printf("Invalid user ID: %s", c.Params("id"))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
		}

		ua, err := h.service.GetUserAnalytics(c.Context(), int64(uid))
		if err != nil {
			h.logger.Printf("Failed to retrieve user analytics for ID %d: %v", uid, err)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}

		h.logger.Printf("Successfully retrieved analytics for user ID: %d", uid)
		return c.JSON(ua)
	}
}
