package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
)

// CloudinaryUploadResponse - Response from Cloudinary upload
type CloudinaryUploadResponse struct {
	SecureURL string `json:"secure_url"`
	PublicID  string `json:"public_id"`
}

// UploadImageToCloudinary - Upload image to Cloudinary
func UploadImageToCloudinary(c *fiber.Ctx) error {
	// Get file from form
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ไม่พบไฟล์ภาพ",
		})
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return c.Status(400).JSON(fiber.Map{
			"error": "ขนาดไฟล์ต้องไม่เกิน 10MB",
		})
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedTypes[file.Header.Get("Content-Type")] {
		return c.Status(400).JSON(fiber.Map{
			"error": "ประเภทไฟล์ไม่ถูกต้อง (รองรับเฉพาะ JPEG, PNG, GIF, WebP)",
		})
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถเปิดไฟล์ได้",
		})
	}
	defer src.Close()

	// Read file into buffer
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, src); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถอ่านไฟล์ได้",
		})
	}

	// Prepare multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file field
	part, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถสร้าง form data ได้",
		})
	}
	if _, err := part.Write(buf.Bytes()); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถเขียนไฟล์ได้",
		})
	}

	// Add upload preset (use unsigned upload preset)
	uploadPreset := os.Getenv("CLOUDINARY_UPLOAD_PRESET")
	if uploadPreset == "" {
		uploadPreset = "unsigned" // Default preset
	}
	writer.WriteField("upload_preset", uploadPreset)

	// Add folder
	writer.WriteField("folder", "finance-transactions")

	writer.Close()

	// Get Cloudinary credentials
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	if cloudName == "" {
		return c.Status(500).JSON(fiber.Map{
			"error": "Cloudinary ไม่ได้ตั้งค่า",
		})
	}

	// Upload to Cloudinary
	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cloudName)
	
	req, err := http.NewRequest("POST", uploadURL, &requestBody)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถสร้าง request ได้",
		})
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถอัพโหลดไป Cloudinary ได้",
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Cloudinary error: %s", string(bodyBytes)),
		})
	}

	// Parse response
	var cloudinaryResp CloudinaryUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&cloudinaryResp); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถอ่าน response จาก Cloudinary ได้",
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"image_url": cloudinaryResp.SecureURL,
		"public_id": cloudinaryResp.PublicID,
	})
}

