package handlers

import (
	"registration-system/database"
	"registration-system/models"

	"github.com/gofiber/fiber/v2"
)

func GetProvinces(c *fiber.Ctx) error {
	var provinces []models.Province

	result := database.DB.Order("name_th").Find(&provinces)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch provinces",
		})
	}

	return c.JSON(provinces)
}

func GetDistricts(c *fiber.Ctx) error {
	provinceID := c.Params("province_id")
	var districts []models.District

	result := database.DB.Where("province_id = ?", provinceID).Order("name_th").Find(&districts)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch districts",
		})
	}

	return c.JSON(districts)
}

func GetSubDistricts(c *fiber.Ctx) error {
	districtID := c.Params("district_id")
	var subDistricts []models.SubDistrict

	result := database.DB.Where("district_id = ?", districtID).Order("name_th").Find(&subDistricts)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch sub-districts",
		})
	}

	return c.JSON(subDistricts)
}
