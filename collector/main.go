package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"
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

	db, err := gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	if err := db.AutoMigrate(&TestRun{}); err != nil {
		panic("failed to migrate database")
	}

	app := fiber.New()
	app.Use(logger.New())
	app.Use(recover.New())

	app.Get("/csv", func(c *fiber.Ctx) error {
		if c.Get("X-API-KEY") != key {
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}

		runs := []TestRun{}
		if err := db.Find(&runs).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		csv := exportToCsv(runs)
		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", "attachment; filename=results.csv")
		return c.SendString(csv)
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

		run := TestRun{
			Protocol:        dto.Protocol,
			Enviroment:      dto.Enviroment,
			TimeSlot:        dto.TimeSlot,
			ClientID:        dto.ClientID,
			ParallelClients: dto.ParallelClients,
			TestBegin:       time.Now(),
		}

		if err := db.Create(&run).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.SendString(fmt.Sprintf("%d", run.ID))
	})

	updateRun := func(c *fiber.Ctx, run *TestRun) error {
		dto := map[string]any{}
		if err := c.BodyParser(&dto); err != nil {
			return err
		}

		nfe := errors.New("key not found")

		getInt64 := func(key string) (int64, error) {
			if val, ok := dto[key]; ok {
				switch v := val.(type) {
				case string:
					return strconv.ParseInt(v, 10, 64)
				case float64:
					return int64(v), nil
				case int64:
					return v, nil
				case int:
					return int64(v), nil
				default:
					return 0, fmt.Errorf("type %T not supported", v)
				}
			}
			return 0, nfe
		}

		getFloat64 := func(key string) (float64, error) {
			if val, ok := dto[key]; ok {
				switch v := val.(type) {
				case string:
					return strconv.ParseFloat(v, 64)
				case float64:
					return v, nil
				case int64:
					return float64(v), nil
				case int:
					return float64(v), nil
				default:
					return 0, fmt.Errorf("type %T not supported", v)
				}
			}
			return 0, nfe
		}

		getString := func(key string) (string, error) {
			if val, ok := dto[key]; ok {
				switch v := val.(type) {
				case string:
					return v, nil
				case float64:
					return strconv.FormatFloat(v, 'f', -1, 64), nil
				case int64:
					return strconv.FormatInt(v, 10), nil
				case int:
					return strconv.Itoa(v), nil
				default:
					return "", fmt.Errorf("type %T not supported", v)
				}
			}
			return "", nfe
		}

		if v, err := getInt64("TransferStartUnix"); err == nil {
			run.TransferStartUnix = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getInt64("TransferEndUnix"); err == nil {
			run.TransferEndUnix = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getFloat64("ThroughputMbps"); err == nil {
			run.ThroughputMbps = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getInt64("BytesSentTotal"); err == nil {
			run.BytesSentTotal = int64(v)
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getInt64("BytesPayload"); err == nil {
			run.BytesPayload = int64(v)
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getFloat64("CpuClientPercentBefore"); err == nil {
			run.CpuClientPercentBefore = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getFloat64("CpuClientPercentAfter"); err == nil {
			run.CpuClientPercentAfter = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getFloat64("CpuClientPercentWhile"); err == nil {
			run.CpuClientPercentWhile = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getFloat64("CpuServerPercentBefore"); err == nil {
			run.CpuServerPercentBefore = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getFloat64("CpuServerPercentAfter"); err == nil {
			run.CpuServerPercentAfter = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getFloat64("CpuServerPercentWhile"); err == nil {
			run.CpuServerPercentWhile = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getInt64("RamClientBytesBefore"); err == nil {
			run.RamClientBytesBefore = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getInt64("RamClientBytesAfter"); err == nil {
			run.RamClientBytesAfter = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getInt64("RamClientBytesWhile"); err == nil {
			run.RamClientBytesWhile = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getInt64("RamServerBytesBefore"); err == nil {
			run.RamServerBytesBefore = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getInt64("RamServerBytesAfter"); err == nil {
			run.RamServerBytesAfter = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getInt64("RamServerBytesWhile"); err == nil {
			run.RamServerBytesWhile = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getInt64("LostPackets"); err == nil {
			run.LostPackets = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getInt64("Retransmissions"); err == nil {
			run.Retransmissions = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getInt64("ConnectionDuration"); err == nil {
			run.ConnectionDuration = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getInt64("StreamDuration"); err == nil {
			run.StreamDuration = v
		} else if err.Error() != "key not found" {
			return err
		}

		if v, err := getString("Error"); err == nil {
			run.Error = v
		} else if err.Error() != "key not found" {
			return err
		}

		return nil
	}

	app.Put("/:id/update", func(c *fiber.Ctx) error {
		if c.Get("X-API-KEY") != key {
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}

		idStr := c.Params("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		run := TestRun{}
		if err := db.First(&run, id).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}

		if err := updateRun(c, &run); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := db.Save(&run).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	app.Post("/:id/end", func(c *fiber.Ctx) error {
		if c.Get("X-API-KEY") != key {
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}

		idStr := c.Params("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		run := TestRun{}
		if err := db.First(&run, id).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}

		if err := updateRun(c, &run); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		run.TestEnd = time.Now()

		if err := db.Save(&run).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})

	app.Listen(":" + port)
}
