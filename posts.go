package main

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type postInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func readInput(c fiber.Ctx) (postInput, error) {
	var in postInput

	if err := c.Bind().Body(&in); err != nil {
		return in, fiber.NewError(fiber.StatusBadRequest, "invalid JSON")
	}

	in.Title = strings.TrimSpace(in.Title)
	in.Content = strings.TrimSpace(in.Content)

	if in.Title == "" || in.Content == "" {
		return in, fiber.NewError(fiber.StatusBadRequest, "title and content are required")

	}

	return in, nil
}

func findPost(db *gorm.DB, c fiber.Ctx) (Post, error) {
	id := fiber.Params[int](c, "id", 0) // 0 if the url part is not a number
	if id <= 0 {
		return Post{}, fiber.NewError(fiber.StatusBadRequest, "id must be a positive number")
	}
	var post Post
	err := db.First(&post, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Post{}, fiber.NewError(fiber.StatusNotFound, "post not found")
	}

	if err != nil {
		return Post{}, err
	}

	return post, nil
}

func listPosts(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		var posts []Post
		err := db.Order("id desc").Limit(50).Find(&posts).Error //newest first

		if err != nil {
			return err
		}

		return c.JSON(posts)
	}
}

func getPost(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		post, err := findPost(db, c)
		if err != nil {
			return err
		}

		return c.JSON(post)
	}
}

func createPost(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		in, err := readInput(c)
		if err != nil {
			return err
		}

		post := Post{
			Title:    in.Title,
			Content:  in.Content,
			AuthorID: fiber.Locals[uint](c, "userID"),
		}

		if err := db.Create(&post).Error; err != nil {
			return err
		}

		return c.Status(fiber.StatusCreated).JSON(post)
	}
}

func updatePost(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		post, err := findPost(db, c)
		if err != nil {
			return nil
		}

		if post.AuthorID != fiber.Locals[uint](c, "userID") {
			return fiber.NewError(fiber.StatusForbidden, "you can only edit your own posts")
		}
		in, err := readInput(c)
		if err != nil {
			return nil
		}

		post.Title = in.Title
		post.Content = in.Content
		if err := db.Save(&post).Error; err != nil {
			return err
		}

		return c.JSON(post)
	}
}

func deletePost(db *gorm.DB) fiber.Handler {
	return func(c fiber.Ctx) error {
		post, err := findPost(db, c)
		if err != nil {
			return err
		}
		if post.AuthorID != fiber.Locals[uint](c, "userID") {
			return fiber.NewError(fiber.StatusForbidden, "you can only delete your own posts")
		}

		if err := db.Delete(&post).Error; err != nil {
			return err
		}
		return c.SendStatus(fiber.StatusNoContent) // 204: success, nothing to send back
	}
}
