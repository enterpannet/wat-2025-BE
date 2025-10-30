package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"registration-system/database"
	"registration-system/middleware"
	"registration-system/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	User    *UserResponse `json:"user,omitempty"`
}

type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

// Login handles user login with HTTP-only cookies
func Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "กรุณากรอกชื่อผู้ใช้และรหัสผ่าน",
		})
	}

	// Find user
	var user models.User
	if err := database.DB.Where("username = ? AND is_active = ?", req.Username, true).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง",
		})
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง",
		})
	}

	// Generate session token and store it
	sessionToken := generateSessionToken(user.ID)
	middleware.StoreSession(sessionToken, user.ID)

	// Set HTTP-only cookie with session
	cookie := new(fiber.Cookie)
	cookie.Name = "session_id"
	cookie.Value = sessionToken
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.HTTPOnly = true
	cookie.Secure = false // Set to true in production with HTTPS
	cookie.SameSite = "Lax"

	c.Cookie(cookie)

	return c.JSON(LoginResponse{
		Success: true,
		Message: "เข้าสู่ระบบสำเร็จ",
		User: &UserResponse{
			ID:       user.ID,
			Username: user.Username,
			FullName: user.FullName,
		},
	})
}

// Logout handles user logout
func Logout(c *fiber.Ctx) error {
	// Get and delete session
	sessionID := c.Cookies("session_id")
	if sessionID != "" {
		middleware.DeleteSession(sessionID)
	}

	// Clear cookie
	cookie := new(fiber.Cookie)
	cookie.Name = "session_id"
	cookie.Value = ""
	cookie.Expires = time.Now().Add(-1 * time.Hour)
	cookie.HTTPOnly = true

	c.Cookie(cookie)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "ออกจากระบบสำเร็จ",
	})
}

// GetCurrentUser returns the currently logged in user
func GetCurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "ไม่พบข้อมูลผู้ใช้",
		})
	}

	return c.JSON(UserResponse{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
	})
}

// RegisterAdmin creates a new admin user
func RegisterAdmin(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	// Validate input
	if req.Username == "" || req.Password == "" || req.FullName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "กรุณากรอกข้อมูลให้ครบถ้วน",
		})
	}

	// Check if username already exists
	var existingUser models.User
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ชื่อผู้ใช้นี้ถูกใช้งานแล้ว",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการสร้างบัญชี",
		})
	}

	// Create user
	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		FullName: req.FullName,
		IsActive: true,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการสร้างบัญชี",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "สร้างบัญชีสำเร็จ",
		"user": UserResponse{
			ID:       user.ID,
			Username: user.Username,
			FullName: user.FullName,
		},
	})
}

// generateSessionToken creates a simple session token
func generateSessionToken(userID uint) string {
	// Generate random bytes
	bytes := make([]byte, 16)
	rand.Read(bytes)
	randomString := hex.EncodeToString(bytes)

	// Create session token with userID prefix for easy lookup
	return fmt.Sprintf("%d-%s", userID, randomString)
}
