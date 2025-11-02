package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Username string `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Password string `gorm:"type:varchar(255);not null" json:"-"` // Never send password in JSON
	FullName string `gorm:"type:varchar(200);not null" json:"full_name"`
	IsActive bool   `gorm:"default:true" json:"is_active"`
	Role     string `gorm:"type:varchar(50);default:'registration'" json:"role"` // "registration" or "finance"
}

type Province struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	NameTh    string     `gorm:"type:varchar(100);not null" json:"name_th"`
	NameEn    string     `gorm:"type:varchar(100);not null" json:"name_en"`
	Districts []District `json:"districts,omitempty"`
}

type District struct {
	ID           uint          `gorm:"primaryKey" json:"id"`
	ProvinceID   uint          `gorm:"not null" json:"province_id"`
	Province     Province      `json:"province,omitempty"`
	NameTh       string        `gorm:"type:varchar(100);not null" json:"name_th"`
	NameEn       string        `gorm:"type:varchar(100);not null" json:"name_en"`
	SubDistricts []SubDistrict `json:"sub_districts,omitempty"`
}

type SubDistrict struct {
	ID         uint     `gorm:"primaryKey" json:"id"`
	DistrictID uint     `gorm:"not null" json:"district_id"`
	District   District `json:"district,omitempty"`
	NameTh     string   `gorm:"type:varchar(100);not null" json:"name_th"`
	NameEn     string   `gorm:"type:varchar(100);not null" json:"name_en"`
	ZipCode    string   `gorm:"type:varchar(5)" json:"zip_code"`
}

type Registration struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	FullName  string    `gorm:"type:varchar(200);not null" json:"full_name"`
	Nickname  string    `gorm:"type:varchar(100)" json:"nickname"`
	BirthDate time.Time `gorm:"not null" json:"birth_date"`

	ProvinceID    uint        `gorm:"not null" json:"province_id"`
	Province      Province    `json:"province,omitempty"`
	DistrictID    uint        `gorm:"not null" json:"district_id"`
	District      District    `json:"district,omitempty"`
	SubDistrictID uint        `gorm:"not null" json:"sub_district_id"`
	SubDistrict   SubDistrict `json:"sub_district,omitempty"`
	AddressDetail string      `gorm:"type:text;not null" json:"address_detail"`

	PhoneNumber      string `gorm:"type:varchar(20);not null" json:"phone_number"`
	TempleName       string `gorm:"type:varchar(200)" json:"temple_name"`
	MedicalCondition string `gorm:"type:text" json:"medical_condition"`
	Vassa            int    `gorm:"default:0" json:"vassa"` // พรรษา

	// Chanting Status - สถานะการสวด
	ChantedPariwat bool `gorm:"default:false" json:"chanted_pariwat"` // สวดปริวาสแล้ว
	ChantedManat   bool `gorm:"default:false" json:"chanted_manat"`   // สวดมานัดแล้ว
	ChantedOkApan  bool `gorm:"default:false" json:"chanted_ok_apan"` // สวดออกอาพานแล้ว
}

// Transaction - รายรับรายจ่าย
type Transaction struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Type        string    `gorm:"type:varchar(20);not null" json:"type"` // "income" or "expense"
	Amount      float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Description string    `gorm:"type:text;not null" json:"description"`
	Date        time.Time `gorm:"not null" json:"date"`
	Category    string    `gorm:"type:varchar(100)" json:"category"` // หมวดหมู่ เช่น "บุญบารมี", "ค่าใช้จ่ายทั่วไป"

	// Relationship
	UserID      uint      `gorm:"not null" json:"user_id"`
	User        User      `json:"user,omitempty"`
}

// ActivityLog - บันทึกการทำกิจกรรม (ต้อง login - รู้ว่าใครทำ)
type ActivityLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`

	Action      string    `gorm:"type:varchar(200);not null" json:"action"` // สิ่งที่ทำ เช่น "เพิ่มรายรับ", "แก้ไขข้อมูล", "ลบรายการ"
	Description string    `gorm:"type:text" json:"description"` // รายละเอียดเพิ่มเติม
	Module      string    `gorm:"type:varchar(50)" json:"module"` // โมดูล เช่น "transaction", "registration"

	// Relationship
	UserID      uint      `gorm:"not null" json:"user_id"`
	User        User      `json:"user,omitempty"`
}

// DeviceLog - บันทึกข้อมูลอุปกรณ์ (ไม่ต้อง login - PDPA compliant)
type DeviceLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`

	// Device Information (PDPA compliant - ไม่เก็บข้อมูลส่วนบุคคลที่ระบุตัวตน)
	DeviceType  string    `gorm:"type:varchar(50)" json:"device_type"` // "mobile", "tablet", "desktop"
	DeviceInfo  string    `gorm:"type:text" json:"device_info"` // รายละเอียดอุปกรณ์ (ไม่มีข้อมูลส่วนบุคคล)
	Action      string    `gorm:"type:varchar(200);not null" json:"action"` // สิ่งที่ทำ เช่น "ลงทะเบียน", "ดูข้อมูล"
	Description string    `gorm:"type:text" json:"description"` // รายละเอียด
	Module      string    `gorm:"type:varchar(50)" json:"module"` // โมดูล เช่น "registration", "public"
	
	// Optional - ไม่เก็บข้อมูลที่ระบุตัวตน
	IPAddress   string    `gorm:"type:varchar(50)" json:"ip_address"` // IP address (อาจลบส่วนสุดท้ายเพื่อความเป็นส่วนตัว)
}
