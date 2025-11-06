package handlers

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gofiber/fiber/v2"
)

// CloudinaryUploadResponse - Response from Cloudinary upload
type CloudinaryUploadResponse struct {
	SecureURL string `json:"secure_url"`
	PublicID  string `json:"public_id"`
}

// generateCloudinarySignature - Generate signature for Cloudinary signed upload
func generateCloudinarySignature(params map[string]string, apiSecret string) string {
	// Sort parameters by key
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build signature string
	var signatureParts []string
	for _, k := range keys {
		if k != "file" && k != "api_key" {
			signatureParts = append(signatureParts, fmt.Sprintf("%s=%s", k, params[k]))
		}
	}
	signatureString := strings.Join(signatureParts, "&") + apiSecret

	// Generate SHA-1 hash
	hash := sha1.Sum([]byte(signatureString))
	return hex.EncodeToString(hash[:])
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

	// Resize and compress image if needed
	imageData := buf.Bytes()
	originalSize := len(imageData)
	log.Printf("Original image size: %d bytes", originalSize)

	// Decode image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		log.Printf("Warning: Could not decode image for resizing: %v. Uploading original.", err)
		// Continue with original image if decode fails
	} else {
		// Get image bounds
		bounds := img.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()
		log.Printf("Image dimensions: %dx%d", width, height)

		// Resize if larger than 1920x1920
		maxDimension := 1920
		if width > maxDimension || height > maxDimension {
			var resizedImg image.Image
			if width > height {
				resizedImg = imaging.Resize(img, maxDimension, 0, imaging.Lanczos)
			} else {
				resizedImg = imaging.Resize(img, 0, maxDimension, imaging.Lanczos)
			}
			img = resizedImg
			log.Printf("Resized image dimensions: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
		}

		// Encode as JPEG with compression
		var compressedBuf bytes.Buffer
		err = jpeg.Encode(&compressedBuf, img, &jpeg.Options{Quality: 85})
		if err != nil {
			log.Printf("Warning: Could not compress image: %v. Uploading original.", err)
			// Continue with original image if compression fails
		} else {
			compressedSize := compressedBuf.Len()
			log.Printf("Compressed image size: %d bytes (reduced by %.1f%%)", compressedSize, float64(originalSize-compressedSize)/float64(originalSize)*100)
			
			// Use compressed image if it's smaller
			if compressedSize < originalSize {
				imageData = compressedBuf.Bytes()
				log.Printf("Using compressed image")
			} else {
				log.Printf("Compressed image is larger, using original")
			}
		}
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
	if _, err := part.Write(imageData); err != nil {
		log.Printf("Error writing file to form: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถเขียนไฟล์ได้",
		})
	}

	// Get Cloudinary credentials
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" {
		log.Println("Error: CLOUDINARY_CLOUD_NAME not set")
		return c.Status(500).JSON(fiber.Map{
			"error": "Cloudinary ไม่ได้ตั้งค่า (CLOUDINARY_CLOUD_NAME)",
		})
	}
	if apiKey == "" {
		log.Println("Error: CLOUDINARY_API_KEY not set")
		return c.Status(500).JSON(fiber.Map{
			"error": "Cloudinary ไม่ได้ตั้งค่า (CLOUDINARY_API_KEY)",
		})
	}
	if apiSecret == "" {
		log.Println("Error: CLOUDINARY_API_SECRET not set")
		return c.Status(500).JSON(fiber.Map{
			"error": "Cloudinary ไม่ได้ตั้งค่า (CLOUDINARY_API_SECRET)",
		})
	}

	log.Printf("Using Cloudinary cloud name: %s", cloudName)
	log.Printf("Using Cloudinary API key: %s", apiKey)

	// Generate timestamp
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Prepare parameters for signature
	params := map[string]string{
		"timestamp": timestamp,
		"folder":    "finance-transactions",
	}

	// Generate signature
	signature := generateCloudinarySignature(params, apiSecret)
	log.Printf("Generated signature: %s", signature)

	// Add API key, timestamp, signature, and folder
	writer.WriteField("api_key", apiKey)
	writer.WriteField("timestamp", timestamp)
	writer.WriteField("signature", signature)
	writer.WriteField("folder", "finance-transactions")

	writer.Close()

	// Upload to Cloudinary
	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cloudName)
	log.Printf("Uploading to: %s", uploadURL)

	// Execute request with retry mechanism
	client := &http.Client{
		Timeout: 30 * time.Second, // Use 30 seconds timeout per attempt
	}
	
	var resp *http.Response
	maxRetries := 3
	var lastErr error
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Starting upload request to Cloudinary (attempt %d/%d)...", attempt, maxRetries)
		
		// Recreate request body for each retry (it gets consumed)
		var retryBody bytes.Buffer
		retryWriter := multipart.NewWriter(&retryBody)
		
		// Re-add file
		retryPart, partErr := retryWriter.CreateFormFile("file", file.Filename)
		if partErr != nil {
			log.Printf("Error creating retry form file: %v", partErr)
			lastErr = partErr
			break
		}
		if _, partErr := retryPart.Write(imageData); partErr != nil {
			log.Printf("Error writing retry file: %v", partErr)
			lastErr = partErr
			break
		}
		
		// Regenerate timestamp and signature for retry
		retryTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
		retryParams := map[string]string{
			"timestamp": retryTimestamp,
			"folder":    "finance-transactions",
		}
		retrySignature := generateCloudinarySignature(retryParams, apiSecret)
		
		retryWriter.WriteField("api_key", apiKey)
		retryWriter.WriteField("timestamp", retryTimestamp)
		retryWriter.WriteField("signature", retrySignature)
		retryWriter.WriteField("folder", "finance-transactions")
		retryWriter.Close()
		
		// Create new request
		retryReq, reqErr := http.NewRequest("POST", uploadURL, &retryBody)
		if reqErr != nil {
			log.Printf("Error creating retry request: %v", reqErr)
			lastErr = reqErr
			break
		}
		retryReq.Header.Set("Content-Type", retryWriter.FormDataContentType())
		
		resp, lastErr = client.Do(retryReq)
		if lastErr == nil {
			// Success!
			break
		}
		
		log.Printf("Attempt %d failed: %v", attempt, lastErr)
		
		// If not the last attempt, wait before retrying
		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * time.Second
			log.Printf("Waiting %v before retry...", waitTime)
			time.Sleep(waitTime)
		}
	}
	
	if lastErr != nil {
		log.Printf("Error executing request to Cloudinary after %d attempts: %v", maxRetries, lastErr)
		// Check if it's a timeout error
		if strings.Contains(lastErr.Error(), "timeout") || strings.Contains(lastErr.Error(), "i/o timeout") {
			return c.Status(500).JSON(fiber.Map{
				"error": "ไม่สามารถเชื่อมต่อกับ Cloudinary ได้ (timeout) กรุณาตรวจสอบ network connection หรือลองใหม่อีกครั้ง",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("ไม่สามารถเชื่อมต่อกับ Cloudinary ได้: %v", lastErr),
		})
	}
	
	if resp == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่ได้รับ response จาก Cloudinary",
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


