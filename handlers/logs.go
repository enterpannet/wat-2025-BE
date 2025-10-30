package handlers

import (
	"registration-system/database"
	"registration-system/models"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ActivityLogRequest struct {
	Action      string `json:"action"`
	Description string `json:"description"`
	Module      string `json:"module"`
}

type DeviceLogRequest struct {
	DeviceType  string `json:"device_type"`
	DeviceInfo  string `json:"device_info"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Module      string `json:"module"`
	IPAddress   string `json:"ip_address"`
}

// GetActivityLogs - ดึงบันทึกการทำกิจกรรม (ต้อง login)
func GetActivityLogs(c *fiber.Ctx) error {
	var logs []models.ActivityLog

	query := database.DB.Order("created_at DESC")

	// Filter ตาม user (ถ้าต้องการ)
	userID := c.Locals("userID")
	if userID != nil {
		// query = query.Where("user_id = ?", userID)
	}

	result := query.Preload("User").Find(&logs)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถดึงข้อมูลได้",
		})
	}

	return c.JSON(logs)
}

// CreateActivityLog - สร้างบันทึกการทำกิจกรรม (ต้อง login)
func CreateActivityLog(c *fiber.Ctx) error {
	var req ActivityLogRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	userID := c.Locals("userID").(uint)

	log := models.ActivityLog{
		Action:      req.Action,
		Description: req.Description,
		Module:      req.Module,
		UserID:      userID,
	}

	result := database.DB.Create(&log)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถบันทึกข้อมูลได้",
		})
	}

	database.DB.Preload("User").First(&log, log.ID)
	return c.Status(201).JSON(log)
}

// CreateDeviceLog - สร้างบันทึกข้อมูลอุปกรณ์ (ไม่ต้อง login - PDPA compliant)
func CreateDeviceLog(c *fiber.Ctx) error {
	var req DeviceLogRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	// รับ IP address จาก request
	ipAddress := c.IP()
	if req.IPAddress != "" {
		ipAddress = req.IPAddress
	}

	// ปรับปรุง IP address เพื่อความเป็นส่วนตัว (PDPA)
	// ลบส่วนสุดท้ายของ IP v4 หรือ hash สำหรับ IP v6
	if strings.Contains(ipAddress, ".") {
		parts := strings.Split(ipAddress, ".")
		if len(parts) == 4 {
			ipAddress = strings.Join(parts[:3], ".") + ".xxx" // แสดงแค่ 3 ส่วนแรก
		}
	}

	log := models.DeviceLog{
		DeviceType:  req.DeviceType,
		DeviceInfo:  req.DeviceInfo,
		Action:      req.Action,
		Description: req.Description,
		Module:      req.Module,
		IPAddress:   ipAddress,
	}

	result := database.DB.Create(&log)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถบันทึกข้อมูลได้",
		})
	}

	return c.Status(201).JSON(log)
}

// GetDeviceLogs - ดึงบันทึกข้อมูลอุปกรณ์ (ต้อง login เพื่อดู)
func GetDeviceLogs(c *fiber.Ctx) error {
	var logs []models.DeviceLog

	result := database.DB.Order("created_at DESC").Find(&logs)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถดึงข้อมูลได้",
		})
	}

	return c.JSON(logs)
}

