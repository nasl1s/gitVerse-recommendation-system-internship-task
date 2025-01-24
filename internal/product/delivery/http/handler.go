package http

import (
	"strconv"

	"recommendation-system/internal/product/models"
	"recommendation-system/internal/product/service"
	"recommendation-system/pkg/auth"
	log "recommendation-system/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

// @title           Product Service API
// @version         1.0
// @description     API for managing products in the recommendation system.
// @host            localhost:8081
// @BasePath        /api/products
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

type Handler struct {
	service service.ProductService
	logger  *log.Logger
}

func NewHandler(s service.ProductService, logger *log.Logger) *Handler {
	return &Handler{service: s, logger: logger}
}

func NewFiberApp(h *Handler, jwtSecret string) *fiber.App {
	app := fiber.New()

	api := app.Group("/api")

	api.Use(auth.JWTMiddleware(auth.JWTConfig{
		Secret: jwtSecret,
	}))

	products := api.Group("/products")

	products.Post("/", h.CreateProduct)
	products.Get("/:id", h.GetProduct)
	products.Put("/:id", h.UpdateProduct)
	products.Get("/", h.GetAllProducts)
	products.Delete("/:id", h.DeleteProduct)

	return app
}

// CreateProduct godoc
// @Summary      Create a new product
// @Description  Create a new product with name, description, price, and category.
// @Tags         products :8081
// @Accept       json
// @Produce      json
// @Param        product  body      models.Product  true  "Product to create"
// @Param Authorization header string true "Bearer {token}"
// @Security     BearerAuth
// @Success      201      {object}  models.Product
// @Failure      400      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /products [post]
func (h *Handler) CreateProduct(c *fiber.Ctx) error {
	h.logger.Println("Handling CreateProduct request")
	var product models.Product
	if err := c.BodyParser(&product); err != nil {
		h.logger.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{"error": "Invalid request body"})
	}

	if product.Name == "" || product.Description == "" || product.Price <= 0 {
		h.logger.Println("Invalid product details")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Name, Description and Price are required and must be valid",
		})
	}

	h.logger.Printf("Creating product: %s", product.Name)
	if err := h.service.CreateProduct(c.Context(), &product); err != nil {
		h.logger.Printf("Failed to create product: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{"error": err.Error()})
	}

	h.logger.Printf("Successfully created product with ID: %d", product.ID)
	return c.Status(fiber.StatusCreated).JSON(product)
}

// GetProduct godoc
// @Summary      Get product by ID
// @Description  Retrieve product information by product ID, including likes, dislikes, and purchase count.
// @Tags         products :8081
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Param Authorization header string true "Bearer {token}"
// @Security     BearerAuth
// @Success      200  {object}  models.Product
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /products/{id} [get]
func (h *Handler) GetProduct(c *fiber.Ctx) error {
	h.logger.Println("Handling GetProduct request")
	id, err := c.ParamsInt("id")
	if err != nil {
		h.logger.Printf("Invalid product ID: %s", c.Params("id"))
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{"error": "Invalid product ID"})
	}

	h.logger.Printf("Fetching product with ID: %d", id)
	product, err := h.service.GetProduct(c.Context(), int64(id))
	if err != nil {
		h.logger.Printf("Failed to fetch product with ID %d: %v", id, err)
		return c.Status(fiber.StatusNotFound).JSON(map[string]interface{}{"error": "Product not found"})
	}

	h.logger.Printf("Successfully fetched product with ID: %d", id)
	return c.JSON(product)
}

// UpdateProduct godoc
// @Summary      Update product information
// @Description  Update product information by product ID (including category).
// @Tags         products :8081
// @Accept       json
// @Produce      json
// @Param        id      path      int           true  "Product ID"
// @Param        product body      models.Product  true  "Updated product information"
// @Param Authorization header string true "Bearer {token}"
// @Security     BearerAuth
// @Success      200     {object}  models.Product
// @Failure      400     {object}  map[string]interface{}
// @Failure      500     {object}  map[string]interface{}
// @Router       /products/{id} [put]
func (h *Handler) UpdateProduct(c *fiber.Ctx) error {
	h.logger.Println("Handling UpdateProduct request")
	id, err := c.ParamsInt("id")
	if err != nil {
		h.logger.Printf("Invalid product ID: %s", c.Params("id"))
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{"error": "Invalid product ID"})
	}

	var product models.Product
	if err := c.BodyParser(&product); err != nil {
		h.logger.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{"error": "Invalid request body"})
	}

	if product.Name == "" || product.Description == "" || product.Price <= 0 {
		h.logger.Println("Invalid product details")
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Name, Description and Price are required and must be valid",
		})
	}

	product.ID = int64(id)

	h.logger.Printf("Updating product with ID: %d", id)
	if err := h.service.UpdateProduct(c.Context(), &product); err != nil {
		h.logger.Printf("Failed to update product with ID %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]interface{}{"error": err.Error()})
	}

	h.logger.Printf("Successfully updated product with ID: %d", id)
	return c.JSON(product)
}

// GetAllProducts godoc
// @Summary      Get all products
// @Description  Retrieve a list of all products with pagination, including likes, dislikes, and purchase count.
// @Tags         products :8081
// @Accept       json
// @Produce      json
// @Param        page     query     int  false  "Page number"
// @Param        pageSize query    int  false  "Number of products per page"
// @Param Authorization header string true "Bearer {token}"
// @Security     BearerAuth
// @Success      200      {array}   models.Product
// @Failure      400      {object}  map[string]interface{}
// @Failure      500      {object}  map[string]interface{}
// @Router       /products [get]
func (h *Handler) GetAllProducts(c *fiber.Ctx) error {
	h.logger.Println("Handling GetAllProducts request")
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		h.logger.Println("Invalid page number")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid page number"})
	}

	pageSize, err := strconv.Atoi(c.Query("pageSize", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		h.logger.Println("Invalid page size")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid page size"})
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	h.logger.Printf("Fetching products with limit: %d, offset: %d", limit, offset)
	products, err := h.service.GetAllProducts(c.Context(), limit, offset)
	if err != nil {
		h.logger.Printf("Failed to fetch products: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.logger.Printf("Successfully fetched %d products", len(products))
	return c.JSON(products)
}

// DeleteProduct godoc
// @Summary      Delete a product
// @Description  Delete a product by its ID.
// @Tags         products :8081
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Param Authorization header string true "Bearer {token}"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /products/{id} [delete]
func (h *Handler) DeleteProduct(c *fiber.Ctx) error {
	h.logger.Println("Handling DeleteProduct request")
	id, err := c.ParamsInt("id")
	if err != nil {
		h.logger.Printf("Invalid product ID: %s", c.Params("id"))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product ID"})
	}

	h.logger.Printf("Deleting product with ID: %d", id)
	if err := h.service.DeleteProduct(c.Context(), int64(id)); err != nil {
		h.logger.Printf("Failed to delete product with ID %d: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.logger.Printf("Successfully deleted product with ID: %d", id)
	return c.JSON(fiber.Map{"message": "Product deleted successfully"})
}
