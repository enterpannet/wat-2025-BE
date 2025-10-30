package middleware

import (
	"registration-system/database"
	"registration-system/models"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthRequired middleware checks if user is authenticated
func AuthRequired(c *fiber.Ctx) error {
	// Get session cookie
	sessionID := c.Cookies("session_id")

	if sessionID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "กรุณาเข้าสู่ระบบ",
		})
	}

	// In a simple implementation, we store userID in the cookie value
	// In production, use proper session storage (Redis) or JWT

	// For now, we'll use a simple approach: store sessions in memory
	// This is just for demonstration - use Redis or database in production
	userID := getUserIDFromSession(sessionID)

	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "กรุณาเข้าสู่ระบบ",
		})
	}

	// Verify user exists and is active
	var user models.User
	if err := database.DB.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "กรุณาเข้าสู่ระบบ",
		})
	}

	// Store user ID in context
	c.Locals("userID", user.ID)
	c.Locals("username", user.Username)

	return c.Next()
}

// Simple in-memory session storage (replace with Redis in production)
var sessions = make(map[string]uint)

func StoreSession(sessionID string, userID uint) {
	sessions[sessionID] = userID
}

func getUserIDFromSession(sessionID string) uint {
	if userID, ok := sessions[sessionID]; ok {
		return userID
	}
	return 0
}

func DeleteSession(sessionID string) {
	delete(sessions, sessionID)
}

// ParseSessionID extracts user ID from simple session format
// Format: "userID-randomString"
func ParseSessionID(sessionID string) uint {
	parts := strings.Split(sessionID, "-")
	if len(parts) < 2 {
		return 0
	}

	userID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return 0
	}

	return uint(userID)
}
