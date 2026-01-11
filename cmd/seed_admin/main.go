package main

import (
	"fmt"
	"log"

	"github.com/habbazettt/mahad-service-go/config"
	"github.com/habbazettt/mahad-service-go/models"
	"github.com/habbazettt/mahad-service-go/utils"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error memuat file .env: %v. Pastikan Anda menjalankan perintah dari folder root 'absensi-service'.", err)
	}

	db := config.ConnectDB()
	fmt.Println("✅ Berhasil terhubung ke database...")

	// Seed Admin saja
	err = seedDefaultAdmin(db)
	if err != nil {
		log.Fatalf("❌ Gagal melakukan seeding data Admin: %v", err)
	}

	fmt.Println("\n✅ Seeding admin berhasil diselesaikan!")
}

// seedDefaultAdmin membuat admin default jika belum ada
func seedDefaultAdmin(db *gorm.DB) error {
	fmt.Println("\n--- Memulai Seeding Data Admin Default ---")

	// Cek apakah admin sudah ada
	var existingAdmin models.Admin
	result := db.Where("email = ?", "admin@mtadigital.com").First(&existingAdmin)

	if result.Error == gorm.ErrRecordNotFound {
		// Hash password
		hashedPassword, err := utils.HashPassword("admin123")
		if err != nil {
			return fmt.Errorf("gagal hash password admin: %w", err)
		}

		// Buat admin default
		admin := models.Admin{
			Nama:     "Administrator",
			Email:    "admin@mtadigital.com",
			Password: hashedPassword,
		}

		if err := db.Create(&admin).Error; err != nil {
			return fmt.Errorf("gagal membuat admin: %w", err)
		}

		fmt.Printf("👍 Berhasil membuat admin default: %s (%s)\n", admin.Nama, admin.Email)
		fmt.Println("   ⚠️  Password default: admin123 - Segera ganti password ini!")
	} else if result.Error != nil {
		return result.Error
	} else {
		fmt.Printf("ℹ️  Admin dengan email %s sudah ada, dilewati.\n", existingAdmin.Email)
	}

	return nil
}
