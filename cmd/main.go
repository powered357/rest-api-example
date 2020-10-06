package main

import (
	"context"

	"github.com/evt/simple-web-server/config"
	"github.com/evt/simple-web-server/controller"
	"github.com/evt/simple-web-server/db"
	libError "github.com/evt/simple-web-server/lib/error"
	"github.com/evt/simple-web-server/lib/validator"
	"github.com/evt/simple-web-server/logger"
	"github.com/evt/simple-web-server/repository/pg"
	"github.com/evt/simple-web-server/service/web"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"log"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	// config
	cfg := config.Get()

	// logger
	l := logger.Get()

	// connect to Postgres
	pgDB, err := db.Dial(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Run Postgres migrations
	log.Println("Running PostgreSQL migrations...")
	if err := runMigrations(cfg); err != nil {
		log.Fatal(err)
	}

	// Init repositories
	userRepo := pg.NewUserRepo(pgDB)

	// Init services
	userService := web.NewUserWebService(ctx, userRepo)

	// Init controllers
	userController := controller.NewUsers(ctx, userService, l)

	// Initialize Echo instance
	e := echo.New()
	e.Validator = validator.NewValidator()
	e.HTTPErrorHandler = libError.Error

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	userRoutes := e.Group("/users")
	userRoutes.GET("/:id", userController.Get)
	userRoutes.DELETE("/:id", userController.Delete)
	//userRoutes.PUT("/:id", userController.Update)
	userRoutes.POST("/", userController.Create)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))

	return nil
}

// runMigrations runs Postgres migrations
func runMigrations(cfg *config.Config) error {
	if cfg.PgMigrationsPath == "" {
		return errors.New("No cfg.PgMigrationsPath provided")
	}
	if cfg.PgURL == "" {
		return errors.New("No cfg.PgURL provided")
	}
	m, err := migrate.New(
		cfg.PgMigrationsPath,
		cfg.PgURL,
	)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
