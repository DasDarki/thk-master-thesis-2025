package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "2500"
	}

	key := os.Getenv("API_KEY")
	if key == "" {
		key = "thk_masterthesis_2025_hwtwswrtc"
	}

	app := fiber.New()
	app.Use(logger.New())
	app.Use(recover.New())

	app.Get("/csv", func(c *fiber.Ctx) error {
		if c.Get("X-API-KEY") != key {
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}

		if c.Query("emergency") == "true" {
			return c.SendFile("emergency.csv")
		}

		return c.SendFile("results.csv")
	})

	app.Post("/begin", func(c *fiber.Ctx) error {
		if c.Get("X-API-KEY") != key {
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}

		dto := struct {
			Protocol        Protocol
			Enviroment      Enviroment
			TimeSlot        TimeSlot
			ClientID        int
			ParallelClients int
		}{}
		if err := c.BodyParser(&dto); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		run := runManager.beginRun(dto.Protocol, dto.Enviroment, dto.TimeSlot, dto.ClientID, dto.ParallelClients)

		return c.SendString(run.ID.String())
	})

	app.Post("/:id/end", func(c *fiber.Ctx) error {
		if c.Get("X-API-KEY") != key {
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}

		idStr := c.Params("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		data := TestRunData{}
		if err := c.BodyParser(&data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := runManager.endRun(id, data); err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	app.Listen(":" + port)
}
