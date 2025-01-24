package http

import (
	"strconv"

	"recommendation-system/internal/user/models"
	"recommendation-system/internal/user/service"
	"recommendation-system/pkg/auth"
	log "recommendation-system/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// @title           User Service API
// @version         1.0
// @description     API for managing users in the recommendation system.
// @host            localhost:8080
// @BasePath        /api/users
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

type Handler struct {
	service service.UserService
	logger  *log.Logger
}

func NewHandler(s service.UserService, logger *log.Logger) *Handler {
	return &Handler{service: s, logger: logger}
}

func NewFiberApp(h *Handler, jwtSecret string) *fiber.App {
	app := fiber.New()

	api := app.Group("/api")

	api.Use(auth.JWTMiddleware(auth.JWTConfig{
		Secret: jwtSecret,
	}))

	users := api.Group("/users")

	users.Get("/:id", h.GetUser)
	users.Put("/:id", h.UpdateUser)
	users.Get("/", h.GetAllUsers)

	users.Post("/:id/like", h.LikeProduct)
	users.Post("/:id/dislike", h.DislikeProduct)

	users.Get("/:id/actions", h.GetUserActions)
	users.Get("/:id/purchases", h.GetUserPurchases)
	users.Post("/:id/purchase", h.PurchaseProduct)

	return app
}

// GetUser godoc
// @Summary      Get user by ID
// @Description  Retrieve user information by user ID.
// @Tags         users :8080
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Param Authorization header string true "Bearer {token}"
// @Success      200  {object}  models.User
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /users/{id} [get]
func (h *Handler) GetUser(c *fiber.Ctx) error {
	h.logger.Println("Handling GetUser request")
	id, err := c.ParamsInt("id")
	if err != nil {
		h.logger.Printf("Invalid user ID: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	h.logger.Printf("Fetching user with ID: %d", id)
	user, err := h.service.GetUser(c.Context(), int64(id))
	if err != nil {
		h.logger.Printf("User not found: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	h.logger.Printf("Successfully fetched user with ID: %d", id)
	return c.JSON(user)
}

// UpdateUser godoc
// @Summary      Update user profile
// @Description  Update user information by user ID.
// @Tags         users :8080
// @Accept       json
// @Produce      json
// @Param        id    path      int          true  "User ID"
// @Param        user  body      models.User  true  "Updated user information"
// @Param Authorization header string true "Bearer {token}"
// @Security     BearerAuth
// @Success      200   {object}  models.User
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /users/{id} [put]
func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	h.logger.Println("Handling UpdateUser request")
	id, err := c.ParamsInt("id")
	if err != nil {
		h.logger.Printf("Invalid user ID: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	h.logger.Printf("Updating user with ID: %d", id)
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		h.logger.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	user.ID = int64(id)
	if err := h.service.UpdateUser(c.Context(), &user); err != nil {
		h.logger.Printf("Failed to update user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.logger.Printf("Successfully updated user with ID: %d", id)
	return c.JSON(user)
}

// GetAllUsers godoc
// @Summary      Get all users
// @Description  Retrieve a list of all users with pagination.
// @Tags         users :8080
// @Accept       json
// @Produce      json
// @Param        page     query     int  false  "Page number"
// @Param        pageSize query    int  false  "Number of users per page"
// @Param Authorization header string true "Bearer {token}"
// @Security     BearerAuth
// @Success      200      {array}   models.User
// @Failure      400      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /users [get]
func (h *Handler) GetAllUsers(c *fiber.Ctx) error {
	h.logger.Println("Handling GetAllUsers request")
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		h.logger.Printf("Invalid page number: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid page number"})
	}

	pageSize, err := strconv.Atoi(c.Query("pageSize", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		h.logger.Printf("Invalid page size: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid page size"})
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	h.logger.Printf("Fetching users with limit %d and offset %d", limit, offset)
	users, err := h.service.GetAllUsers(c.Context(), limit, offset)
	if err != nil {
		h.logger.Printf("Failed to fetch users: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.logger.Printf("Successfully fetched %d users", len(users))
	return c.JSON(users)
}

// PurchaseProduct godoc
// @Summary      Purchase a product
// @Description  User purchases a product.
// @Tags         users :8080
// @Accept       json
// @Produce      json
// @Param        id          path      int  true  "User ID"
// @Param        product_id  query     int  true  "Product ID"
// @Param Authorization header string true "Bearer {token}"
// @Security     BearerAuth
// @Success      200         {object}  map[string]interface{}
// @Failure      400         {object}  map[string]interface{}
// @Failure      500         {object}  map[string]interface{}
// @Router       /users/{id}/purchase [post]
func (h *Handler) PurchaseProduct(c *fiber.Ctx) error {
	h.logger.Println("Handling PurchaseProduct request")
	userID, err := c.ParamsInt("id")
	if err != nil {
		h.logger.Printf("Invalid user ID: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	productIDStr := c.Query("product_id")
	if productIDStr == "" {
		h.logger.Println("Product ID is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Product ID is required"})
	}

	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil || productID <= 0 {
		h.logger.Printf("Invalid product ID: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product ID"})
	}

	h.logger.Printf("User ID %d is purchasing product ID %d", userID, productID)
	if err := h.service.PurchaseProduct(c.Context(), int64(userID), productID); err != nil {
		h.logger.Printf("Failed to purchase product: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.logger.Printf("Successfully processed purchase for user ID %d and product ID %d", userID, productID)
	return c.JSON(fiber.Map{"message": "Product purchased successfully"})
}

// LikeProduct godoc
// @Summary      Like a product
// @Description  User likes a product.
// @Tags         users :8080
// @Accept       json
// @Produce      json
// @Param        id          path      int  true  "User ID"
// @Param        product_id  query     int  true  "Product ID"
// @Param Authorization header string true "Bearer {token}"
// @Security     BearerAuth
// @Success      200         {object}  map[string]interface{}
// @Failure      400         {object}  map[string]interface{}
// @Failure      500         {object}  map[string]interface{}
// @Router       /users/{id}/like [post]
func (h *Handler) LikeProduct(c *fiber.Ctx) error {
	h.logger.Println("Handling LikeProduct request")
	userID, err := c.ParamsInt("id")
	if err != nil {
		h.logger.Printf("Invalid user ID: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	productIDStr := c.Query("product_id")
	if productIDStr == "" {
		h.logger.Println("Product ID is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Product ID is required"})
	}

	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil || productID <= 0 {
		h.logger.Printf("Invalid product ID: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product ID"})
	}

	h.logger.Printf("User ID %d is liking product ID %d", userID, productID)
	if err := h.service.LikeProduct(c.Context(), int64(userID), productID); err != nil {
		h.logger.Printf("Failed to like product: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.logger.Printf("Successfully liked product ID %d for user ID %d", productID, userID)
	return c.JSON(fiber.Map{"message": "Product liked successfully"})
}

// DislikeProduct godoc
// @Summary      Dislike a product
// @Description  User dislikes a product.
// @Tags         users :8080
// @Accept       json
// @Produce      json
// @Param        id          path      int  true  "User ID"
// @Param        product_id  query     int  true  "Product ID"
// @Param Authorization header string true "Bearer {token}"
// @Security     BearerAuth
// @Success      200         {object}  map[string]interface{}
// @Failure      400         {object}  map[string]interface{}
// @Failure      500         {object}  map[string]interface{}
// @Router       /users/{id}/dislike [post]
func (h *Handler) DislikeProduct(c *fiber.Ctx) error {
	h.logger.Println("Handling DislikeProduct request")
	userID, err := c.ParamsInt("id")
	if err != nil {
		h.logger.Printf("Invalid user ID: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	productIDStr := c.Query("product_id")
	if productIDStr == "" {
		h.logger.Println("Product ID is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Product ID is required"})
	}

	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil || productID <= 0 {
		h.logger.Printf("Invalid product ID: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product ID"})
	}

	h.logger.Printf("User ID %d is disliking product ID %d", userID, productID)
	if err := h.service.DislikeProduct(c.Context(), int64(userID), productID); err != nil {
		h.logger.Printf("Failed to dislike product: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.logger.Printf("Successfully disliked product ID %d for user ID %d", productID, userID)
	return c.JSON(fiber.Map{"message": "Product disliked successfully"})
}

// GetUserActions godoc
// @Summary      Get user actions
// @Description  Retrieve all likes and dislikes of a user, with optional filtering by product ID.
// @Tags         users :8080
// @Accept       json
// @Produce      json
// @Param        id         path      int  true  "User ID"
// @Param        product_id query     int  false "Product ID to filter"
// @Param Authorization header string true "Bearer {token}"
// @Security     BearerAuth
// @Success      200        {object}  map[string]interface{}
// @Failure      400        {object}  map[string]interface{}
// @Failure      500        {object}  map[string]interface{}
// @Router       /users/{id}/actions [get]
func (h *Handler) GetUserActions(c *fiber.Ctx) error {
	h.logger.Println("Handling GetUserActions request")
	userID, err := c.ParamsInt("id")
	if err != nil {
		h.logger.Printf("Invalid user ID: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	productIDStr := c.Query("product_id")
	var productID *int64
	if productIDStr != "" {
		pid, err := strconv.ParseInt(productIDStr, 10, 64)
		if err != nil || pid <= 0 {
			h.logger.Printf("Invalid product ID: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product ID"})
		}
		productID = &pid
	}

	h.logger.Printf("Fetching actions for user ID: %d", userID)
	actions, err := h.service.GetUserActions(c.Context(), int64(userID), productID)
	if err != nil {
		h.logger.Printf("Failed to fetch user actions: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.logger.Printf("Successfully fetched actions for user ID: %d", userID)
	return c.JSON(actions)
}

// GetUserPurchases godoc
// @Summary      Get user purchases
// @Description  Retrieve a list of purchases made by the user with pagination.
// @Tags         users :8080
// @Accept       json
// @Produce      json
// @Param        id        path      int  true  "User ID"
// @Param        page      query     int  false "Page number"
// @Param        pageSize  query     int  false "Number of purchases per page"
// @Param Authorization header string true "Bearer {token}"
// @Security     BearerAuth
// @Success      200       {array}   models.Purchase
// @Failure      400       {object}  map[string]interface{}
// @Failure      500       {object}  map[string]interface{}
// @Router       /users/{id}/purchases [get]
func (h *Handler) GetUserPurchases(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid page number"})
	}

	pageSize, err := strconv.Atoi(c.Query("pageSize", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid page size"})
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	purchases, err := h.service.GetUserPurchases(c.Context(), int64(userID), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(purchases)
}
