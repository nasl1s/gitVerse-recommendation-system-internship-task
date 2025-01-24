package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "recommendation-system/pkg/logger"

	_ "recommendation-system/cmd/analytics-service/docs"

	"github.com/spf13/viper"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	"recommendation-system/internal/analytics/delivery/http"
	"recommendation-system/internal/analytics/repository"
	"recommendation-system/internal/analytics/service"
	"recommendation-system/pkg/db"
	"recommendation-system/pkg/kafka"

	kafkaGo "github.com/segmentio/kafka-go"
)

// @title           Analytics Service API
// @version         1.0
// @description     API for collecting and retrieving analytics data.
// @host            localhost:8083
// @BasePath        /api

func main() {
	logFile := "logger/logger.log"
	logger, err := log.NewLogger(logFile, "analytics-service", "analytics-app", "development")
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

	analyticsRepo := repository.NewAnalyticsRepository(database, logger)
	analyticsService := service.NewAnalyticsService(analyticsRepo, kafkaClient, logger)

	go func() {
		ctx := context.Background()
		topics := []string{"user_updates", "product_updates"}
		groupID := "analytics_service_group"
		if err := kafkaClient.SubscribeToTopicsFallback(ctx, topics, groupID, func(m kafkaGo.Message) error {
			return analyticsService.ProcessKafkaMessage(ctx, m)
		}); err != nil {
			logger.Printf("Error subscribing to Kafka topics: %v", err)
		}
	}()
	logger.Println("Kafka topics subscription started")

	jwtSecret := viper.GetString("jwt.secret")
	if jwtSecret == "" {
		logger.Fatal("JWT secret is not set")
	}
	logger.Println("JWT secret loaded")

	app := http.NewFiberApp(analyticsService, jwtSecret, logger)
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	go func() {
		if err := app.Listen(viper.GetString("server.analytics_service_address")); err != nil {
			logger.Fatalf("Failed to start analytics server: %v", err)
		}
	}()
	logger.Printf("Analytics Service is running on %s", viper.GetString("server.analytics_service_address"))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := app.Shutdown(); err != nil {
		logger.Fatalf("Failed to shutdown analytics server: %v", err)
	}
	logger.Println("Analytics Service stopped gracefully")
}

func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs/")

	viper.SetDefault("server.analytics_service_address", ":8083")
	viper.SetDefault("db.sslmode", "disable")
	viper.SetDefault("kafka.brokers", []string{"kafka:9092"})
	viper.SetDefault("jwt.secret", "your_secret_key")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}
