package main

import (
	"log"
	"os"
	"registration-system/database"
	"registration-system/handlers"
	"registration-system/middleware"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	database.Connect()
	database.Migrate()

	app := fiber.New(fiber.Config{
		AppName: "Registration System API",
	})

	// Get CORS origins from environment
	corsOriginsEnv := os.Getenv("CORS_ORIGINS")
	var corsOrigins []string
	if corsOriginsEnv == "" {
		corsOrigins = []string{"http://localhost:5173"} // Default for development
	} else {
		// Split comma-separated origins
		origins := strings.Split(corsOriginsEnv, ",")
		corsOrigins = make([]string, 0, len(origins))
		for _, origin := range origins {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				corsOrigins = append(corsOrigins, origin)
			}
		}
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(corsOrigins, ","),
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
		ExposeHeaders:    "Content-Length, Content-Type",
	}))

	app.Use(logger.New())

	api := app.Group("/api")

	// Public routes - ระบบลงทะเบียนสาธารณะ (ไม่ต้อง login)
	public := api.Group("/public")
	public.Get("/provinces", handlers.GetProvinces)
	public.Get("/provinces/:province_id/districts", handlers.GetDistricts)
	public.Get("/districts/:district_id/sub-districts", handlers.GetSubDistricts)
	public.Post("/registrations", handlers.CreateRegistration)
	public.Post("/device-logs", handlers.CreateDeviceLog) // บันทึกข้อมูลอุปกรณ์ (ไม่ต้อง login - PDPA compliant)

	// Auth routes - สำหรับ admin login/register
	auth := api.Group("/auth")
	auth.Post("/login", handlers.Login)
	auth.Post("/logout", handlers.Logout)
	auth.Post("/register", handlers.RegisterAdmin) // สร้าง admin user ใหม่

	// Admin routes - ต้อง login ก่อน (จัดการข้อมูลที่ลงทะเบียนมา)
	admin := api.Group("/admin", middleware.AuthRequired)
	admin.Get("/me", handlers.GetCurrentUser)
	admin.Get("/registrations", handlers.GetRegistrations)
	admin.Get("/registrations/:id", handlers.GetRegistration)
	admin.Put("/registrations/:id", handlers.UpdateRegistration)
	admin.Delete("/registrations/:id", handlers.DeleteRegistration)
	admin.Put("/registrations/:id/chanting", handlers.UpdateChantingStatus)

	// Activity Log routes - บันทึกการทำกิจกรรม (ต้อง login)
	admin.Get("/activity-logs", handlers.GetActivityLogs)
	admin.Post("/activity-logs", handlers.CreateActivityLog)

	// Summary routes - สรุปข้อมูลทั้งหมด (ต้อง login) - สำหรับ backward compatibility
	admin.Get("/summary", handlers.GetSummary)

	// Device Log routes - บันทึกข้อมูลอุปกรณ์ (ดูต้อง login, สร้างไม่ต้อง)
	admin.Get("/device-logs", handlers.GetDeviceLogs)

	// User Management routes - จัดการผู้ใช้ admin
	admin.Get("/users", handlers.GetAllUsers)
	admin.Put("/users/:id", handlers.UpdateUser)
	admin.Delete("/users/:id", handlers.DeleteUser)

	// Finance routes - ระบบรายรับรายจ่าย (แยกออกมา ไม่ปนกับระบบอื่น)
	// ใช้ login เดียวกัน แต่แยก path ออกมา
	finance := api.Group("/finance", middleware.AuthRequired)
	finance.Get("/transactions", handlers.GetFinanceTransactions)
	finance.Get("/transactions/:id", handlers.GetFinanceTransaction)
	finance.Post("/transactions", handlers.CreateFinanceTransaction)
	finance.Put("/transactions/:id", handlers.UpdateFinanceTransaction)
	finance.Delete("/transactions/:id", handlers.DeleteFinanceTransaction)
	finance.Get("/summary", handlers.GetFinanceSummary)
	finance.Post("/upload-image", handlers.UploadImageToCloudinary) // Upload image to Cloudinary

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
