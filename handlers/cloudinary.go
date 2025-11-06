package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

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
		log.Printf("Error getting file from form: %v", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "ไม่พบไฟล์ภาพ",
		})
	}

	log.Printf("Received file: %s, Size: %d bytes", file.Filename, file.Size)

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return c.Status(400).JSON(fiber.Map{
			"error": "ขนาดไฟล์ต้องไม่เกิน 10MB",
		})
	}

	if file.Size == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "ไฟล์ภาพว่างเปล่า",
		})
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		log.Printf("Error opening file: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถเปิดไฟล์ได้",
		})
	}
	defer src.Close()

	// Read file into buffer
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, src); err != nil {
		log.Printf("Error reading file: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถอ่านไฟล์ได้",
		})
	}

	// Detect content type from file content (more reliable)
	contentType := http.DetectContentType(buf.Bytes()[:512]) // Read first 512 bytes
	log.Printf("Detected content type: %s", contentType)

	// Validate file type (more flexible)
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"application/octet-stream": true, // Some browsers send this for JPEG
	}
	
	// Also check Content-Type from header as fallback
	headerContentType := file.Header.Get("Content-Type")
	if headerContentType != "" {
		log.Printf("Header content type: %s", headerContentType)
		// Normalize content type (remove parameters)
		headerContentType = strings.Split(headerContentType, ";")[0]
		headerContentType = strings.TrimSpace(headerContentType)
		if allowedTypes[strings.ToLower(headerContentType)] {
			contentType = headerContentType
		}
	}

	// Normalize detected content type
	contentType = strings.Split(contentType, ";")[0]
	contentType = strings.TrimSpace(contentType)
	
	if !allowedTypes[strings.ToLower(contentType)] {
		log.Printf("Invalid content type: %s (header: %s)", contentType, headerContentType)
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("ประเภทไฟล์ไม่ถูกต้อง (รองรับเฉพาะ JPEG, PNG, GIF, WebP) - ได้รับ: %s", contentType),
		})
	}

	// Prepare multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file field
	part, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		log.Printf("Error creating form file: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถสร้าง form data ได้",
		})
	}
	if _, err := part.Write(buf.Bytes()); err != nil {
		log.Printf("Error writing file to form: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถเขียนไฟล์ได้",
		})
	}

	// Add upload preset (use unsigned upload preset)
	uploadPreset := os.Getenv("CLOUDINARY_UPLOAD_PRESET")
	if uploadPreset == "" {
		log.Println("Warning: CLOUDINARY_UPLOAD_PRESET not set, using default")
		uploadPreset = "unsigned" // Default preset
	}
	log.Printf("Using upload preset: %s", uploadPreset)
	writer.WriteField("upload_preset", uploadPreset)

	// Add folder
	writer.WriteField("folder", "finance-transactions")

	writer.Close()

	// Get Cloudinary credentials
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	if cloudName == "" {
		log.Println("Error: CLOUDINARY_CLOUD_NAME not set")
		return c.Status(500).JSON(fiber.Map{
			"error": "Cloudinary ไม่ได้ตั้งค่า",
		})
	}
	log.Printf("Using Cloudinary cloud name: %s", cloudName)

	// Upload to Cloudinary
	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cloudName)
	log.Printf("Uploading to: %s", uploadURL)
	
	req, err := http.NewRequest("POST", uploadURL, &requestBody)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถสร้าง request ได้",
		})
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	client := &http.Client{
		Timeout: 60 * time.Second, // Increase timeout to 60 seconds for larger files
	}
	log.Printf("Starting upload request to Cloudinary...")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error executing request to Cloudinary: %v", err)
		// Check if it's a timeout error
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "i/o timeout") {
			return c.Status(500).JSON(fiber.Map{
				"error": "การอัพโหลดใช้เวลานานเกินไป กรุณาลองใหม่อีกครั้ง หรือลดขนาดภาพ",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("ไม่สามารถเชื่อมต่อกับ Cloudinary ได้: %v", err),
		})
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถอ่าน response จาก Cloudinary ได้",
		})
	}

	log.Printf("Cloudinary response status: %d", resp.StatusCode)
	log.Printf("Cloudinary response body: %s", string(bodyBytes))

	if resp.StatusCode != 200 {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Cloudinary error (status %d): %s", resp.StatusCode, string(bodyBytes)),
		})
	}

	// Parse response
	var cloudinaryResp CloudinaryUploadResponse
	if err := json.Unmarshal(bodyBytes, &cloudinaryResp); err != nil {
		log.Printf("Error parsing Cloudinary response: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("ไม่สามารถอ่าน response จาก Cloudinary ได้: %v", err),
		})
	}

	if cloudinaryResp.SecureURL == "" {
		log.Printf("Error: Cloudinary response missing secure_url")
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่ได้รับ URL ภาพจาก Cloudinary",
		})
	}

	log.Printf("Successfully uploaded image: %s", cloudinaryResp.SecureURL)

	return c.JSON(fiber.Map{
		"success":   true,
		"image_url": cloudinaryResp.SecureURL,
		"public_id": cloudinaryResp.PublicID,
	})
}

