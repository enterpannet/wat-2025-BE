package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Shared data models for seeding to avoid re-declarations across files in this package

type ProvinceData struct {
	ID     int    `json:"id"`
	NameTh string `json:"name_th"`
	NameEn string `json:"name_en"`
}

type DistrictData struct {
	ID         int    `json:"id"`
	ProvinceID int    `json:"province_id"`
	NameTh     string `json:"name_th"`
	NameEn     string `json:"name_en"`
}

// Latest schema variant (amphure_id, zip_code is string)
type SubDistrictDataLatest struct {
	ID         int    `json:"id"`
	DistrictID int    `json:"amphure_id"`
	NameTh     string `json:"name_th"`
	NameEn     string `json:"name_en"`
	ZipCode    string `json:"zip_code"`
}

// Flexible schema variant (district_id, zip_code can be string/number)
type SubDistrictDataFlexible struct {
	ID         int         `json:"id"`
	DistrictID int         `json:"district_id"`
	NameTh     string      `json:"name_th"`
	NameEn     string      `json:"name_en"`
	ZipCode    interface{} `json:"zip_code"`
}

func fetchJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch data from %s: %v", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return nil
}
