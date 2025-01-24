package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "recommendation-system/pkg/logger"

	"github.com/spf13/viper"
	fiberSwagger "github.com/swaggo/fiber-swagger"

	_ "recommendation-system/cmd/sso-service/docs"
	"recommendation-system/internal/sso/delivery/http"
	"recommendation-system/internal/sso/repository"
	"recommendation-system/internal/sso/service"
	"recommendation-system/pkg/db"
	"recommendation-system/pkg/kafka"
)

// @title           SSO Service API
// @version         1.0
// @description     API for Single Sign-On (SSO) functionalities.
// @host            localhost:8084
// @BasePath        /api

func main() {
	logFile := "logger/logger.log"
	logger, err := log.NewLogger(logFile, "sso-service", "sso-app", "development")
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

	jwtSecret := viper.GetString("jwt.secret")
	if jwtSecret == "" {
		logger.Fatal("JWT secret is not set")
	}
	logger.Println("JWT secret loaded")

	userRepo := repository.NewUserRepository(database, logger)
	userService := service.NewUserService(userRepo, kafkaClient, []byte(jwtSecret), logger)
	userHandler := http.NewHandler(userService, logger)

	app := http.NewFiberApp(userHandler)

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	go func() {
		if err := app.Listen(viper.GetString("server.sso_service_address")); err != nil {
			logger.Fatalf("Failed to start SSO server: %v", err)
		}
	}()

	logger.Printf("SSO Service is running on %s", viper.GetString("server.sso_service_address"))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := app.Shutdown(); err != nil {
		logger.Fatalf("Failed to shutdown SSO server: %v", err)
	}
	logger.Println("SSO Service stopped gracefully")
}

func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs/")

	viper.SetDefault("server.sso_service_address", ":8084")
	viper.SetDefault("db.sslmode", "disable")
	viper.SetDefault("kafka.brokers", []string{"kafka:9092"})
	viper.SetDefault("redis.host", "redis:6379")
	viper.SetDefault("jwt.secret", "your_secret_key")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}
