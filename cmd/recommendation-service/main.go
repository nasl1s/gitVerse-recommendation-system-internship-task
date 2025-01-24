package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "recommendation-system/pkg/logger"

	_ "recommendation-system/cmd/recommendation-service/docs"

	"github.com/spf13/viper"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	"recommendation-system/internal/recommendation/delivery/http"
	"recommendation-system/internal/recommendation/repository"
	"recommendation-system/internal/recommendation/service"
	"recommendation-system/pkg/db"
	"recommendation-system/pkg/kafka"
	"recommendation-system/pkg/redis"

	kafkaGo "github.com/segmentio/kafka-go"
)

// @title           Recommendation Service API
// @version         1.0
// @description     API for managing recommendations in the recommendation system.
// @host            localhost:8082
// @BasePath        /api

func main() {
	logFile := "logger/logger.log"
	logger, err := log.NewLogger(logFile, "recommendation-service", "recommendation-app", "development")
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

	if err := kafkaClient.CreateTopic("recommendation_updates", 1, 1); err != nil {
		logger.Printf("Topic 'recommendation_updates' may already exist or failed to create: %v", err)
	}

	redisHosts := viper.GetStringSlice("redis.host")
	if len(redisHosts) == 0 {
		logger.Fatalf("No Redis hosts specified in the configuration")
	}
	redisClient := redis.NewRedisClient(redisHosts[0], "", 0)
	logger.Println("Redis client initialized")

	recommendationRepo := repository.NewRecommendationRepository(database, logger)
	recommendationService := service.NewRecommendationService(recommendationRepo, kafkaClient, redisClient, logger)
	recommendationHandler := http.NewHandler(recommendationService, logger)

	jwtSecret := viper.GetString("jwt.secret")
	if jwtSecret == "" {
		logger.Fatal("JWT secret is not set")
	}
	logger.Println("JWT secret loaded")

	app := http.NewFiberApp(recommendationHandler, jwtSecret)
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	go func() {
		if err := app.Listen(viper.GetString("server.recommendation_service_address")); err != nil {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()
	logger.Printf("Recommendation Service is running on %s", viper.GetString("server.recommendation_service_address"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		topics := []string{"user_updates", "product_updates"}
		groupID := "recommendation_service_group"
		if err := kafkaClient.SubscribeToTopicsFallback(ctx, topics, groupID, func(m kafkaGo.Message) error {
			return recommendationService.ProcessKafkaMessage(ctx, m)
		}); err != nil {
			logger.Printf("Error subscribing to Kafka topics: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancel()

	time.Sleep(2 * time.Second)

	if err := app.Shutdown(); err != nil {
		logger.Fatalf("Failed to shutdown server: %v", err)
	}
	logger.Println("Recommendation Service stopped gracefully")
}

func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs/")

	viper.SetDefault("server.recommendation_service_address", ":8082")
	viper.SetDefault("db.sslmode", "disable")
	viper.SetDefault("kafka.brokers", []string{"kafka:9092"})
	viper.SetDefault("jwt.secret", "your_secret_key")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}
