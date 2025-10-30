package main

import (
	"fmt"
	"log"
	"os"
	"registration-system/database"
	"registration-system/models"

	"github.com/joho/godotenv"
)

// types and fetchJSON are shared via common.go

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	database.Connect()

	log.Println("Fetching sub-districts data...")
	var subDistrictsData []SubDistrictDataFlexible
	if err := fetchJSON("https://raw.githubusercontent.com/kongvut/thai-province-data/refs/heads/master/api/latest/sub_district.json", &subDistrictsData); err != nil {
		log.Fatal("Error fetching sub-districts:", err)
	}
	log.Printf("Found %d sub-districts\n", len(subDistrictsData))

	// Clear existing sub-districts
	log.Println("Clearing existing sub-districts...")
	database.DB.Exec("TRUNCATE TABLE sub_districts RESTART IDENTITY CASCADE")
	log.Println("Existing sub-districts cleared")

	log.Println("Seeding sub-districts...")
	successCount := 0
	skipCount := 0
	errorCount := 0

	for i, subDistrict := range subDistrictsData {
		// Skip if district_id is 0 or invalid
		if subDistrict.DistrictID == 0 {
			skipCount++
			continue
		}

		// Convert zip_code to string
		zipCode := ""
		switch v := subDistrict.ZipCode.(type) {
		case string:
			zipCode = v
		case float64:
			zipCode = fmt.Sprintf("%.0f", v)
		case int:
			zipCode = fmt.Sprintf("%d", v)
		}

		sd := models.SubDistrict{
			ID:         uint(subDistrict.ID),
			DistrictID: uint(subDistrict.DistrictID),
			NameTh:     subDistrict.NameTh,
			NameEn:     subDistrict.NameEn,
			ZipCode:    zipCode,
		}

		if err := database.DB.Create(&sd).Error; err != nil {
			errorCount++
			if errorCount <= 5 {
				log.Printf("Error creating sub-district %s (district_id=%d): %v\n", subDistrict.NameTh, subDistrict.DistrictID, err)
			}
			continue
		}

		successCount++
		if (i+1)%500 == 0 {
			fmt.Printf("Processed %d/%d sub-districts (Success: %d, Skipped: %d, Errors: %d)\n",
				i+1, len(subDistrictsData), successCount, skipCount, errorCount)
		}
	}

	log.Println("========================================")
	log.Println("🎉 Sub-districts seeding completed!")
	log.Printf("📊 Total processed: %d\n", len(subDistrictsData))
	log.Printf("✅ Successfully seeded: %d\n", successCount)
	log.Printf("⏭️  Skipped (invalid district_id): %d\n", skipCount)
	log.Printf("❌ Errors: %d\n", errorCount)
	log.Println("========================================")

	os.Exit(0)
}
