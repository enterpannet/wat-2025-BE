package handlers

import (
	"fmt"
	"registration-system/database"
	"registration-system/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

// FinanceTransactionRequest - Request body for finance transactions
type FinanceTransactionRequest struct {
	Type        string   `json:"type"`        // "income" or "expense"
	Amount      float64  `json:"amount"`     // จำนวนเงิน
	Description string   `json:"description"` // รายละเอียด
	Date        string   `json:"date"`       // Format: "2006-01-02"
	Category    string   `json:"category"`    // หมวดหมู่ เช่น "บุญบารมี", "ค่าใช้จ่ายทั่วไป"
	ImageURLs   []string `json:"image_urls,omitempty"`   // URLs ของภาพจาก Cloudinary (สูงสุด 5 ภาพ)
}

// GetFinanceTransactions - ดึงรายการรายรับรายจ่ายทั้งหมด (Finance System)
func GetFinanceTransactions(c *fiber.Ctx) error {
	var transactions []models.Transaction

	// Query parameters สำหรับ filter
	typeFilter := c.Query("type") // "income" or "expense"
	categoryFilter := c.Query("category")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := database.DB.Order("date DESC, created_at DESC")

	// Filter by type
	if typeFilter == "income" || typeFilter == "expense" {
		query = query.Where("type = ?", typeFilter)
	}

	// Filter by category
	if categoryFilter != "" {
		query = query.Where("category = ?", categoryFilter)
	}

	// Filter by date range
	if startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("date >= ?", date)
		}
	}
	if endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("date <= ?", date)
		}
	}

	result := query.Preload("User").Find(&transactions)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถดึงข้อมูลได้",
		})
	}

	return c.JSON(transactions)
}

// GetFinanceTransaction - ดึงรายการรายรับรายจ่ายรายการเดียว
func GetFinanceTransaction(c *fiber.Ctx) error {
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

// CreateFinanceTransaction - สร้างรายการรายรับรายจ่ายใหม่
func CreateFinanceTransaction(c *fiber.Ctx) error {
	var req FinanceTransactionRequest

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

	if req.Description == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "กรุณากรอกรายละเอียด",
		})
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "รูปแบบวันที่ไม่ถูกต้อง (ใช้ YYYY-MM-DD)",
		})
	}

	userID := c.Locals("userID").(uint)

	// Validate image URLs (max 5 images)
	if len(req.ImageURLs) > 5 {
		return c.Status(400).JSON(fiber.Map{
			"error": "สามารถอัพโหลดได้สูงสุด 5 ภาพ",
		})
	}

	transaction := models.Transaction{
		Type:        req.Type,
		Amount:      req.Amount,
		Description: req.Description,
		Date:        date,
		Category:    req.Category,
		ImageURLs:   models.StringArray(req.ImageURLs),
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
		Action:      "เพิ่มรายการ " + req.Type + " (Finance)",
		Description: req.Description + " - จำนวน " + formatFinanceMoney(req.Amount),
		Module:      "finance",
		UserID:      userID,
	}
	database.DB.Create(&activityLog)

	return c.Status(201).JSON(transaction)
}

// UpdateFinanceTransaction - อัปเดตรายการรายรับรายจ่าย
func UpdateFinanceTransaction(c *fiber.Ctx) error {
	id := c.Params("id")
	var transaction models.Transaction

	if err := database.DB.First(&transaction, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "ไม่พบข้อมูล",
		})
	}

	var req FinanceTransactionRequest
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
	// Update image URLs if provided (max 5 images)
	if req.ImageURLs != nil {
		if len(req.ImageURLs) > 5 {
			return c.Status(400).JSON(fiber.Map{
				"error": "สามารถอัพโหลดได้สูงสุด 5 ภาพ",
			})
		}
		transaction.ImageURLs = models.StringArray(req.ImageURLs)
	}

	if err := database.DB.Save(&transaction).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถอัพเดทข้อมูลได้",
		})
	}

	// บันทึก Activity Log
	userID := c.Locals("userID").(uint)
	activityLog := models.ActivityLog{
		Action:      "แก้ไขรายการ " + transaction.Type + " (Finance)",
		Description: transaction.Description + " - จำนวน " + formatFinanceMoney(transaction.Amount),
		Module:      "finance",
		UserID:      userID,
	}
	database.DB.Create(&activityLog)

	database.DB.Preload("User").First(&transaction, transaction.ID)
	return c.JSON(transaction)
}

// DeleteFinanceTransaction - ลบรายการรายรับรายจ่าย
func DeleteFinanceTransaction(c *fiber.Ctx) error {
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
		Action:      "ลบรายการ " + transaction.Type + " (Finance)",
		Description: transaction.Description + " - จำนวน " + formatFinanceMoney(transaction.Amount),
		Module:      "finance",
		UserID:      userID,
	}
	database.DB.Create(&activityLog)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "ลบข้อมูลสำเร็จ",
	})
}

// GetFinanceSummary - สรุปข้อมูลรายรับรายจ่าย
func GetFinanceSummary(c *fiber.Ctx) error {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var totalIncome float64
	var totalExpense float64
	var incomeCount int64
	var expenseCount int64

	// Build base conditions
	var conditions []interface{}
	var whereClause string

	if startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			whereClause += "date >= ?"
			conditions = append(conditions, date)
		}
	}
	if endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			if whereClause != "" {
				whereClause += " AND "
			}
			whereClause += "date <= ?"
			conditions = append(conditions, date)
		}
	}

	// Calculate income totals
	incomeQuery := database.DB.Model(&models.Transaction{}).Where("type = ?", "income")
	if whereClause != "" {
		incomeQuery = incomeQuery.Where(whereClause, conditions...)
	}
	incomeQuery.Select("COALESCE(SUM(amount), 0)").Scan(&totalIncome)
	incomeQuery.Count(&incomeCount)

	// Calculate expense totals
	expenseQuery := database.DB.Model(&models.Transaction{}).Where("type = ?", "expense")
	if whereClause != "" {
		expenseQuery = expenseQuery.Where(whereClause, conditions...)
	}
	expenseQuery.Select("COALESCE(SUM(amount), 0)").Scan(&totalExpense)
	expenseQuery.Count(&expenseCount)

	netAmount := totalIncome - totalExpense

	// Get categories breakdown
	var categories []struct {
		Category string  `json:"category"`
		Type     string  `json:"type"`
		Total    float64 `json:"total"`
		Count    int64   `json:"count"`
	}

	categoryQuery := database.DB.Model(&models.Transaction{}).
		Select("category, type, COALESCE(SUM(amount), 0) as total, COUNT(*) as count").
		Group("category, type")

	if startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			categoryQuery = categoryQuery.Where("date >= ?", date)
		}
	}
	if endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			categoryQuery = categoryQuery.Where("date <= ?", date)
		}
	}

	categoryQuery.Scan(&categories)

	return c.JSON(fiber.Map{
		"summary": fiber.Map{
			"total_income":  totalIncome,
			"total_expense": totalExpense,
			"net_amount":    netAmount,
			"income_count":  incomeCount,
			"expense_count": expenseCount,
		},
		"categories": categories,
	})
}

// Helper function - format money for display
func formatFinanceMoney(amount float64) string {
	return fmt.Sprintf("฿%.2f", amount)
}

