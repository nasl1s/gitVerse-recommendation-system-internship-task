package http

import (
	"recommendation-system/internal/recommendation/service"
	"recommendation-system/pkg/auth"
	log "recommendation-system/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// @title           Recommendation Service API
// @version         1.0
// @description     API for getting recommendations for users in the recommendation system.
// @host            localhost:8082
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

type Handler struct {
	service service.RecommendationService
	logger  *log.Logger
}

func NewHandler(s service.RecommendationService, logger *log.Logger) *Handler {
	return &Handler{
		service: s,
		logger:  logger,
	}
}

func NewFiberApp(h *Handler, jwtSecret string) *fiber.App {
	app := fiber.New()
	api := app.Group("/api")

	api.Use(auth.JWTMiddleware(auth.JWTConfig{
		Secret: jwtSecret,
	}))

	recommendations := api.Group("/recommendations")
	recommendations.Get("/:user_id/latest", h.GetLatestRecommendation)

	return app
}

// GetLatestRecommendation godoc
// @Summary      Get the latest recommendation
// @Description  Retrieve the most up-to-date recommended products for this user.
// @Tags         recommendations :8082
// @Accept       json
// @Produce      json
// @Param        user_id   path      int  true  "User ID"
// @Param Authorization header string true "Bearer {token}"
// @Security     BearerAuth
// @Success      200       {object}  map[string]interface{}
// @Failure      400       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]interface{}
// @Router       /recommendations/{user_id}/latest [get]
func (h *Handler) GetLatestRecommendation(c *fiber.Ctx) error {
	h.logger.Println("Processing request to get the latest recommendation")

	userID, err := c.ParamsInt("user_id")
	if err != nil || userID <= 0 {
		h.logger.Printf("Invalid user ID: %s", c.Params("user_id"))
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Invalid user ID",
		})
	}

	h.logger.Printf("Fetching latest recommendations for user ID: %d", userID)
	productIDs, err := h.service.GetLatestRecommendation(c.Context(), int64(userID))
	if err != nil {
		h.logger.Printf("Failed to retrieve recommendations for user ID %d: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{
			"error": err.Error(),
		})
	}

	h.logger.Printf("Successfully fetched recommendations for user ID: %d", userID)
	return c.JSON(fiber.Map{
		"recommended_product_ids": productIDs,
	})
}
