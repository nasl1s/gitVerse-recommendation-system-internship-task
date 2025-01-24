package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "recommendation-system/pkg/logger"

	_ "recommendation-system/cmd/product-service/docs"

	"github.com/spf13/viper"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	"recommendation-system/internal/product/delivery/http"
	"recommendation-system/internal/product/repository"
	"recommendation-system/internal/product/service"
	"recommendation-system/pkg/db"
	"recommendation-system/pkg/kafka"
)

// @title           Product Service API
// @version         1.0
// @description     API for managing products in the recommendation system.
// @host            localhost:8081
// @BasePath        /api

func main() {
	logFile := "logger/logger.log"
	logger, err := log.NewLogger(logFile, "product-service", "product-app", "development")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	if err := initConfig(); err != nil {
		logger.Fatalf("Error loading config: %v", err)
	}

	dbConfig := db.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetInt("db.port"),
		User:     viper.GetString("db.user"),
		Password: viper.GetString("db.password"),
		DBName:   viper.GetString("db.name"),
		SSLMode:  viper.GetString("db.sslmode"),
	}

	database, err := db.New(dbConfig)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	logger.Println("Connected to the database")

	dsn := dbConfig.GetDSN()
	if err := db.RunMigrations(dsn); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}
	logger.Println("Database migrations applied successfully")

	kafkaBrokers := viper.GetStringSlice("kafka.brokers")
	kafkaClient := kafka.NewKafkaClient(kafkaBrokers)
	defer kafkaClient.Close()
	logger.Println("Kafka client initialized")

	if err := kafkaClient.CreateTopic("product_updates", 1, 1); err != nil {
		logger.Printf("Topic 'product_updates' may already exist or failed to create: %v", err)
	}

	productRepo := repository.NewProductRepository(database, logger)
	productService := service.NewProductService(productRepo, kafkaClient, logger)
	productHandler := http.NewHandler(productService, logger)

	jwtSecret := viper.GetString("jwt.secret")
	if jwtSecret == "" {
		logger.Fatal("JWT secret is not set")
	}
	logger.Println("JWT secret loaded")

	app := http.NewFiberApp(productHandler, jwtSecret)
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	go func() {
		if err := app.Listen(viper.GetString("server.product_service_address")); err != nil {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()
	logger.Printf("Product Service is running on %s", viper.GetString("server.product_service_address"))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := app.Shutdown(); err != nil {
		logger.Fatalf("Failed to shutdown server: %v", err)
	}
	logger.Println("Product Service stopped gracefully")
}

func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs/")

	viper.SetDefault("server.product_service_address", ":8081")
	viper.SetDefault("db.sslmode", "disable")
	viper.SetDefault("kafka.brokers", []string{"kafka:9092"})
	viper.SetDefault("jwt.secret", "your_secret_key")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}
