package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schema embed.FS

type postInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type postResponse struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func main() {
	ctx := context.Background()
	database, err := connectDatabase(ctx)
	if err != nil {
		panic(err)
	}
	defer database.Close()

	if err := migrate(ctx, database); err != nil {
		panic(err)
	}

	app := fiber.New()
	app.Use(logger.New(logger.Config{
		Format: "${time} ${status} - ${latency} ${method} ${path}\n",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"app":      "gofiber - demo",
			"datetime": time.Now().Format("2006-01-02 3:4:5 PM"),
			"email":    os.Getenv("EMAIL"),
		})
	})

	app.Post("/posts", createPost(database))
	app.Get("/posts", listPosts(database))
	app.Get("/posts/:id", getPost(database))
	app.Put("/posts/:id", updatePost(database))
	app.Delete("/posts/:id", deletePost(database))

	if err := app.Listen(":8000"); err != nil {
		panic(err)
	}
}

func connectDatabase(ctx context.Context) (*pgxpool.Pool, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
			url.QueryEscape(envOrDefault("POSTGRES_USER", "postgres")),
			url.QueryEscape(os.Getenv("POSTGRES_PASSWORD")),
			envOrDefault("POSTGRES_HOST", "localhost"),
			envOrDefault("POSTGRES_PORT", "5432"),
			envOrDefault("POSTGRES_DB", "postgres"),
		)
	}

	database, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	if err := database.Ping(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return database, nil
}

func migrate(ctx context.Context, database *pgxpool.Pool) error {
	schemaSQL, err := schema.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	if _, err := database.Exec(ctx, string(schemaSQL)); err != nil {
		return fmt.Errorf("run migration: %w", err)
	}
	return nil
}

func createPost(database *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input postInput
		if err := c.BodyParser(&input); err != nil {
			return fiber.ErrBadRequest
		}
		if strings.TrimSpace(input.Title) == "" || input.Content == "" {
			return fiber.NewError(fiber.StatusBadRequest, "title and content are required")
		}

		_, err := database.Exec(c.Context(), "INSERT INTO posts (title, content) VALUES ($1, $2)", input.Title, input.Content)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.Status(fiber.StatusCreated).JSON(postResponse(input))
	}
}

func listPosts(database *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := database.Query(c.Context(), "SELECT title, content FROM posts ORDER BY id")
		if err != nil {
			return fiber.ErrInternalServerError
		}
		defer rows.Close()

		posts := make([]postResponse, 0)
		for rows.Next() {
			var post postResponse
			if err := rows.Scan(&post.Title, &post.Content); err != nil {
				return fiber.ErrInternalServerError
			}
			posts = append(posts, post)
		}
		if err := rows.Err(); err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(posts)
	}
}

func getPost(database *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := postID(c)
		if err != nil {
			return fiber.ErrBadRequest
		}

		var post postResponse
		err = database.QueryRow(c.Context(), "SELECT title, content FROM posts WHERE id = $1", id).Scan(&post.Title, &post.Content)
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.ErrNotFound
		}
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(post)
	}
}

func updatePost(database *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := postID(c)
		if err != nil {
			return fiber.ErrBadRequest
		}
		var input postInput
		if err := c.BodyParser(&input); err != nil {
			return fiber.ErrBadRequest
		}
		if strings.TrimSpace(input.Title) == "" || input.Content == "" {
			return fiber.NewError(fiber.StatusBadRequest, "title and content are required")
		}

		result, err := database.Exec(c.Context(), "UPDATE posts SET title = $1, content = $2, updated_at = NOW() WHERE id = $3", input.Title, input.Content, id)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		if result.RowsAffected() == 0 {
			return fiber.ErrNotFound
		}
		return c.JSON(postResponse(input))
	}
}

func deletePost(database *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := postID(c)
		if err != nil {
			return fiber.ErrBadRequest
		}
		result, err := database.Exec(c.Context(), "DELETE FROM posts WHERE id = $1", id)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		if result.RowsAffected() == 0 {
			return fiber.ErrNotFound
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

func postID(c *fiber.Ctx) (int64, error) {
	return strconv.ParseInt(c.Params("id"), 10, 64)
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
