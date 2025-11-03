package handlers

import (
	"registration-system/database"
	"registration-system/models"
	"time"

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

	// สรุปรายรับรายจ่าย
	var totalIncome float64
	var totalExpense float64
	database.DB.Model(&models.Transaction{}).
		Where("type = ?", "income").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalIncome)
	database.DB.Model(&models.Transaction{}).
		Where("type = ?", "expense").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalExpense)

	// รายรับรายจ่ายในเดือนนี้
	var incomeThisMonth float64
	var expenseThisMonth float64
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	database.DB.Model(&models.Transaction{}).
		Where("type = ? AND date >= ?", "income", startOfMonth).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&incomeThisMonth)
	database.DB.Model(&models.Transaction{}).
		Where("type = ? AND date >= ?", "expense", startOfMonth).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&expenseThisMonth)

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
		"transactions": fiber.Map{
			"total_income":    totalIncome,
			"total_expense":   totalExpense,
			"balance":         totalIncome - totalExpense,
			"income_this_month": incomeThisMonth,
			"expense_this_month": expenseThisMonth,
			"balance_this_month": incomeThisMonth - expenseThisMonth,
		},
		"logs": fiber.Map{
			"activity_logs": activityLogCount,
			"device_logs":   deviceLogCount,
		},
	})
}

// GetFinanceSummary - สรุปข้อมูลเฉพาะส่วนรายรับรายจ่าย
func GetFinanceSummary(c *fiber.Ctx) error {
	// สรุปรายรับรายจ่าย
	var totalIncome float64
	var totalExpense float64
	database.DB.Model(&models.Transaction{}).
		Where("type = ?", "income").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalIncome)
	database.DB.Model(&models.Transaction{}).
		Where("type = ?", "expense").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalExpense)

	// รายรับรายจ่ายในเดือนนี้
	var incomeThisMonth float64
	var expenseThisMonth float64
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	database.DB.Model(&models.Transaction{}).
		Where("type = ? AND date >= ?", "income", startOfMonth).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&incomeThisMonth)
	database.DB.Model(&models.Transaction{}).
		Where("type = ? AND date >= ?", "expense", startOfMonth).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&expenseThisMonth)

	// นับจำนวนรายการ
	var totalTransactions int64
	var incomeCount int64
	var expenseCount int64
	database.DB.Model(&models.Transaction{}).Count(&totalTransactions)
	database.DB.Model(&models.Transaction{}).Where("type = ?", "income").Count(&incomeCount)
	database.DB.Model(&models.Transaction{}).Where("type = ?", "expense").Count(&expenseCount)

	return c.JSON(fiber.Map{
		"total_income":        totalIncome,
		"total_expense":       totalExpense,
		"balance":             totalIncome - totalExpense,
		"income_this_month":   incomeThisMonth,
		"expense_this_month":  expenseThisMonth,
		"balance_this_month":  incomeThisMonth - expenseThisMonth,
		"total_transactions":  totalTransactions,
		"income_count":        incomeCount,
		"expense_count":       expenseCount,
	})
}

