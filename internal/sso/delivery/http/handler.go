package http

import (
	"recommendation-system/internal/sso/models"
	"recommendation-system/internal/sso/service"
	log "recommendation-system/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// @title           SSO Service API
// @version         1.0
// @description     API for Single Sign-On (SSO) functionalities.
// @host            localhost:8084
// @BasePath        /api
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

func NewFiberApp(h *Handler) *fiber.App {
	app := fiber.New()

	api := app.Group("/api")
	auth := api.Group("/auth")

	auth.Post("/register", h.RegisterUser)
	auth.Post("/login", h.LoginUser)

	return app
}

// RegisterUser godoc
// @Summary      Register a new user
// @Description  Create a new user with name, email, and password.
// @Tags         auth :8084
// @Accept       json
// @Produce      json
// @Param        user  body      models.RegisterRequest  true  "User to register"
// @Success      201   {object}  models.User
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /auth/register [post]
func (h *Handler) RegisterUser(c *fiber.Ctx) error {
	h.logger.Println("Handling RegisterUser request")
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Printf("Invalid request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		h.logger.Println("Missing required fields: Name, Email, or Password")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name, Email and Password are required"})
	}

	h.logger.Printf("Registering user: %s", req.Email)
	user, err := h.service.RegisterUser(c.Context(), &req)
	if err != nil {
		h.logger.Printf("Failed to register user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.logger.Printf("User registered successfully: %s", req.Email)
	return c.Status(fiber.StatusCreated).JSON(user)
}

// LoginUser godoc
// @Summary      User login
// @Description  Authenticate user and return JWT token.
// @Tags         auth :8084
// @Accept       json
// @Produce      json
// @Param        credentials  body      models.LoginRequest  true  "User credentials"
// @Success      200          {object}  models.LoginResponse
// @Failure      400          {object}  map[string]interface{}
// @Failure      401          {object}  map[string]interface{}
// @Failure      500          {object}  map[string]interface{}
// @Router       /auth/login [post]
func (h *Handler) LoginUser(c *fiber.Ctx) error {
	h.logger.Println("Handling LoginUser request")
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Printf("Invalid request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Email == "" || req.Password == "" {
		h.logger.Println("Missing required fields: Email or Password")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email and Password are required"})
	}

	h.logger.Printf("Authenticating user: %s", req.Email)
	token, err := h.service.LoginUser(c.Context(), &req)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			h.logger.Printf("Invalid credentials for user: %s", req.Email)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
		}
		h.logger.Printf("Failed to login user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.logger.Printf("User logged in successfully: %s", req.Email)
	return c.JSON(models.LoginResponse{Token: token})
}
