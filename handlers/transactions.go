package handlers

import (
	"fmt"
	"registration-system/database"
	"registration-system/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

type TransactionRequest struct {
	Type        string  `json:"type"` // "income" or "expense"
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Date        string  `json:"date"` // Format: "2006-01-02"
	Category    string  `json:"category"`
}

// GetTransactions - ดึงรายการรายรับรายจ่ายทั้งหมด
func GetTransactions(c *fiber.Ctx) error {
	var transactions []models.Transaction

	// ดึงข้อมูลทั้งหมด
	query := database.DB.Order("date DESC, created_at DESC")

	result := query.Preload("User").Find(&transactions)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถดึงข้อมูลได้",
		})
	}

	return c.JSON(transactions)
}

// GetTransaction - ดึงรายการรายรับรายจ่ายรายการเดียว
func GetTransaction(c *fiber.Ctx) error {
	id := c.Params("id")
	var transaction models.Transaction

	result := database.DB.Preload("User").First(&transaction, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "ไม่พบข้อมูล",
		})
	}

	return c.JSON(transaction)
}

// CreateTransaction - สร้างรายการรายรับรายจ่ายใหม่
func CreateTransaction(c *fiber.Ctx) error {
	var req TransactionRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	// Validate
	if req.Type != "income" && req.Type != "expense" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ประเภทต้องเป็น 'income' หรือ 'expense'",
		})
	}

	if req.Amount <= 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "จำนวนเงินต้องมากกว่า 0",
		})
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "รูปแบบวันที่ไม่ถูกต้อง (ใช้ YYYY-MM-DD)",
		})
	}

	userID := c.Locals("userID").(uint)

	transaction := models.Transaction{
		Type:        req.Type,
		Amount:      req.Amount,
		Description: req.Description,
		Date:        date,
		Category:    req.Category,
		UserID:      userID,
	}

	result := database.DB.Create(&transaction)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถบันทึกข้อมูลได้",
		})
	}

	// Load user relationship
	database.DB.Preload("User").First(&transaction, transaction.ID)

	// บันทึก Activity Log
	activityLog := models.ActivityLog{
		Action:      "เพิ่มรายการ " + req.Type,
		Description: req.Description + " - จำนวน " + formatMoney(req.Amount),
		Module:      "transaction",
		UserID:      userID,
	}
	database.DB.Create(&activityLog)

	return c.Status(201).JSON(transaction)
}

// UpdateTransaction - อัปเดตรายการรายรับรายจ่าย
func UpdateTransaction(c *fiber.Ctx) error {
	id := c.Params("id")
	var transaction models.Transaction

	if err := database.DB.First(&transaction, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "ไม่พบข้อมูล",
		})
	}

	var req TransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	if req.Type != "" && req.Type != "income" && req.Type != "expense" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ประเภทต้องเป็น 'income' หรือ 'expense'",
		})
	}

	if req.Date != "" {
		date, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "รูปแบบวันที่ไม่ถูกต้อง",
			})
		}
		transaction.Date = date
	}

	if req.Type != "" {
		transaction.Type = req.Type
	}
	if req.Amount > 0 {
		transaction.Amount = req.Amount
	}
	if req.Description != "" {
		transaction.Description = req.Description
	}
	if req.Category != "" {
		transaction.Category = req.Category
	}

	if err := database.DB.Save(&transaction).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถอัพเดทข้อมูลได้",
		})
	}

	// บันทึก Activity Log
	userID := c.Locals("userID").(uint)
	activityLog := models.ActivityLog{
		Action:      "แก้ไขรายการ " + transaction.Type,
		Description: transaction.Description + " - จำนวน " + formatMoney(transaction.Amount),
		Module:      "transaction",
		UserID:      userID,
	}
	database.DB.Create(&activityLog)

	database.DB.Preload("User").First(&transaction, transaction.ID)
	return c.JSON(transaction)
}

// DeleteTransaction - ลบรายการรายรับรายจ่าย
func DeleteTransaction(c *fiber.Ctx) error {
	id := c.Params("id")
	var transaction models.Transaction

	if err := database.DB.First(&transaction, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "ไม่พบข้อมูล",
		})
	}

	if err := database.DB.Delete(&transaction).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถลบข้อมูลได้",
		})
	}

	// บันทึก Activity Log
	userID := c.Locals("userID").(uint)
	activityLog := models.ActivityLog{
		Action:      "ลบรายการ " + transaction.Type,
		Description: transaction.Description + " - จำนวน " + formatMoney(transaction.Amount),
		Module:      "transaction",
		UserID:      userID,
	}
	database.DB.Create(&activityLog)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "ลบข้อมูลสำเร็จ",
	})
}

// Helper function - format money for display
func formatMoney(amount float64) string {
	return fmt.Sprintf("฿%.2f", amount)
}
