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

	// Clear existing data
	log.Println("Clearing existing data...")
	database.DB.Exec("TRUNCATE TABLE registrations, sub_districts, districts, provinces RESTART IDENTITY CASCADE")
	log.Println("Existing data cleared")

	log.Println("Fetching provinces data...")
	var provincesData []ProvinceData
	if err := fetchJSON("https://raw.githubusercontent.com/kongvut/thai-province-data/refs/heads/master/api/latest/province.json", &provincesData); err != nil {
		log.Fatal("Error fetching provinces:", err)
	}
	log.Printf("Found %d provinces\n", len(provincesData))

	log.Println("Seeding provinces...")
	for i, province := range provincesData {
		p := models.Province{
			ID:     uint(province.ID),
			NameTh: province.NameTh,
			NameEn: province.NameEn,
		}
		if err := database.DB.Create(&p).Error; err != nil {
			log.Printf("Error creating province %s: %v\n", province.NameTh, err)
			continue
		}
		if (i+1)%10 == 0 {
			fmt.Printf("Seeded %d/%d provinces\n", i+1, len(provincesData))
		}
	}
	log.Printf("✓ Seeded %d provinces successfully\n", len(provincesData))

	log.Println("Fetching districts data...")
	var districtsData []DistrictData
	if err := fetchJSON("https://raw.githubusercontent.com/kongvut/thai-province-data/refs/heads/master/api/latest/district.json", &districtsData); err != nil {
		log.Fatal("Error fetching districts:", err)
	}
	log.Printf("Found %d districts\n", len(districtsData))

	log.Println("Seeding districts...")
	for i, district := range districtsData {
		d := models.District{
			ID:         uint(district.ID),
			ProvinceID: uint(district.ProvinceID),
			NameTh:     district.NameTh,
			NameEn:     district.NameEn,
		}
		if err := database.DB.Create(&d).Error; err != nil {
			log.Printf("Error creating district %s: %v\n", district.NameTh, err)
			continue
		}
		if (i+1)%100 == 0 {
			fmt.Printf("Seeded %d/%d districts\n", i+1, len(districtsData))
		}
	}
	log.Printf("✓ Seeded %d districts successfully\n", len(districtsData))

	log.Println("Fetching sub-districts data...")
	var subDistrictsData []SubDistrictDataFlexible
	if err := fetchJSON("https://raw.githubusercontent.com/kongvut/thai-province-data/refs/heads/master/api/latest/sub_district.json", &subDistrictsData); err != nil {
		log.Fatal("Error fetching sub-districts:", err)
	}
	log.Printf("Found %d sub-districts\n", len(subDistrictsData))

	log.Println("Seeding sub-districts...")
	for i, subDistrict := range subDistrictsData {
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
			log.Printf("Error creating sub-district %s: %v\n", subDistrict.NameTh, err)
			continue
		}
		if (i+1)%200 == 0 {
			fmt.Printf("Seeded %d/%d sub-districts\n", i+1, len(subDistrictsData))
		}
	}
	log.Printf("✓ Seeded %d sub-districts successfully\n", len(subDistrictsData))

	log.Println("========================================")
	log.Println("🎉 Database seeded successfully!")
	log.Printf("📊 Total: %d provinces, %d districts, %d sub-districts\n",
		len(provincesData), len(districtsData), len(subDistrictsData))
	log.Println("========================================")

	os.Exit(0)
}
