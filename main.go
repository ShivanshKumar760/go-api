package main

import (
	"errors"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. check the settings
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET must be set and at least 32 characters")
	}

	// 2. connect to the database and create the tables
	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Fatal("db connect: ", err)
	}
	if err := db.AutoMigrate(&User{}, &Post{}); err != nil {
		log.Fatal("migrate: ", err)
	}

	// 3. create the web app. Every error becomes a JSON reply like {"error": "..."}
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			msg := "internal error"
			var e *fiber.Error
			if errors.As(err, &e) { // errors we made with fiber.NewError
				code, msg = e.Code, e.Message
			} else {
				log.Println("error:", err) // log the real reason, hide it from the client
			}
			return c.Status(code).JSON(fiber.Map{"error": msg})
		},
	})
	app.Use(recover.New(), logger.New()) // recover = a crash becomes a 500, not a dead server

	// 4. routes
	app.Get("/health", func(c fiber.Ctx) error { return c.SendString("ok") })

	app.Post("/auth/register", register(db))
	app.Post("/auth/login", login(db))

	// public: anyone can read
	app.Get("/posts", listPosts(db))
	app.Get("/posts/:id", getPost(db))

	// protected: authRequired runs first and stops the request if the token is bad
	app.Post("/posts", authRequired, createPost(db))
	app.Put("/posts/:id", authRequired, updatePost(db))
	app.Delete("/posts/:id", authRequired, deletePost(db))

	// 5. start (Cloud Run and most hosts give us PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
}
