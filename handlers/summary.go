package handlers

import (
	"registration-system/database"
	"registration-system/models"

	"github.com/gofiber/fiber/v2"
)

// GetSummary - สรุปข้อมูลทั้งหมด
func GetSummary(c *fiber.Ctx) error {
	// นับจำนวนการลงทะเบียน
	var registrationCount int64
	database.DB.Model(&models.Registration{}).Count(&registrationCount)

	// นับจำนวนการลงทะเบียนที่สวดแล้ว
	var pariwatCount int64
	var manatCount int64
	var okApanCount int64
	database.DB.Model(&models.Registration{}).Where("chanted_pariwat = ?", true).Count(&pariwatCount)
	database.DB.Model(&models.Registration{}).Where("chanted_manat = ?", true).Count(&manatCount)
	database.DB.Model(&models.Registration{}).Where("chanted_ok_apan = ?", true).Count(&okApanCount)

	// นับจำนวน Activity Logs
	var activityLogCount int64
	database.DB.Model(&models.ActivityLog{}).Count(&activityLogCount)

	// นับจำนวน Device Logs
	var deviceLogCount int64
	database.DB.Model(&models.DeviceLog{}).Count(&deviceLogCount)

	return c.JSON(fiber.Map{
		"registrations": fiber.Map{
			"total":           registrationCount,
			"chanted_pariwat": pariwatCount,
			"chanted_manat":   manatCount,
			"chanted_ok_apan": okApanCount,
		},
		"logs": fiber.Map{
			"activity_logs": activityLogCount,
			"device_logs":   deviceLogCount,
		},
	})
}

