package handlers

import (
	"fmt"
	"registration-system/database"
	"registration-system/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

type TeacherRegistrationRequest struct {
	FullName         string `json:"full_name"`
	Nickname         string `json:"nickname"`
	BirthDate        string `json:"birth_date"`
	ProvinceID       uint   `json:"province_id"`
	DistrictID       uint   `json:"district_id"`
	SubDistrictID    uint   `json:"sub_district_id"`
	AddressDetail    string `json:"address_detail"`
	PhoneNumber      string `json:"phone_number"`
	TempleName       string `json:"temple_name"`
	MedicalCondition string `json:"medical_condition"`
	Vassa            int    `json:"vassa"`
}

func CreateTeacherRegistration(c *fiber.Ctx) error {
	var req TeacherRegistrationRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid birth date format. Use YYYY-MM-DD",
		})
	}

	registration := models.TeacherRegistration{
		FullName:         req.FullName,
		Nickname:         req.Nickname,
		BirthDate:        birthDate,
		ProvinceID:       req.ProvinceID,
		DistrictID:       req.DistrictID,
		SubDistrictID:    req.SubDistrictID,
		AddressDetail:    req.AddressDetail,
		PhoneNumber:      req.PhoneNumber,
		TempleName:       req.TempleName,
		MedicalCondition: req.MedicalCondition,
		Vassa:            req.Vassa,
	}

	if result := database.DB.Create(&registration); result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create teacher registration",
		})
	}

	database.DB.Preload("Province").Preload("District").Preload("SubDistrict").First(&registration, registration.ID)

	return c.Status(201).JSON(registration)
}

func GetTeacherRegistrations(c *fiber.Ctx) error {
	var registrations []models.TeacherRegistration

	if result := database.DB.Preload("Province").Preload("District").Preload("SubDistrict").Order("created_at DESC").Find(&registrations); result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch teacher registrations",
		})
	}

	return c.JSON(registrations)
}

func GetTeacherRegistration(c *fiber.Ctx) error {
	id := c.Params("id")
	var registration models.TeacherRegistration

	if result := database.DB.Preload("Province").Preload("District").Preload("SubDistrict").First(&registration, id); result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Teacher registration not found",
		})
	}

	return c.JSON(registration)
}

func UpdateTeacherRegistration(c *fiber.Ctx) error {
	id := c.Params("id")
	var registration models.TeacherRegistration

	if err := database.DB.First(&registration, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "ไม่พบข้อมูลการลงทะเบียนพระอาจารย์",
		})
	}

	var req TeacherRegistrationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	birthDate, err := time.Parse("2006-01-02", req.BirthDate)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "รูปแบบวันเกิดไม่ถูกต้อง",
		})
	}

	registration.FullName = req.FullName
	registration.Nickname = req.Nickname
	registration.BirthDate = birthDate
	registration.ProvinceID = req.ProvinceID
	registration.DistrictID = req.DistrictID
	registration.SubDistrictID = req.SubDistrictID
	registration.AddressDetail = req.AddressDetail
	registration.PhoneNumber = req.PhoneNumber
	registration.TempleName = req.TempleName
	registration.MedicalCondition = req.MedicalCondition
	registration.Vassa = req.Vassa

	if err := database.DB.Save(&registration).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถอัพเดทข้อมูลได้",
		})
	}

	database.DB.Preload("Province").Preload("District").Preload("SubDistrict").First(&registration, registration.ID)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "อัพเดทข้อมูลสำเร็จ",
		"data":    registration,
	})
}

func DeleteTeacherRegistration(c *fiber.Ctx) error {
	id := c.Params("id")
	var registration models.TeacherRegistration

	if err := database.DB.First(&registration, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "ไม่พบข้อมูลการลงทะเบียนพระอาจารย์",
		})
	}

	userID := c.Locals("userID").(uint)

	if err := database.DB.Delete(&registration).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถลบข้อมูลได้",
		})
	}

	activityLog := models.ActivityLog{
		Action:      "ลบข้อมูลการลงทะเบียนพระอาจารย์",
		Description: fmt.Sprintf("ลบข้อมูลของ %s (เบอร์โทร: %s)", registration.FullName, registration.PhoneNumber),
		Module:      "teacher-registration",
		UserID:      userID,
	}
	database.DB.Create(&activityLog)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "ลบข้อมูลสำเร็จ",
	})
}
