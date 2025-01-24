package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "recommendation-system/pkg/logger"

	_ "recommendation-system/cmd/user-service/docs"
	"recommendation-system/internal/user/delivery/http"
	"recommendation-system/internal/user/repository"
	"recommendation-system/internal/user/service"
	"recommendation-system/pkg/db"
	"recommendation-system/pkg/kafka"
	"recommendation-system/pkg/redis"

	"github.com/spf13/viper"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// @title           User Service API
// @version         1.0
// @description     API for managing users in the recommendation system.
// @host            localhost:8080
// @BasePath        /api

func main() {
	logFile := "logger/logger.log"
	logger, err := log.NewLogger(logFile, "user-service", "user-app", "development")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}

	logger.Println("Starting User Service...")

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

	if err := kafkaClient.CreateTopic("user_updates", 1, 1); err != nil {
		logger.Printf("Topic 'user_updates' may already exist or failed to create: %v", err)
	}

	redisHosts := viper.GetStringSlice("redis.host")
	if len(redisHosts) == 0 {
		logger.Fatalf("No Redis hosts specified in the configuration")
	}
	redisClient := redis.NewRedisClient(redisHosts[0], "", 0)
	logger.Println("Redis client initialized")

	userRepo := repository.NewUserRepository(database, logger)
	userService := service.NewUserService(userRepo, kafkaClient, redisClient, logger)
	userHandler := http.NewHandler(userService, logger)

	jwtSecret := viper.GetString("jwt.secret")
	if jwtSecret == "" {
		logger.Fatal("JWT secret is not set")
	}
	logger.Println("JWT secret loaded")

	app := http.NewFiberApp(userHandler, jwtSecret)

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	go func() {
		if err := app.Listen(viper.GetString("server.user_service_address")); err != nil {
			logger.Fatalf("Failed to start server: %v", err)
		}
		logger.Println("User Service started")
	}()
	logger.Printf("User Service is running on %s", viper.GetString("server.user_service_address"))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := app.Shutdown(); err != nil {
		logger.Fatalf("Failed to shutdown server: %v", err)
	}
	logger.Println("User Service stopped gracefully")
}

func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs/")

	viper.SetDefault("server.user_service_address", ":8080")
	viper.SetDefault("db.sslmode", "disable")
	viper.SetDefault("kafka.brokers", []string{"kafka:9092"})
	viper.SetDefault("jwt.secret", "your_secret_key")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}
