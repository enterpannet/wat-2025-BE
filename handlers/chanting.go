package handlers

import (
	"registration-system/database"
	"registration-system/models"

	"github.com/gofiber/fiber/v2"
)

type UpdateChantingStatusRequest struct {
	ChantedPariwat bool `json:"chanted_pariwat"`
	ChantedManat   bool `json:"chanted_manat"`
	ChantedOkApan  bool `json:"chanted_ok_apan"`
}

// UpdateChantingStatus updates the chanting status for a registration
func UpdateChantingStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	var registration models.Registration

	// Check if registration exists
	if err := database.DB.First(&registration, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "ไม่พบข้อมูลการลงทะเบียน",
		})
	}

	var req UpdateChantingStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	// Update chanting status
	registration.ChantedPariwat = req.ChantedPariwat
	registration.ChantedManat = req.ChantedManat
	registration.ChantedOkApan = req.ChantedOkApan

	if err := database.DB.Save(&registration).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถอัพเดทสถานะได้",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "อัพเดทสถานะสำเร็จ",
		"data":    registration,
	})
}
