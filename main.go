package main

import (
	"golang-fiber-client-poc/app/healthcheck"
	"golang-fiber-client-poc/pkg/handler"
	_ "golang-fiber-client-poc/pkg/log"
	"golang-fiber-client-poc/pkg/tracer"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
	recover "github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

func main() {

	tracer.InitTracer()

	app := fiber.New(fiber.Config{
		IdleTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		Concurrency:  256 * 1024,
	})

	app.Use(recover.New())
	app.Use(otelfiber.Middleware())

	// Sağlık kontrolü endpoint'i ekleyelim
	healthHandler := healthcheck.NewHealthCheckHandler()
	app.Get("/health", handler.Handle(healthHandler))

	app.Get("/test", func(c *fiber.Ctx) error {
		zap.L().Info("Test request")
		return c.SendString("Hello, World 👋!")
	})

	app.Get("/timeout", func(c *fiber.Ctx) error {
		time.Sleep(12 * time.Second)
		zap.L().Info("Timeout request completed")
		return c.SendString("timeout")
	})

	app.Get("/error", func(c *fiber.Ctx) error {
		zap.L().Info("Error request")
		return c.Status(fiber.StatusInternalServerError).SendString("error")
	})

	go func() {
		if err := app.Listen(":8081"); err != nil {
			zap.L().Error("Failed to start server", zap.Error(err))
			os.Exit(1)
		}
	}()

	zap.L().Info("Server started on port 8081")

	gracefulShutdown(app)
}

func gracefulShutdown(app *fiber.App) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	zap.L().Info("Shutting down server...")

	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		zap.L().Error("Failed to shutdown server", zap.Error(err))
	}

	zap.L().Info("Server shutdown successfully")
}
